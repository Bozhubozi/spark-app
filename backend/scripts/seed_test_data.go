package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Load the test user (the one with a real phone number)
	var testUser model.User
	if err := db.Where("phone IS NOT NULL OR email IS NOT NULL").First(&testUser).Error; err != nil {
		log.Fatal("no test user found — register one first via the API")
	}
	fmt.Printf("Test user: %s (%s)\n", testUser.Nickname, testUser.ID)

	// Load 10 random bots
	var bots []model.User
	if err := db.Where("phone IS NULL AND wechat_open_id IS NULL AND id != ?", testUser.ID).
		Order("random()").Limit(10).Find(&bots).Error; err != nil {
		log.Fatalf("load bots: %v", err)
	}
	fmt.Printf("Loaded %d bots\n", len(bots))

	chatRepo := repository.NewChatRepo(db)
	matchRepo := repository.NewMatchRepo(db)

	matchesCreated := 0
	roomsCreated := 0
	var lastRoomID string

	for _, bot := range bots {
		// Create a match record (mutual match)
		now := time.Now()
		match := &model.Match{
			ID:        uuid.New(),
			UserID1:   testUser.ID,
			UserID2:   bot.ID,
			Score:     0.6 + rng.Float64()*0.4,
			Status:    model.MatchStatusMatched,
			CreatedAt: now.Add(-time.Duration(rng.Intn(72)) * time.Hour),
			MatchedAt: &now,
		}
		if err := matchRepo.Create(context.Background(), match); err != nil {
			log.Printf("create match: %v", err)
			continue
		}
		matchesCreated++

		// Create a chat room for the match
		room, err := chatRepo.FindOrCreateRoom(context.Background(), testUser.ID, bot.ID)
		if err != nil {
			log.Printf("create room: %v", err)
			continue
		}
		roomsCreated++
		lastRoomID = room.ID.String()

		// Generate 3-8 messages from the bot
		numMsgs := 3 + rng.Intn(6)
		greetings := []string{
			"嗨！很高兴认识你 👋",
			"你好呀，看到你的资料觉得很有趣",
			"Hi～你的星座跟我很配哦",
			"哈喽，我们好像有共同的兴趣爱好！",
			"你好，你的性格测试结果好有意思",
			"嗨！看了你的卡片，感觉我们挺合的",
		}
		replies := []string{
			"你也喜欢这个吗？太巧了！",
			"哈哈哈哈确实",
			"周末一般都做什么呀？",
			"你是在哪个城市的呢？",
			"我也是！好难得遇到同好",
			"最近有在追什么剧/番吗？",
			"你的推荐真的好棒",
			"下次可以一起去呀",
		}

		for j := 0; j < numMsgs; j++ {
			var content string
			if j == 0 {
				content = greetings[rng.Intn(len(greetings))]
			} else {
				content = replies[rng.Intn(len(replies))]
			}

			msg := &model.Message{
				ID:          uuid.New(),
				RoomID:      room.ID,
				SenderID:    bot.ID,
				ClientMsgID: uuid.New().String(),
				ContentType: 1,
				Content:     content,
				IsRead:      rng.Intn(2) == 0,
				SentAt:      now.Add(time.Duration(j+1) * time.Minute),
			}
			if err := chatRepo.SaveMessage(context.Background(), msg); err != nil {
				log.Printf("save message: %v", err)
			}
		}
	}

	fmt.Printf("Done! %d matches, %d chat rooms created\n", matchesCreated, roomsCreated)
	if lastRoomID != "" {
		fmt.Printf("Last room ID: %s\n", lastRoomID)
	}
}
