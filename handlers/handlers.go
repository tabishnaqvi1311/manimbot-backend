package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	database "github.com/tabishnaqvi1311/manimbot-backend/db"
	"github.com/tabishnaqvi1311/manimbot-backend/models"
	"github.com/tabishnaqvi1311/manimbot-backend/utils"
	"google.golang.org/genai"
)

const SystemPrompt = `
You are an expert in creating educational animations with Manim in the style of 3Blue1Brown. 
Generate Python code using the Manim library to visualize and explain the concept with smooth, elegant animations.

CRITICAL REQUIREMENTS:
- The class MUST be named "Scene" exactly
- Use "from manim import *" for imports
- **ALL COORDINATES MUST BE 3D**: Manim requires 3-dimensional coordinates [x, y, z]
  * For 2D visualizations, set z=0: np.array([x, y, 0])
  * Convert 2D points to 3D: np.append(point_2d, 0) or [x, y, 0]
  * Use Manim vectors: RIGHT*x + UP*y (automatically 3D)
- MINIMUM 60 SECONDS duration - use self.wait() strategically to ensure this
- Style animations like 3Blue1Brown: smooth, thoughtful, mathematically elegant
- Use run_time parameters (typically 1-2 seconds) for smoother animations
- Add rate_func=smooth for fluid motion (e.g., rate_func=rate_functions.smooth)

3Blue1Brown Animation Style:
- Smooth transformations with appropriate run_time (1-3 seconds per animation)
- Use Transform, ReplacementTransform, and FadeTransform for elegant transitions
- Include self.wait(1-3) after important visuals for viewer comprehension
- Use color gradients and visual hierarchy (BLUE, YELLOW, GREEN for emphasis)
- Build complexity gradually - introduce one element at a time
- Use Tex() for mathematical expressions with proper LaTeX formatting
- Animate text and equations appearing with Write() or FadeIn() over 1-2 seconds

COORDINATE HANDLING - CRITICAL:
When working with data points, scatter plots, or clustering:
  # Generate 2D data
  points_2d = np.random.randn(20, 2)
  
  # Convert to 3D for Manim (REQUIRED)
  points_3d = np.array([[x, y, 0] for x, y in points_2d])
  # OR
  points_3d = [np.append(point, 0) for point in points_2d]
  
  # Create dots with 3D coordinates
  dots = VGroup(*[Dot(point, radius=0.1, color=BLUE) for point in points_3d])

For positioning objects:
  obj.move_to(np.array([2, 1, 0]))  # Always 3D
  obj.move_to(RIGHT*2 + UP*1)       # Manim's vector notation (automatically 3D)

Animation Smoothness Tips:
- Always specify run_time for self.play() (minimum 0.5s, typically 1-2s)
- Use self.wait(1-2) between major concepts
- Avoid choppy animations - use Transform instead of removing/adding
- Example: self.play(Transform(obj1, obj2), run_time=2, rate_func=smooth)

Duration Structure (MINIMUM 60 seconds total):
- Introduction with title: 10-15s
- Core explanation with visuals: 25-40s  
- Examples or variations: 20-40s
- Summary or key insight: 10-15s
- Add self.wait(2-3) at the end

Example structure:
from manim import *
import numpy as np

class Scene(Scene):
    def construct(self):
        # Title (10s)
        title = Text("Concept Name", font_size=48)
        self.play(Write(title), run_time=2)
        self.wait(2)
        self.play(FadeOut(title), run_time=1)
        
        # For data visualization (convert 2D to 3D!)
        data_2d = np.random.randn(10, 2)
        data_3d = np.array([[x, y, 0] for x, y in data_2d])
        dots = VGroup(*[Dot(point, radius=0.1, color=BLUE) for point in data_3d])
        self.play(Create(dots), run_time=2)
        self.wait(2)
        
        # More animations...
        self.wait(2)

The code must:
- Convert ALL 2D coordinates to 3D format [x, y, 0]
- Be MINIMUM 60 seconds (use self.wait() to ensure this)
- Use run_time parameters on ALL self.play() calls
- Create smooth, professional animations like 3Blue1Brown
- Work with Manim Community Edition
- Be self-contained and runnable
`

const ExplanationPrompt = `
Provide a clear, comprehensive explanation of this concept in 3-5 paragraphs. 
Focus on the key principles, practical understanding, and real-world applications.
Make it accessible but informative, suitable for learners at various levels.
`

type GenerateRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	ChatID string `json:"chat_id"`
}

