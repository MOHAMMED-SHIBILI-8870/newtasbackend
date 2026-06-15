package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DriverHandler struct {
	usecase *usecase.DriverUsecase
}

func NewDriverHandler(u *usecase.DriverUsecase) *DriverHandler {
	return &DriverHandler{usecase: u}
}

// 📄 Admin: Create Driver
func (h *DriverHandler) CreateDriver(c *fiber.Ctx) error {
	var input dto.DriverCreateRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	driver, err := h.usecase.CreateDriver(c.Context(), input)
	if err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusCreated, "driver profile created successfully", mapDriverToResponse(driver))
}

// 📄 Admin: List Drivers
func (h *DriverHandler) ListDrivers(c *fiber.Ctx) error {
	drivers, err := h.usecase.ListDrivers(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to retrieve drivers", err)
	}

	resps := make([]dto.DriverResponse, 0, len(drivers))
	for _, d := range drivers {
		resps = append(resps, mapDriverToResponse(&d))
	}

	return response.Success(c, fiber.StatusOK, "drivers retrieved successfully", resps)
}

// 📄 Admin: Get Driver By ID
func (h *DriverHandler) GetDriverByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid driver ID", err)
	}

	driver, err := h.usecase.GetDriverByID(c.Context(), uint(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "driver not found", err)
	}

	return response.Success(c, fiber.StatusOK, "driver retrieved successfully", mapDriverToResponse(driver))
}

// 📄 Admin: Update Driver
func (h *DriverHandler) UpdateDriver(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid driver ID", err)
	}

	var input dto.DriverUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	driver, err := h.usecase.UpdateDriver(c.Context(), uint(id), input)
	if err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "driver updated successfully", mapDriverToResponse(driver))
}

// 📄 Admin: Delete Driver
func (h *DriverHandler) DeleteDriver(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid driver ID", err)
	}

	if err := h.usecase.DeleteDriver(c.Context(), uint(id)); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to delete driver", err)
	}

	return response.Success(c, fiber.StatusOK, "driver deleted successfully", nil)
}

// 📄 Admin: Assign Vehicle to Driver
func (h *DriverHandler) AssignVehicle(c *fiber.Ctx) error {
	driverID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid driver ID", err)
	}

	var req struct {
		VehicleID uint `json:"vehicle_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.AssignVehicleToDriver(c.Context(), uint(driverID), req.VehicleID); err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "vehicle assigned to driver successfully", nil)
}

// 📄 Admin: Assign Driver to Booking
func (h *DriverHandler) AssignDriverToBooking(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking ID", err)
	}

	var req struct {
		DriverID uint `json:"driver_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.AssignDriverToBooking(c.Context(), uint(bookingID), req.DriverID); err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "driver assigned to booking successfully", nil)
}

// =====================================
// DRIVER PORTAL FUNCTIONS
// =====================================

// 📄 Driver: Get Dashboard
func (h *DriverHandler) GetDriverDashboard(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	stats, err := h.usecase.GetDriverDashboard(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to get dashboard details", err)
	}

	return response.Success(c, fiber.StatusOK, "dashboard loaded successfully", stats)
}

// 📄 Driver: Get Assigned Trips
func (h *DriverHandler) GetDriverTrips(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	bookings, err := h.usecase.GetDriverTrips(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load assigned trips", err)
	}

	resps := make([]entity.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resps = append(resps, mapBookingToResponse(&b))
	}

	return response.Success(c, fiber.StatusOK, "trips loaded successfully", resps)
}

// 📄 Driver: Get Assigned Trip Details
func (h *DriverHandler) GetDriverTripByID(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking ID", err)
	}

	booking, err := h.usecase.GetDriverTripByID(c.Context(), userID, uint(bookingID))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "trip details loaded successfully", mapBookingToResponse(booking))
}

// 📄 Driver: Update Trip Status
func (h *DriverHandler) UpdateTripStatus(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking ID", err)
	}

	var req dto.TripStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.UpdateTripStatus(c.Context(), userID, uint(bookingID), req.Status); err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "trip status updated successfully", nil)
}

