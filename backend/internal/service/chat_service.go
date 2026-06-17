package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
)

type ChatService struct {
	chatRepo *repository.ChatRepo
	userRepo *repository.UserRepo
}

func NewChatService(cr *repository.ChatRepo, ur *repository.UserRepo) *ChatService {
	return &ChatService{chatRepo: cr, userRepo: ur}
}

func (s *ChatService) GetOrCreateRoom(ctx context.Context, user1, user2 uuid.UUID) (*model.ChatRoom, error) {
	return s.chatRepo.FindOrCreateRoom(ctx, user1, user2)
}

func (s *ChatService) SaveMessage(ctx context.Context, msg *model.Message) error {
	if msg.ClientMsgID == "" {
		msg.ClientMsgID = uuid.New().String()
	}
	return s.chatRepo.SaveMessage(ctx, msg)
}

func (s *ChatService) GetMessages(ctx context.Context, roomID uuid.UUID, before time.Time, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 50 {
		limit = 30
	}
	if before.IsZero() {
		before = time.Now().Add(time.Hour)
	}
	msgs, err := s.chatRepo.FindMessages(ctx, roomID, before, limit)
	if err != nil {
		return nil, fmt.Errorf("find messages: %w", err)
	}
	// Reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *ChatService) GetRooms(ctx context.Context, userID uuid.UUID) ([]model.ChatRoom, error) {
	return s.chatRepo.FindRooms(ctx, userID)
}

func (s *ChatService) GetRoomsWithDetails(ctx context.Context, userID uuid.UUID) ([]repository.RoomWithLastMsg, error) {
	return s.chatRepo.FindRoomsWithDetails(ctx, userID)
}

func (s *ChatService) MarkRead(ctx context.Context, roomID, userID uuid.UUID) error {
	return s.chatRepo.MarkRead(ctx, roomID, userID)
}
