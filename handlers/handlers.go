package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tabishnaqvi1311/manimbot-backend/utils"

	// "github.com/tabishnaqvi1311/manimbot-backend/utils"
	"google.golang.org/genai"
)

const SystemPrompt = `
You are an expert in creating educational animations with Manim. 
Generate Python code using the Manim library to visualize and explain the concept.
Your response should only contain a complete, working Manim Python code that demonstrates the concept visually, nothing extra such as comments or docstrings

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

	if apiKey == "" {
		return "", fmt.Errorf("env not set")
	}

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
	startTime := time.Now()

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
	elapsed := time.Since(startTime)
	fmt.Printf("generated manim code in [%s]\n", elapsed)

	startTime = time.Now()
	code := utils.ExtractCode(content)
	if code == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error extracting code"})
		return
	}
	elapsed = time.Since(startTime)
	fmt.Printf("extracted code in [%s]\n", elapsed)
	
	startTime = time.Now()
	video, err := utils.RunCode(code)
	if err != nil {
		fmt.Println(err)
		dir, err := os.Getwd()
		if err != nil {
			return
		}
		os.RemoveAll(dir + "/static")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error running code"})
		return
	}
	elapsed = time.Since(startTime)
	fmt.Printf("ran code in [%s]\n", elapsed)
	
	startTime = time.Now()
	s3Url, err := utils.UploadToS3(video) 
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return;
	}
	elapsed = time.Since(startTime)
	fmt.Printf("uploaded to s3 in [%s]\n", elapsed)

	fmt.Println(video)

	c.JSON(http.StatusOK, gin.H{"message": s3Url})
}
