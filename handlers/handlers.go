package handlers

import (
	"fmt"
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

func generateManim(prompt string, c *gin.Context) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	ctx := c.Request.Context()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})

	if err != nil {
		fmt.Println("error starting cleint", err)
		return
	}

	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(prompt), nil)

	if err != nil {
		fmt.Println("error generating")
		return;
	}

	fmt.Println(result.Candidates[0].Content.Parts[0].Text)
}

func HandleGenerate(c *gin.Context){
	generateManim(SystemPrompt, c)
}