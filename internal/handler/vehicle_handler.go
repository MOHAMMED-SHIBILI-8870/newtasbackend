package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type VehicleHandler struct {
	usecase *usecase.VehicleUsecase
}

func NewVehicleHandler(usecase *usecase.VehicleUsecase) *VehicleHandler {
	return &VehicleHandler{usecase: usecase}
}

func vehicleStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "invalid"):
		return fiber.StatusBadRequest
	case strings.Contains(msg, "insufficient"):
		return fiber.StatusConflict
	default:
		return fiber.StatusBadRequest
	}
}

func mapVehicleResponse(vehicle entity.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID:             vehicle.ID,
		AgencyID:       vehicle.AgencyID,
		Name:           vehicle.Name,
		Type:           vehicle.Type,
		TotalSeats:     vehicle.TotalSeats,
		AvailableSeats: vehicle.AvailableSeats,
		PricePerPerson: vehicle.PricePerPerson,
		Status:         vehicle.Status,
		TripID:         vehicle.TripID,
		CreatedAt:      vehicle.CreatedAt,
		UpdatedAt:      vehicle.UpdatedAt,
	}
}

func (h *VehicleHandler) ListVehicles(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)

	vehicles, err := h.usecase.ListVehicles(c.Context(), userID, role)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load vehicles", err)
	}

	items := make([]dto.VehicleResponse, 0, len(vehicles))
	for _, vehicle := range vehicles {
		items = append(items, mapVehicleResponse(vehicle))
	}

	return response.Success(c, fiber.StatusOK, "vehicles loaded successfully", items)
}

func (h *VehicleHandler) GetVehicleByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid vehicle id", err)
	}

	vehicle, err := h.usecase.GetVehicleByID(c.Context(), uint(id))
	if err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to load vehicle", err)
	}
	if vehicle == nil {
		return response.Fail(c, fiber.StatusNotFound, "vehicle not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "vehicle loaded successfully", mapVehicleResponse(*vehicle))
}

func (h *VehicleHandler) GetVehicleByTripID(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("trip_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	vehicle, err := h.usecase.GetVehicleByTripID(c.Context(), uint(tripID))
	if err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to load vehicle", err)
	}
	if vehicle == nil {
		return response.Success(c, fiber.StatusOK, "vehicle not assigned to trip", nil)
	}

	return response.Success(c, fiber.StatusOK, "vehicle loaded successfully", mapVehicleResponse(*vehicle))
}

func (h *VehicleHandler) CreateVehicle(c *fiber.Ctx) error {
	var input dto.VehicleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	vehicle := &entity.Vehicle{
		AgencyID:       input.AgencyID,
		Name:           input.Name,
		Type:           input.Type,
		TotalSeats:     input.TotalSeats,
		AvailableSeats: input.AvailableSeats,
		PricePerPerson: input.PricePerPerson,
		Status:         input.Status,
		TripID:         input.TripID,
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	if err := h.usecase.CreateVehicle(c.Context(), userID, role, vehicle); err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to create vehicle", err)
	}

	return response.Success(c, fiber.StatusCreated, "vehicle created successfully", mapVehicleResponse(*vehicle))
}

func (h *VehicleHandler) UpdateVehicle(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid vehicle id", err)
	}

	var input dto.VehicleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	vehicle := &entity.Vehicle{
		AgencyID:       input.AgencyID,
		Name:           input.Name,
		Type:           input.Type,
		TotalSeats:     input.TotalSeats,
		AvailableSeats: input.AvailableSeats,
		PricePerPerson: input.PricePerPerson,
		Status:         input.Status,
		TripID:         input.TripID,
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	if err := h.usecase.UpdateVehicle(c.Context(), userID, role, uint(id), vehicle); err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to update vehicle", err)
	}

	return response.Success(c, fiber.StatusOK, "vehicle updated successfully", nil)
}

func (h *VehicleHandler) DeleteVehicle(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid vehicle id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	if err := h.usecase.DeleteVehicle(c.Context(), userID, role, uint(id)); err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to delete vehicle", err)
	}

	return response.Success(c, fiber.StatusOK, "vehicle deleted successfully", nil)
}

func (h *VehicleHandler) AssignVehicleToTrip(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid vehicle id", err)
	}

	var input dto.AssignVehicleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	if err := h.usecase.AssignVehicleToTrip(c.Context(), userID, role, uint(id), input.TripID); err != nil {
		return response.Fail(c, vehicleStatusFromErr(err), "failed to assign vehicle", err)
	}

	return response.Success(c, fiber.StatusOK, "vehicle assigned successfully", nil)
}
