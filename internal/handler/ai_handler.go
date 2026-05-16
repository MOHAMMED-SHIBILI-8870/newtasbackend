// internal/handler/ai_handler.go
package handler

import (
	"backend/internal/entity"
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

type AIHandler struct {
	GeminiClient *genai.Client
}

func NewAIHandler(client *genai.Client) *AIHandler {
	return &AIHandler{GeminiClient: client}
}

const TravelAgentSystemPrompt = `
You are an expert AI Travel Agent. Your job is to help users design a custom travel package when they don't like the default ones.
Follow these rules:
1. Be enthusiastic, polite, and helpful.
2. Ask questions one or two at a time so you don't overwhelm the user.
3. Find out: Destination preferences, budget, who is traveling, trip duration, and favorite activities.
4. If the user provides a custom typed answer, acknowledge it and adapt.
5. Once you have enough information, generate a clear, day-by-day customized itinerary with a summary.
`

func (h *AIHandler) CustomTripChat(c *fiber.Ctx) error {
	var req entity.ChatRequest

	// Parse incoming message history from frontend
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Use gemini-2.5-flash: it's fast, smart, and fully supported on the free tier
	model := "gemini-2.5-flash"

	// Define runtime configurations like your system prompt instructions
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: TravelAgentSystemPrompt}},
		},
	}

	// Request content generation using history context
	resp, err := h.GeminiClient.Models.GenerateContent(
		context.Background(),
		model,
		req.Messages,
		config,
	)

	if err != nil {
		log.Printf("Gemini API Error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate response from Gemini AI",
		})
	}

	// Extract the text output from the response candidates safely
	var aiReply string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		aiReply = resp.Candidates[0].Content.Parts[0].Text
	} else {
		aiReply = "I am having trouble processing that right now. Let's try again!"
	}

	return c.JSON(entity.ChatResponse{Reply: aiReply})
}