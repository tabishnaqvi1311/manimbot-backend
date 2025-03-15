package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tabishnaqvi1311/manimbot-backend/handlers"
)

func main(){
	router := gin.Default()

	router.GET("/", func(c *gin.Context){
		c.JSON(200, gin.H{"message": "testingg"})
	})

	router.GET("/generate", handlers.HandleGenerate)

	router.Run("localhost:8000")
}