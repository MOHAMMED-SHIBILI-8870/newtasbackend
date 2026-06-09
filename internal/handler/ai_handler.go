package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

type AIHandler struct {
	client     *genai.Client
	usecase    *usecase.AITripRequestUsecase
	ragUsecase *usecase.RAGUsecase
}

func NewAIHandler(client *genai.Client, aiUsecase *usecase.AITripRequestUsecase, ragUsecase *usecase.RAGUsecase) *AIHandler {

	return &AIHandler{
		client:     client,
		usecase:    aiUsecase,
		ragUsecase: ragUsecase,
	}
}

type TripRequest struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Days        int    `json:"days"`
	TripType    string `json:"trip_type"`
	BudgetLevel string `json:"budget_level"`
	Members     int    `json:"members"`
	Children    int    `json:"children"`
	HotelType   string `json:"hotel_type"`
	Transport   string `json:"transport"`
	CreatedBy   string `json:"created_by"`
}

func (h *AIHandler) GenerateTripPlan(c *fiber.Ctx) error {
	var req TripRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if req.From == "" || req.To == "" {
		return response.Fail(c, fiber.StatusBadRequest, "from and to locations are required", nil)
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

	ragContext := ""

	if h.ragUsecase != nil {

		contextData, err := h.ragUsecase.BuildContext(
			c.Context(),
			req.To,
		)

		if err == nil {
			ragContext = contextData
		}
	}

	prompt := fmt.Sprintf(`
You are a professional travel planner.

Reference Trips From Database:

%s

User Request:

From: %s
To: %s
Duration: %d days
Trip Type: %s
Budget Level: %s
Members: %d
Children: %d
Hotel Type: %s
Transport: %s

Instructions:

1. Use the reference trips as guidance.
2. Do not copy exactly.
3. Create a fresh itinerary.
4. Include day-wise plans.
5. Include attractions.
6. Include hotel suggestions.
7. Include food suggestions.
8. Include estimated budget.

`,
		ragContext,
		req.From,
		req.To,
		req.Days,
		req.TripType,
		req.BudgetLevel,
		req.Members,
		req.Children,
		req.HotelType,
		req.Transport,
	)

	result, err := h.generatePlan(c.Context(), prompt, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "rate limit exceeded") {
			status = fiber.StatusTooManyRequests
		}
		return response.Fail(c, status, "failed to generate AI response", err)
	}

	return response.Success(c, fiber.StatusOK, "trip plan generated successfully", fiber.Map{
		"created_by": req.CreatedBy,
		"prompt":     prompt,
		"result":     result,
	})
}

func (h *AIHandler) CreateTripRequest(c *fiber.Ctx) error {
	if h.usecase == nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, "ai request service unavailable", nil)
	}

	var input dto.AITripRequestInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	request, err := h.usecase.CreateRequest(c.Context(), userID, entity.AITripRequestInput{
		From:          input.From,
		To:            input.To,
		Days:          input.Days,
		TripType:      input.TripType,
		BudgetLevel:   input.BudgetLevel,
		Members:       input.Members,
		Children:      input.Children,
		HotelType:     input.HotelType,
		Transport:     input.Transport,
		Prompt:        input.Prompt,
		GeneratedPlan: input.GeneratedPlan,
	})
	if err != nil {
		return response.Fail(c, aiTripRequestStatusFromErr(err), "failed to submit ai trip request", err)
	}

	return response.Success(c, fiber.StatusCreated, "ai trip request submitted successfully", toAITripRequestResponse(request))
}

func (h *AIHandler) GetMyTripRequests(c *fiber.Ctx) error {
	if h.usecase == nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, "ai request service unavailable", nil)
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	requests, err := h.usecase.GetRequests(c.Context(), middleware.GetAuthRole(c), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load ai trip requests", err)
	}

	return response.Success(c, fiber.StatusOK, "ai trip requests loaded successfully", toAITripRequestResponses(requests))
}

func (h *AIHandler) GetAllTripRequests(c *fiber.Ctx) error {
	if h.usecase == nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, "ai request service unavailable", nil)
	}

	requests, err := h.usecase.GetRequests(c.Context(), middleware.GetAuthRole(c), middleware.GetAuthUserID(c))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load ai trip requests", err)
	}

	return response.Success(c, fiber.StatusOK, "ai trip requests loaded successfully", toAITripRequestResponses(requests))
}