type ChatResponse struct {
	ChatID      string    `json:"chat_id"`
	MessageID   string    `json:"message_id"`
	VideoURL    string    `json:"video_url"`
	Explanation string    `json:"explanation"`
	Duration    int       `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChatHistoryResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatDetailResponse struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Messages  []MessageResponse `json:"messages"`
	CreatedAt time.Time         `json:"created_at"`
}

type MessageResponse struct {
	ID          string    `json:"id"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	VideoURL    string    `json:"video_url,omitempty"`
	Explanation string    `json:"explanation,omitempty"`
	Duration    int       `json:"duration,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func assessComplexity(prompt string) string {
	prompt = strings.ToLower(prompt)
	complexKeywords := []string{"quantum", "relativity", "calculus", "theorem", "theory", "proof", "derive", "integrate", "differential"}
	moderateKeywords := []string{"explain", "how does", "why", "concept", "principle"}

	for _, keyword := range complexKeywords {
		if strings.Contains(prompt, keyword) {
			return "complex"
		}
	}

	for _, keyword := range moderateKeywords {
		if strings.Contains(prompt, keyword) {
			return "moderate"
		}
	}

	return "simple"
}

func generateTitle(prompt string) string {
	words := strings.Fields(prompt)

	skipWords := map[string]bool{
		"can": true, "you": true, "help": true, "me": true, "explain": true,
		"show": true, "demonstrate": true, "visualize": true, "create": true,
		"make": true, "generate": true, "the": true, "a": true, "an": true,
	}

	var titleWords []string
	for _, word := range words {
		if !skipWords[strings.ToLower(word)] {
			titleWords = append(titleWords, word)
		}
		if len(titleWords) >= 5 {
			break
		}
	}

	if len(titleWords) == 0 {
		return truncateTitle(prompt)
	}

	title := strings.Join(titleWords, " ")
	if len(title) > 50 {
		return title[:47] + "..."
	}
	return title
}

func generateManim(ctx context.Context, prompt string, complexity string, previousError string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("env not set")
	}

	durationGuide := ""
	switch complexity {
	case "simple":
		durationGuide = "Target: 60-120 seconds. Create smooth animations with proper run_time. Show one clear example with elegant transformations."
	case "moderate":
		durationGuide = "Target: 120-240 seconds. Use multiple examples with smooth transitions. Include intermediate steps with self.wait() for comprehension."
	case "complex":
		durationGuide = "Target: 240-600 seconds. Break into clear segments. Show step-by-step derivations with patient pacing. Multiple examples with detailed explanations."
	}

	fullPrompt := SystemPrompt + "\n" + durationGuide + "\n\nUser request: " + prompt

	if previousError != "" {
		fullPrompt += fmt.Sprintf("\n\nIMPORTANT: Previous attempt failed with error:\n%s\n\nPlease fix this error. Common issues:\n- 2D coordinates not converted to 3D (use np.array([x, y, 0]) or np.append(point, 0))\n- Missing imports (numpy as np)\n- Incorrect Dot() positioning (must use 3D coordinates)\n\nEnsure ALL coordinates are 3D format.", previousError)
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", err
	}

	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(fullPrompt), nil)
	if err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func generateExplanation(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("env not set")
	}

	fullPrompt := ExplanationPrompt + "\n\nTopic: " + prompt

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", err
	}

	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(fullPrompt), nil)
	if err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func isCoordinateError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "broadcast together with shapes") ||
		strings.Contains(errStr, "(32,3)") ||
		strings.Contains(errStr, "(2,)") ||
		strings.Contains(errStr, "operands could not be broadcast")
}

func HandleGenerate(c *gin.Context) {
	var req GenerateRequest
	startTime := time.Now()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	clerkUserID := c.GetHeader("X-User-ID")
	if clerkUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("clerk_id = ?", clerkUserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var chat models.Chat
	if req.ChatID != "" {
		if err := database.DB.Where("id = ? AND user_id = ?", req.ChatID, user.ID).First(&chat).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
			return
		}
	} else {
		chat = models.Chat{
			ID:     uuid.New().String(),
			UserID: user.ID,
			Title:  generateTitle(req.Prompt),
		}
		if err := database.DB.Create(&chat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create chat"})
			return
		}
	}

	userMessage := models.Message{
		ID:      uuid.New().String(),
		ChatID:  chat.ID,
		Role:    "user",
		Content: req.Prompt,
	}
	if err := database.DB.Create(&userMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
		return
	}

	complexity := assessComplexity(req.Prompt)

	maxRetries := 2
	var lastError string
	var video string
	var actualDuration int
	var s3Url string
	var explanation string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("Retry attempt %d/%d due to coordinate error\n", attempt, maxRetries)
		}

		content, err := generateManim(c.Request.Context(), req.Prompt, complexity, lastError)
		if err != nil {
			fmt.Println("error generating manim code:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate animation code"})
			return
		}
		fmt.Printf("generated manim code in [%s]\n", time.Since(startTime))

		if attempt == 0 {
			explanation, err = generateExplanation(c.Request.Context(), req.Prompt)
			if err != nil {
				fmt.Println("error generating explanation:", err)
				explanation = "Explanation unavailable"
			}
		}

		startTime = time.Now()
		code := utils.ExtractCode(content)
		if code == "" {
			fmt.Println("error: could not extract code from response")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract animation code"})
			return
		}
		fmt.Printf("extracted code in [%s]\n", time.Since(startTime))

		startTime = time.Now()
		video, actualDuration, err = utils.RunCode(code)
		if err != nil {
			fmt.Println("error running code:", err)
			dir, _ := os.Getwd()
			os.RemoveAll(dir + "/static")

			if isCoordinateError(err) && attempt < maxRetries {
				lastError = err.Error()
				fmt.Printf("Detected coordinate error, retrying with error context...\n")
				continue
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("animation generation failed: %v", err)})
			return
		}

		if actualDuration < 60 {
			fmt.Printf("Warning: Video duration (%ds) is below minimum.\n", actualDuration)
			dir, _ := os.Getwd()
			os.RemoveAll(dir + "/static")

			if attempt < maxRetries {
				lastError = fmt.Sprintf("Video duration was only %d seconds, need at least 60 seconds", actualDuration)
				continue
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("animation too short (%ds). Animations must be at least 60 seconds", actualDuration)})
			return
		}

		fmt.Printf("ran code and measured duration (%ds) in [%s]\n", actualDuration, time.Since(startTime))

		startTime = time.Now()
		s3Url, err = utils.UploadToS3(video)
		if err != nil {
			fmt.Println("error uploading to s3:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video"})
			return
		}
		fmt.Printf("uploaded to s3 in [%s]\n", time.Since(startTime))

		break
	}

	assistantMessage := models.Message{
		ID:          uuid.New().String(),
		ChatID:      chat.ID,
		Role:        "assistant",
		Content:     req.Prompt,
		VideoURL:    s3Url,
		Explanation: explanation,
		Duration:    actualDuration,
	}
	if err := database.DB.Create(&assistantMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save response"})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		ChatID:      chat.ID,
		MessageID:   assistantMessage.ID,
		VideoURL:    s3Url,
		Explanation: explanation,
		Duration:    actualDuration,
		CreatedAt:   assistantMessage.CreatedAt,
	})
}

func GetChatHistory(c *gin.Context) {
	clerkUserID := c.GetHeader("X-User-ID")
	if clerkUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("clerk_id = ?", clerkUserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var chats []models.Chat
	if err := database.DB.Where("user_id = ?", user.ID).Order("updated_at DESC").Find(&chats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch chats"})
		return
	}

	response := make([]ChatHistoryResponse, len(chats))
	for i, chat := range chats {
		response[i] = ChatHistoryResponse{
			ID:        chat.ID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

func GetChatDetail(c *gin.Context) {
	chatID := c.Param("id")
	clerkUserID := c.GetHeader("X-User-ID")

	if clerkUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("clerk_id = ?", clerkUserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var chat models.Chat
	if err := database.DB.Preload("Messages").Where("id = ? AND user_id = ?", chatID, user.ID).First(&chat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	}

	messages := make([]MessageResponse, len(chat.Messages))
	for i, msg := range chat.Messages {
		messages[i] = MessageResponse{
			ID:          msg.ID,
			Role:        msg.Role,
			Content:     msg.Content,
			VideoURL:    msg.VideoURL,
			Explanation: msg.Explanation,
			Duration:    msg.Duration,
			CreatedAt:   msg.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, ChatDetailResponse{
		ID:        chat.ID,
		Title:     chat.Title,
		Messages:  messages,
		CreatedAt: chat.CreatedAt,
	})
}

func DeleteChat(c *gin.Context) {
	chatID := c.Param("id")
	clerkUserID := c.GetHeader("X-User-ID")

	if clerkUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("clerk_id = ?", clerkUserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", chatID, user.ID).Delete(&models.Chat{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete chat"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "chat deleted successfully"})
}

func CreateOrGetUser(c *gin.Context) {
	var req struct {
		ClerkID  string `json:"clerk_id" binding:"required"`
		Email    string `json:"email" binding:"required"`
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var user models.User
	result := database.DB.Where("clerk_id = ?", req.ClerkID).First(&user)

	if result.Error != nil {
		user = models.User{
			ID:       uuid.New().String(),
			ClerkID:  req.ClerkID,
			Email:    req.Email,
			FullName: req.FullName,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	c.JSON(http.StatusOK, user)
}

func truncateTitle(prompt string) string {
	if len(prompt) > 50 {
		return prompt[:47] + "..."
	}
	return prompt
}