// 📄 Driver: Get Vehicle Details
func (h *DriverHandler) GetDriverVehicle(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	vehicle, err := h.usecase.GetDriverVehicle(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, err.Error(), err)
	}

	resp := dto.VehicleResponse{
		ID:             vehicle.ID,
		AgencyID:       vehicle.AgencyID,
		DriverID:       vehicle.DriverID,
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

	return response.Success(c, fiber.StatusOK, "vehicle details loaded successfully", resp)
}

// 📄 Driver: Get Profile Details
func (h *DriverHandler) GetDriverProfile(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	driver, err := h.usecase.GetDriverProfile(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "profile details loaded successfully", mapDriverToResponse(driver))
}

// 📄 Driver: Update Profile Details
func (h *DriverHandler) UpdateDriverProfile(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	var req dto.DriverProfileUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.UpdateDriverProfile(c.Context(), userID, req); err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "profile updated successfully", nil)
}

// 📄 Driver: Update Live Tracking Coordinates
func (h *DriverHandler) UpdateDriverTracking(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	var req dto.DriverTrackingUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.UpdateDriverTracking(c.Context(), userID, req.Latitude, req.Longitude); err != nil {
		return response.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), err)
	}

	return response.Success(c, fiber.StatusOK, "location tracked successfully", nil)
}

// =====================================
// HELPERS
// =====================================

func mapDriverToResponse(driver *entity.Driver) dto.DriverResponse {
	resp := dto.DriverResponse{
		ID:            driver.ID,
		UserID:        driver.UserID,
		Name:          driver.Name,
		Email:         driver.Email,
		Phone:         driver.Phone,
		Address:       driver.Address,
		LicenseNumber: driver.LicenseNumber,
		LicenseExpiry: driver.LicenseExpiry,
		Status:        driver.Status,
		VehicleID:     driver.VehicleID,
		CreatedAt:     driver.CreatedAt,
		UpdatedAt:     driver.UpdatedAt,
	}
	if driver.Vehicle != nil {
		resp.Vehicle = &dto.VehicleResponse{
			ID:             driver.Vehicle.ID,
			AgencyID:       driver.Vehicle.AgencyID,
			DriverID:       driver.Vehicle.DriverID,
			Name:           driver.Vehicle.Name,
			Type:           driver.Vehicle.Type,
			TotalSeats:     driver.Vehicle.TotalSeats,
			AvailableSeats: driver.Vehicle.AvailableSeats,
			PricePerPerson: driver.Vehicle.PricePerPerson,
			Status:         driver.Vehicle.Status,
			TripID:         driver.Vehicle.TripID,
			CreatedAt:      driver.Vehicle.CreatedAt,
			UpdatedAt:      driver.Vehicle.UpdatedAt,
		}
	}
	return resp
}

func mapBookingToResponse(b *entity.Booking) entity.BookingResponse {
	var userResp *entity.UserResponse
	if b.User.ID != 0 {
		userResp = &entity.UserResponse{
			ID:    b.User.ID,
			Name:  b.User.FullName,
			Email: b.User.Email,
		}
	}

	resp := entity.BookingResponse{
		ID:              b.ID,
		Status:          b.Status,
		UserID:          b.UserID,
		User:            userResp,
		TripID:          b.TripID,
		Trip:            &b.Trip,
		SlotID:          b.SlotID,
		VehicleID:       b.VehicleID,
		BookingType:     b.BookingType,
		SeatsBooked:     b.SeatsBooked,
		CouponCode:      b.CouponCode,
		DiscountPercent: b.DiscountPercent,
		BaseAmount:      b.BaseAmount,
		FinalAmount:     b.FinalAmount,
		BalanceAmount:   b.BalanceAmount,
		PaymentStatus:   b.PaymentStatus,
		CreatedAt:       b.CreatedAt,
		StartDate:       b.StartDate,
		EndDate:         b.EndDate,
	}

	if b.DriverID != nil {
		resp.DriverID = b.DriverID
	}
	if b.Driver != nil {
		resp.Driver = b.Driver
	}

	return resp
}
