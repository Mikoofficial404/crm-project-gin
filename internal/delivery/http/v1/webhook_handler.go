package v1

import (
	"context"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	leadService *service.LeadService
	waService   *service.WhatsAppService
	userRepo    *postgres.UserRepository
	leadRepo    *postgres.LeadRepository
	activitySvc *service.ActivityService
	notifSvc    *service.NotificationService
	redisClien  *redis.Client
	aiService   *service.AIService
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
) *WebhookHandler {
	return &WebhookHandler{
		leadService: leadService,
		waService:   waService,
		userRepo:    userRepo,
		leadRepo:    leadRepo,
		activitySvc: activitySvc,
		notifSvc:    notifSvc,
		redisClien:  redisClients,
		aiService:   aiService,
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

	fmt.Printf("CHAT WA MASUK DARI %s (%s): %s\n", payload.Payload.FromName, payload.Payload.From, payload.Payload.Body)

	phone := strings.Split(payload.Payload.From, "@")[0]

	existingLead, _ := h.leadRepo.GetLeadByPhone(phone)
	if existingLead == nil {
		fmt.Println("[Webhook Klien baru terdeteksi! Membuat Lead otomatis...")
		salesAll, err := h.userRepo.GetAllUsers()
		if err != nil {
			fmt.Println("Error", err)
		}
		turnNo, err := h.redisClien.Get(context.Background(), "sales_turn_index").Int()
		if err != nil {
			turnNo = 0
		}
		assigneedID := salesAll[turnNo].ID
		var newNo int
		newNo = turnNo + 1
		if newNo >= len(salesAll) {
			newNo = 0
		}
		h.redisClien.Set(context.Background(), "sales_turn_index", newNo, 0)
		newLead, errCreate := h.leadService.CreateLead(payload.Payload.FromName, "", phone, assigneedID, nil)
		if errCreate != nil {
			fmt.Println("Gagal membuat Lead otomatis:", errCreate)
		} else {
			h.activitySvc.CreateActivity("WhatsApp", payload.Payload.Body, newLead.ID, assigneedID, "")
			go func() {
				aiReply, errAi := h.aiService.GenerateSalesReply(payload.Payload.FromName, payload.Payload.Body)
				if errAi != nil {
					fmt.Println("[AI Error]:", errAi)
					h.waService.SendWA(phone, "Terima kasih telah menghubungi kami. Tim kami akan segera membalas pesan Anda.")
					h.activitySvc.CreateActivity("Catatan", "Bot Reply: Terima kasih telah menghubungi kami...", newLead.ID, assigneedID, "")
				} else {
					h.waService.SendWA(phone, aiReply)
					h.activitySvc.CreateActivity("Catatan", "AI Reply: "+aiReply, newLead.ID, assigneedID, "")
				}
			}()
		}
	} else {
		fmt.Printf("[Webhook] 👤 Klien Lama (%s) mengirim pesan.\n", existingLead.Name)
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
			fmt.Println("Gagal mencatat Aktivitas:", errAct)
		} else {
			notifTitle := fmt.Sprintf("Pesan Wa: %s", existingLead.Name)
			errNotif := h.notifSvc.CreateNotification(assigneeID, notifTitle, payload.Payload.Body)
			if errNotif != nil {
				fmt.Println("Gagal membuat Notifikasi:", errNotif)
			}

			wsMessage := fmt.Sprintf(`{"type":"new_whatsapp","title":"%s","message":"%s"}`, notifTitle, payload.Payload.Body)
			errWs := websocket.SendMessageToUser(assigneeID, wsMessage)
			if errWs != nil {
				fmt.Printf("⚠️ User %s sedang offline, WebSocket dilewati.\n", assigneeID)
			} else {
				fmt.Println("Sinyal WebSocket berhasil ditembakkan ke layar!")
			}

		}
	}

	c.JSON(200, gin.H{"status": "ok"})
}
