package service

import (
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"encoding/json"
	"errors"
	"fmt"
)

type DealCommentService struct {
	repo         *postgres.DealCommentRepository
	notifService *NotificationService
}

func NewDealCommentService(repo *postgres.DealCommentRepository, notifService *NotificationService) *DealCommentService {
	return &DealCommentService{
		repo:         repo,
		notifService: notifService,
	}
}

func (s *DealCommentService) AddComment(dealID, userID, content string, mentionedUsers []string) error {
	if dealID == "" || userID == "" || content == "" {
		return errors.New("dealID, userID, dan content wajib diisi")
	}

	mentionedJSON := "[]"
	if len(mentionedUsers) > 0 {
		b, _ := json.Marshal(mentionedUsers)
		mentionedJSON = string(b)
	}

	_, err := s.repo.CreateComment(dealID, userID, content, mentionedJSON)
	if err != nil {
		return err
	}

	for _, mentionedUserID := range mentionedUsers {
		if mentionedUserID == "" {
			continue
		}
		s.notifService.CreateNotification(
			mentionedUserID,
			"Kamu di-mention di sebuah komentar",
			content,
			fmt.Sprintf("/deals/%s", dealID),
		)
		message := fmt.Sprintf("Kamu di-mention di komentar deal oleh %s: %s", userID, content)
		websocket.SendMessageToUser(mentionedUserID, message)
	}

	return nil
}

func (s *DealCommentService) GetComments(dealID string, page, limit int) ([]entity.DealComment, int64, error) {
	if dealID == "" {
		return nil, 0, errors.New("dealID wajib diisi")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetCommentsByDealID(dealID, page, limit)
}

func (s *DealCommentService) DeleteComment(commentID string) error {
	if commentID == "" {
		return errors.New("commentID wajib diisi")
	}
	return s.repo.DeleteComment(commentID)
}
