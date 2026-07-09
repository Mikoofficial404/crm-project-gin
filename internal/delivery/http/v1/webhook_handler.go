package v1

import (
	"context"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type GowaWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Body     string `json:"body"`
		From     string `json:"from"`
		FromName string `json:"from_name"`
		IsFromMe bool   `json:"is_from_me"`
	} `json:"payload"`
}

type WebhookHandler struct {
	leadService    *service.LeadService
	waService      *service.WhatsAppService
	userRepo       *postgres.UserRepository
	leadRepo       *postgres.LeadRepository
	activitySvc    *service.ActivityService
	notifSvc       *service.NotificationService
	redisClien     *redis.Client
	aiService      *service.AIService
	contactService *service.ContactService
}

func NewWebhookHandler(
	leadService *service.LeadService,
	waService *service.WhatsAppService,
	userRepo *postgres.UserRepository,
	leadRepo *postgres.LeadRepository,
	activitySvc *service.ActivityService,
	notifSvc *service.NotificationService,
	redisClients *redis.Client,
	aiService *service.AIService,
	contactService *service.ContactService,
) *WebhookHandler {
	return &WebhookHandler{
		leadService:    leadService,
		waService:      waService,
		userRepo:       userRepo,
		leadRepo:       leadRepo,
		activitySvc:    activitySvc,
		notifSvc:       notifSvc,
		redisClien:     redisClients,
		aiService:      aiService,
		contactService: contactService,
	}
}

func (h *WebhookHandler) ReceiveWhatsApp(c *gin.Context) {
	var payload GowaWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(200, gin.H{"status": "ok"})
		return
	}

	if payload.Event != "message" || payload.Payload.IsFromMe {
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"from": payload.Payload.From,
		"name": payload.Payload.FromName,
	}).Info("Chat WA masuk")

	phone := strings.Split(payload.Payload.From, "@")[0]

	existingLead, _ := h.leadRepo.GetLeadByPhone(phone)
	if existingLead == nil {
		logrus.Info("[Webhook] Klien baru terdeteksi! Membuat Lead otomatis...")

		salesAll, err := h.userRepo.GetAllUsers()
		if err != nil {
			logrus.WithError(err).Error("[Webhook] Gagal get users")
			c.JSON(200, gin.H{"status": "ok"})
			return
		}

		var salesOnly []entity.User
		for _, u := range salesAll {
			if u.Role == "sales" && u.IsOnline {
				salesOnly = append(salesOnly, u)
			}
		}

		if len(salesOnly) == 0 {
			logrus.Error("[Webhook] Tidak ada sales tersedia untuk di-assign")
			c.JSON(200, gin.H{"status": "ok"})
			return
		}

		turnNo, err := h.redisClien.Get(context.Background(), "sales_turn_index").Int()
		if err != nil {
			turnNo = 0
		}

		if turnNo >= len(salesOnly) {
			turnNo = 0
		}

		assignedID := salesOnly[turnNo].ID
		newNo := turnNo + 1
		if newNo >= len(salesOnly) {
			newNo = 0
		}
		h.redisClien.Set(context.Background(), "sales_turn_index", newNo, 0)

		contact, errContact := h.contactService.FindOrCreateContact(payload.Payload.FromName, phone, assignedID)
		if errContact != nil {
			logrus.WithError(errContact).Warn("[Webhook] Gagal FindOrCreateContact")
		}

		var contactID *string
		if contact != nil {
			contactID = &contact.ID
		}

		newLead, errCreate := h.leadService.CreateLeadWithContact(payload.Payload.FromName, "", phone, assignedID, nil, contactID)
		if errCreate != nil {
			logrus.WithError(errCreate).Error("[Webhook] Gagal membuat Lead otomatis")
		} else {
			h.activitySvc.CreateActivity("WhatsApp", payload.Payload.Body, newLead.ID, assignedID, "")
			go func() {
				aiReply, errAi := h.aiService.GenerateSalesReply(payload.Payload.FromName, payload.Payload.Body)
				if errAi != nil {
					logrus.WithError(errAi).Warn("[Webhook] AI Error, fallback ke bot reply")
					h.waService.SendWA(phone, "Terima kasih telah menghubungi kami. Tim kami akan segera membalas pesan Anda.")
					h.activitySvc.CreateActivity("Catatan", "Bot Reply: Terima kasih telah menghubungi kami...", newLead.ID, assignedID, "")
				} else {
					h.waService.SendWA(phone, aiReply)
					h.activitySvc.CreateActivity("Catatan", "AI Reply: "+aiReply, newLead.ID, assignedID, "")
				}
			}()
		}
	} else {
		logrus.WithField("name", existingLead.Name).Info("[Webhook] Klien lama mengirim pesan")
		assigneeID := existingLead.AssignedTo
		if assigneeID == "" || assigneeID == "00000000-0000-0000-0000-000000000000" {
			firstUser, _ := h.userRepo.GetFirstUser()
			if firstUser != nil {
				assigneeID = firstUser.ID
			} else {
				assigneeID = "00000000-0000-0000-0000-000000000000"
			}
		}

		_, errAct := h.activitySvc.CreateActivity("WhatsApp", payload.Payload.Body, existingLead.ID, assigneeID, "")
		if errAct != nil {
			logrus.WithError(errAct).Error("[Webhook] Gagal mencatat Aktivitas")
		} else {
			notifTitle := fmt.Sprintf("Pesan Wa: %s", existingLead.Name)
			errNotif := h.notifSvc.CreateNotification(assigneeID, notifTitle, payload.Payload.Body, "")
			if errNotif != nil {
				logrus.WithError(errNotif).Warn("[Webhook] Gagal membuat Notifikasi")
			}

			wsMessage := fmt.Sprintf(`{"type":"new_whatsapp","title":"%s","message":"%s"}`, notifTitle, payload.Payload.Body)
			errWs := websocket.SendMessageToUser(assigneeID, wsMessage)
			if errWs != nil {
				logrus.WithField("user_id", assigneeID).Info("[Webhook] User sedang offline, WebSocket dilewati")
			} else {
				logrus.Info("[Webhook] Sinyal WebSocket berhasil dikirim")
			}
		}
	}

	c.JSON(200, gin.H{"status": "ok"})
}
