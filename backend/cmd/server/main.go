package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/handler"
	"github.com/spark-app/backend/internal/middleware"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"github.com/spark-app/backend/internal/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&model.UserReport{}); err != nil {
		log.Printf("[WARN] migrate UserReport: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("[WARN] redis unavailable: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	chatRepo := repository.NewChatRepo(db)
	interestRepo := repository.NewInterestRepo(db)

	// Services
	authSvc := service.NewAuthService(cfg, userRepo)
	matchSvc := service.NewMatchService(matchRepo, interestRepo, userRepo, rdb)
	chatSvc := service.NewChatService(chatRepo, userRepo)
	wsHub := service.NewWSHub(rdb)

	personalityReportSvc := service.NewPersonalityReportService(rdb)
	zodiacSvc := service.NewZodiacService()
	horoscopeSvc := service.NewHoroscopeService()
	icebreakerSvc := service.NewIcebreakerService(zodiacSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userRepo, interestRepo, personalityReportSvc, zodiacSvc, horoscopeSvc)
	matchH := handler.NewMatchHandler(matchSvc, wsHub, userRepo, zodiacSvc, icebreakerSvc, interestRepo, chatRepo)
	chatH := handler.NewChatHandler(chatSvc, chatRepo, wsHub)
	dfaFilter := util.BuildDFAFilter()
	wsH := handler.NewWSHandler(wsHub, chatSvc, authSvc, chatRepo, dfaFilter)

	// Rate limiters
	authLimiter := middleware.NewRateLimiter(5, 10)    // 5 req/s, burst 10
	authLimiter.StartCleanup(5 * time.Minute)
	generalLimiter := middleware.NewRateLimiter(20, 50) // 20 req/s, burst 50
	generalLimiter.StartCleanup(5 * time.Minute)

	r := gin.Default()

	// CORS for web dev
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Public auth (rate-limited per IP+path)
	auth := r.Group("/api/v1/auth")
	auth.Use(authLimiter.PerPath())
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/wechat-login", authH.WechatLogin)
	}

	// WebSocket (public but validates token via query param)
	r.GET("/ws", wsH.HandleWebSocket)

	// Protected (rate-limited per IP)
	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired(authSvc))
	api.Use(generalLimiter.PerIP())
	{
		api.GET("/user/profile", userH.GetProfile)
		api.PUT("/user/profile", userH.UpdateProfile)
		api.GET("/user/tags", userH.GetTags)
		api.PUT("/user/interests", userH.SaveInterests)
		api.GET("/user/personality/questions", userH.GetPersonalityQuestions)
		api.POST("/user/personality", userH.SubmitPersonality)
		api.GET("/user/personality", userH.GetPersonality)
		api.GET("/user/personality/report", userH.GetPersonalityReport)
		api.GET("/user/horoscope", userH.GetHoroscope)
		api.GET("/user/avatars", userH.GetAvatarComponents)
		api.POST("/user/device-token", userH.SaveDeviceToken)
		api.POST("/user/account/cancel", userH.DeleteAccount)
		api.POST("/user/report", userH.ReportUser)
			api.POST("/user/account/restore", userH.RestoreAccount)

		api.GET("/match/likes-count", matchH.LikesCount)
			api.GET("/match/likers", matchH.GetLikers)
			api.GET("/match/remaining", matchH.RemainingSwipes)
			api.GET("/match/blocked", matchH.GetBlocked)
			api.POST("/match/unblock", matchH.Unblock)
			api.GET("/match/candidates", matchH.GetCandidates)
		api.POST("/match/swipe", matchH.Swipe)
		api.GET("/match/list", matchH.GetMatches)
		api.GET("/match/zodiac-compat/:target_user_id", matchH.ZodiacCompat)

		api.POST("/chat/rooms", chatH.GetOrCreateRoom)
			api.GET("/chat/rooms", chatH.GetRooms)
		api.GET("/chat/rooms/:room_id/messages", chatH.GetMessages)
		api.POST("/chat/rooms/:room_id/read", chatH.MarkRead)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Spark backend starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
