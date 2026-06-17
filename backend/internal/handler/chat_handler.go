package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
)

type ChatHandler struct {
	svc      *service.ChatService
	chatRepo *repository.ChatRepo
	wsHub    *service.WSHub
}

func NewChatHandler(svc *service.ChatService, cr *repository.ChatRepo, wsHub *service.WSHub) *ChatHandler {
	return &ChatHandler{svc: svc, chatRepo: cr, wsHub: wsHub}
}

func (h *ChatHandler) GetRooms(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	rooms, err := h.svc.GetRoomsWithDetails(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomID, _ := uuid.Parse(c.Param("room_id"))
	beforeStr := c.DefaultQuery("before", "")
	var before time.Time
	if beforeStr != "" {
		before, _ = time.Parse(time.RFC3339, beforeStr)
	}
	msgs, err := h.svc.GetMessages(c.Request.Context(), roomID, before, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

type getOrCreateRoomReq struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

func (h *ChatHandler) GetOrCreateRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)

	var req getOrCreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id required"})
		return
	}
	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	room, err := h.svc.GetOrCreateRoom(c.Request.Context(), uid, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	roomID, _ := uuid.Parse(c.Param("room_id"))
	if err := h.svc.MarkRead(c.Request.Context(), roomID, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast read status to the partner
	room, err := h.chatRepo.FindByID(c.Request.Context(), roomID)
	if err == nil && room != nil {
		receiverID := room.UserID1
		if receiverID == uid {
			receiverID = room.UserID2
		}
		readData, _ := json.Marshal(map[string]string{
			"room_id":   roomID.String(),
			"reader_id": uid.String(),
		})
		h.wsHub.SendToUser(receiverID, &service.WSMessage{
			Type: "chat.message.read",
			Data: readData,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
