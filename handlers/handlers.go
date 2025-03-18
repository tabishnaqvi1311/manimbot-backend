package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

const SystemPrompt = `
You are an expert in creating educational animations with Manim. 
Generate Python code using the Manim library to visualize and explain the concept.
Your response should contain:
1. A short explanation of the concept
2. A complete, working Manim Python code that demonstrates the concept visually

The code must:
- Import necessary Manim modules
- Define a Scene class
- Not be overly complex (keep rendering time short, under 10 seconds if possible)
- Work with Manim Community Edition
- Be self-contained and runnable
- Create engaging, colorful, and instructive animations
`

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

func generateManim(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	fullPrompt := SystemPrompt + "\n" + prompt

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})

	if err != nil {
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(fullPrompt),
		nil,
	)

	if err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func HandleGenerate(c *gin.Context) {
	var req GenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "prompt cannot be empty"})
		return
	}

	content, err := generateManim(c.Request.Context(), req.Prompt)
	if err != nil {
		fmt.Println("error generating manim code ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": content})
}
