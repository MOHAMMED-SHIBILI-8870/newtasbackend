package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

type AIHandler struct {
	client *genai.Client
}

func NewAIHandler(client *genai.Client) *AIHandler {
	return &AIHandler{
		client: client,
	}
}

// ================= REQUEST STRUCT =================

type TripRequest struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Days        int    `json:"days"`
	TripType    string `json:"trip_type"`
	BudgetLevel string `json:"budget_level"`
	Members     int    `json:"members"`

	// Optional Admin Features
	Children   int    `json:"children"`
	HotelType  string `json:"hotel_type"`
	Transport  string `json:"transport"`
	CreatedBy  string `json:"created_by"` // user / admin
}

// ================= GENERATE TRIP PLAN =================

func (h *AIHandler) GenerateTripPlan(c *fiber.Ctx) error {

	var req TripRequest

	// ================= PARSE BODY =================

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// ================= VALIDATION =================

	if req.From == "" || req.To == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "From and To locations are required",
		})
	}

	if req.Days <= 0 {
		req.Days = 1
	}

	if req.Members <= 0 {
		req.Members = 1
	}

	if req.CreatedBy == "" {
		req.CreatedBy = "user"
	}

	// ================= AI PROMPT =================

	prompt := fmt.Sprintf(`
Create a concise %s trip plan.

Trip Details:
- From: %s
- To: %s
- Duration: %d days
- Travelers: %d people
- Children: %d
- Budget Level: %s
- Hotel Preference: %s
- Transport Preference: %s
- Created By: %s

Include:
1. Day-wise itinerary
2. Tourist attractions
3. Food recommendations
4. Transport suggestions
5. Hotel suggestions
6. Estimated total budget
7. Useful travel tips

Keep the response clean, short, and easy to read.
`,
		req.TripType,
		req.From,
		req.To,
		req.Days,
		req.Members,
		req.Children,
		req.BudgetLevel,
		req.HotelType,
		req.Transport,
		req.CreatedBy,
	)

	// ================= TIMEOUT =================

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	// ================= GEMINI REQUEST =================

	resp, err := h.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)

	// ================= ERROR HANDLING =================

	if err != nil {

		fmt.Println("GEMINI ERROR:", err)

		// Rate limit handling
		if strings.Contains(err.Error(), "429") ||
			strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {

			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please wait and try again.",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to generate AI response",
			"details": err.Error(),
		})
	}

	// ================= EMPTY RESPONSE =================

	if resp == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Empty AI response",
		})
	}

	// ================= SUCCESS RESPONSE =================

	return c.JSON(fiber.Map{
		"success": true,
		"created_by": req.CreatedBy,
		"result":  resp.Text(),
	})
}