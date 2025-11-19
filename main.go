package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	database "github.com/tabishnaqvi1311/manimbot-backend/db"
	"github.com/tabishnaqvi1311/manimbot-backend/handlers"
)

func main() {
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://feynman-tech.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	api := router.Group("/api")
	{
		api.POST("/users", handlers.CreateOrGetUser)
		api.POST("/generate", handlers.HandleGenerate)
		api.GET("/chats", handlers.GetChatHistory)
		api.GET("/chats/:id", handlers.GetChatDetail)
		api.DELETE("/chats/:id", handlers.DeleteChat)
	}

	log.Println("Server starting on :8000")
	router.Run("0.0.0.0:8000")
}
