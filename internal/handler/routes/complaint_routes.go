package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func ComplaintRoutes(app fiber.Router, h *handler.ComplaintHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	complaints := app.Group("/complaints", auth)
	complaints.Post("/", h.CreateComplaint)
	complaints.Get("/me", h.ListMyComplaints)
	complaints.Get("/:id", h.GetComplaintByID)

	adminComplaints := app.Group("/admin/complaints", auth)
	adminComplaints.Get("/", permission("manage_complaints"), h.ListAllComplaints)
	adminComplaints.Patch("/:id/status", permission("manage_complaints"), h.UpdateComplaintStatus)
	adminComplaints.Delete("/:id", permission("manage_complaints"), h.DeleteComplaint)
}
