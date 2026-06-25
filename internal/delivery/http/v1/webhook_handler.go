package v1

import (
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
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
}

func NewWebhookHandler(
	leadService *service.LeadService,
	waService *service.WhatsAppService,
	userRepo *postgres.UserRepository,
	leadRepo *postgres.LeadRepository,
	activitySvc *service.ActivityService,
	notifSvc *service.NotificationService,
) *WebhookHandler {
	return &WebhookHandler{
		leadService: leadService,
		waService:   waService,
		userRepo:    userRepo,
		leadRepo:    leadRepo,
		activitySvc: activitySvc,
		notifSvc:    notifSvc,
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

		admins, _ := h.userRepo.GetAdmins()
		assigneeID := ""
		if len(admins) > 0 {
			assigneeID = admins[0].ID
		} else {
			firstUser, _ := h.userRepo.GetFirstUser()
			if firstUser != nil {
				assigneeID = firstUser.ID
			} else {
				assigneeID = "00000000-0000-0000-0000-000000000000"
			}
		}

		_, err := h.leadService.CreateLead(payload.Payload.FromName, "", phone, assigneeID, nil)
		if err != nil {
			fmt.Println("Gagal membuat Lead otomatis:", err)
		} else {
			pesan := fmt.Sprintf("Mas %s! PENGEN MACBOOK AH elah", payload.Payload.FromName)
			go h.waService.SendWA(phone, pesan)
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

		_, errAct := h.activitySvc.CreateActivity("WhatsApp", payload.Payload.Body, existingLead.ID, assigneeID)
		if errAct != nil {
			fmt.Println("Gagal mencatat Aktivitas:", errAct)
		} else {
			notifTitle := fmt.Sprintf("WA Baru: %s", existingLead.Name)
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
