package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type VerificationHandler struct {
	verificationUsecase usecase.VerificationUsecase
}

func NewVerificationHandler(verificationUsecase usecase.VerificationUsecase) *VerificationHandler {
	return &VerificationHandler{
		verificationUsecase: verificationUsecase,
	}
}

func (h *VerificationHandler) SubmitVerification(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID")
	var userID uint
	if userIDStr != nil {
		userID = uint(userIDStr.(float64))
	} else {
		// Try string conversion if needed, but normally middleware sets it as float64 from JWT
		id, ok := c.Locals("userID").(uint)
		if ok {
			userID = id
		}
	}

	bookingIDStr := c.FormValue("booking_id")
	bookingID, _ := strconv.ParseUint(bookingIDStr, 10, 32)

	fullName := c.FormValue("full_name")
	address := c.FormValue("address")
	phoneNumber := c.FormValue("phone_number")
	membersStr := c.FormValue("members")
	members, _ := strconv.Atoi(membersStr)
	if members == 0 {
		members = 1
	}

	// Handle file upload
	file, err := c.FormFile("id_image")
	var idImageURL string
	if err == nil && file != nil {
		// ensure uploads directory exists
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, 0755)
		}

		filename := fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userID, filepath.Base(file.Filename))
		filePath := filepath.Join(uploadDir, filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "Failed to save ID image",err)
		}
		idImageURL = "/uploads/" + filename
	}

	verification := &entity.Verification{
		UserID:      userID,
		BookingID:   uint(bookingID),
		FullName:    fullName,
		Address:     address,
		PhoneNumber: phoneNumber,
		IDImageURL:  idImageURL,
		Members:     members,
		Status:      "pending",
	}

	if err := h.verificationUsecase.SubmitVerification(verification); err != nil {
		fmt.Println("===== SUBMIT VERIFICATION HIT =====")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to submit verification",err)
	}

	return response.Success(c, fiber.StatusCreated, "Verification submitted successfully", verification)
}


func (h *VerificationHandler) GetVerification(c *fiber.Ctx) error {
    bookingID, err := strconv.ParseUint(c.Params("bookingId"), 10, 32)
    if err != nil {
        return response.Fail(c, fiber.StatusBadRequest, "Invalid booking ID", err)
    }

    verification, err := h.verificationUsecase.GetVerificationByBookingID(uint(bookingID))
    if err != nil {
        return response.Fail(c, fiber.StatusNotFound, "Verification not found", err)
    }

    return response.Success(c, fiber.StatusOK, "Verification found", verification)
}