func (h *AIHandler) ApproveTripRequest(c *fiber.Ctx) error {
	return h.reviewTripRequest(c, true)
}

func (h *AIHandler) RejectTripRequest(c *fiber.Ctx) error {
	return h.reviewTripRequest(c, false)
}

func (h *AIHandler) reviewTripRequest(c *fiber.Ctx, approve bool) error {
	if h.usecase == nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, "ai request service unavailable", nil)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid ai request id", err)
	}

	var input dto.AITripReviewRequest
	if err := c.BodyParser(&input); err != nil && len(c.Body()) > 0 {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	adminID := middleware.GetAuthUserID(c)
	if adminID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	request, err := h.usecase.ReviewRequest(c.Context(), adminID, uint(id), approve, input.AdminNote)
	if err != nil {
		return response.Fail(c, aiTripRequestStatusFromErr(err), "failed to review ai trip request", err)
	}

	message := "ai trip request rejected successfully"
	if approve {
		message = "ai trip request approved successfully"
	}

	return response.Success(c, fiber.StatusOK, message, toAITripRequestResponse(request))
}

func (h *AIHandler) generatePlan(ctx context.Context, prompt string, req TripRequest) (string, error) {
	if h.client == nil {
		return generateFallbackTripPlan(req), nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := h.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			return "", fmt.Errorf("rate limit exceeded")
		}
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf("empty AI response")
	}

	return resp.Text(), nil
}

func generateFallbackTripPlan(req TripRequest) string {
	days := req.Days
	if days <= 0 {
		days = 1
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# %s to %s\n\n", req.From, req.To))
	builder.WriteString("## Snapshot\n")
	builder.WriteString(fmt.Sprintf("- Duration: %d days\n", days))
	builder.WriteString(fmt.Sprintf("- Trip type: %s\n", req.TripType))
	builder.WriteString(fmt.Sprintf("- Budget: %s\n", req.BudgetLevel))
	builder.WriteString(fmt.Sprintf("- Guests: %d adults, %d children\n\n", req.Members, req.Children))
	builder.WriteString("## Suggested Itinerary\n")
	for day := 1; day <= days; day++ {
		builder.WriteString(fmt.Sprintf("### Day %d\n", day))
		builder.WriteString(fmt.Sprintf("- Morning: Explore a signature landmark in %s\n", req.To))
		builder.WriteString("- Afternoon: Local food and neighborhood walk\n")
		builder.WriteString("- Evening: Relax with a scenic activity or cultural show\n\n")
	}
	builder.WriteString("## Travel Notes\n")
	builder.WriteString(fmt.Sprintf("- Use %s for primary movement.\n", req.Transport))
	builder.WriteString(fmt.Sprintf("- Stay in a %s property.\n", req.HotelType))
	builder.WriteString("- Keep one flexible block each day for weather or rest.\n")
	return builder.String()
}

func aiTripRequestStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "already been reviewed"):
		return fiber.StatusConflict
	case strings.Contains(msg, "unauthorized"):
		return fiber.StatusUnauthorized
	default:
		return fiber.StatusBadRequest
	}
}

func toAITripRequestResponses(requests []entity.AITripRequest) []dto.AITripRequestResponse {
	responses := make([]dto.AITripRequestResponse, 0, len(requests))
	for _, request := range requests {
		responses = append(responses, toAITripRequestResponse(&request))
	}
	return responses
}

func toAITripRequestResponse(request *entity.AITripRequest) dto.AITripRequestResponse {
	if request == nil {
		return dto.AITripRequestResponse{}
	}

	response := dto.AITripRequestResponse{
		ID:            request.ID,
		From:          request.From,
		To:            request.To,
		Days:          request.Days,
		TripType:      request.TripType,
		BudgetLevel:   request.BudgetLevel,
		Members:       request.Members,
		Children:      request.Children,
		HotelType:     request.HotelType,
		Transport:     request.Transport,
		Prompt:        request.Prompt,
		GeneratedPlan: request.GeneratedPlan,
		Status:        request.Status,
		AdminNote:     request.AdminNote,
		TripID:        request.TripID,
		ReviewedByID:  request.ReviewedByID,
		ReviewedAt:    request.ReviewedAt,
		CreatedAt:     request.CreatedAt,
		UpdatedAt:     request.UpdatedAt,
	}

	if request.UserID != 0 {
		user := request.User
		if user.ID == 0 {
			user.ID = request.UserID
		}
		response.User = dto.AITripRequestUser{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		}
	}

	return response
}
