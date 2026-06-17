package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupChatDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available, skipping integration test: %v", err)
	}
	return db
}

func createTempUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	user := model.User{ID: uid, Nickname: "test_" + uid.String()[:8], PasswordHash: string(hash)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create temp user: %v", err)
	}
	t.Cleanup(func() { db.Delete(&user) })
	return uid
}

func TestFindOrCreateRoom(t *testing.T) {
	db := setupChatDB(t)
	repo := NewChatRepo(db)
	ctx := context.Background()

	uid1 := createTempUser(t, db)
	uid2 := createTempUser(t, db)

	// Create room
	room, err := repo.FindOrCreateRoom(ctx, uid1, uid2)
	if err != nil {
		t.Fatalf("FindOrCreateRoom: %v", err)
	}
	if room.ID == uuid.Nil {
		t.Error("room ID should not be nil")
	}
	if room.UserID1 != uid1 && room.UserID1 != uid2 {
		t.Error("room should contain uid1")
	}

	// Find same room (should not create duplicate)
	room2, err := repo.FindOrCreateRoom(ctx, uid1, uid2)
	if err != nil {
		t.Fatalf("second FindOrCreateRoom: %v", err)
	}
	if room2.ID != room.ID {
		t.Error("should return the same room on second call")
	}

	// Reverse order should also work
	room3, err := repo.FindOrCreateRoom(ctx, uid2, uid1)
	if err != nil {
		t.Fatalf("reversed FindOrCreateRoom: %v", err)
	}
	if room3.ID != room.ID {
		t.Error("reverse order should return the same room")
	}

	// Cleanup
	db.Delete(&room)
}

func TestSaveAndFindMessages(t *testing.T) {
	db := setupChatDB(t)
	repo := NewChatRepo(db)
	ctx := context.Background()

	uid1 := createTempUser(t, db)
	uid2 := createTempUser(t, db)

	room, err := repo.FindOrCreateRoom(ctx, uid1, uid2)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	// Save messages
	now := time.Now()
	for i := 0; i < 5; i++ {
		msg := &model.Message{
			ID:          uuid.New(),
			RoomID:      room.ID,
			SenderID:    uid1,
			ClientMsgID: uuid.New().String(),
			ContentType: 1,
			Content:     "test message",
			SentAt:      now.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("SaveMessage %d: %v", i, err)
		}
	}

	// Find messages
	msgs, err := repo.FindMessages(ctx, room.ID, time.Now().Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("FindMessages: %v", err)
	}
	if len(msgs) != 5 {
		t.Errorf("expected 5 messages, got %d", len(msgs))
	}

	// Verify chronological order (oldest first after service reversal)
	// Repo returns DESC, service reverses to ASC

	// Mark read
	if err := repo.MarkRead(ctx, room.ID, uid2); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	// Cleanup
	for _, m := range msgs {
		db.Delete(&m)
	}
	db.Delete(&room)
}

func TestFindRooms(t *testing.T) {
	db := setupChatDB(t)
	repo := NewChatRepo(db)
	ctx := context.Background()

	uid1 := createTempUser(t, db)
	uid2 := createTempUser(t, db)
	uid3 := createTempUser(t, db)

	room12, _ := repo.FindOrCreateRoom(ctx, uid1, uid2)
	room13, _ := repo.FindOrCreateRoom(ctx, uid1, uid3)

	rooms, err := repo.FindRooms(ctx, uid1)
	if err != nil {
		t.Fatalf("FindRooms: %v", err)
	}
	if len(rooms) < 2 {
		t.Errorf("expected at least 2 rooms, got %d", len(rooms))
	}

	// Cleanup
	db.Delete(&room12)
	db.Delete(&room13)
}
