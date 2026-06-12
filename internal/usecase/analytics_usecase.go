package usecase

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type AnalyticsUsecase struct {
	db *gorm.DB
}

func NewAnalyticsUsecase(db *gorm.DB) *AnalyticsUsecase {
	return &AnalyticsUsecase{db: db}
}

func (u *AnalyticsUsecase) GetDashboardMetrics(ctx context.Context, role string) (*dto.AnalyticsDashboardResponse, error) {
	var metrics dto.AnalyticsDashboardResponse

	// Total Revenue (Sum of successful payments amount)
	var revenue float64
	u.db.WithContext(ctx).Model(&entity.Payment{}).Where("status IN ('success', 'paid')").Select("COALESCE(SUM(amount), 0)").Scan(&revenue)
	metrics.TotalRevenue = revenue

	// Total Bookings
	u.db.WithContext(ctx).Model(&entity.Booking{}).Where("status != 'cancelled'").Count(&metrics.TotalBookings)

	// Total Users
	u.db.WithContext(ctx).Model(&entity.User{}).Count(&metrics.TotalUsers)

	// Active Trips
	u.db.WithContext(ctx).Model(&entity.Trip{}).Where("status = 'active'").Count(&metrics.ActiveTrips)

	// Available Vehicles
	u.db.WithContext(ctx).Model(&entity.Vehicle{}).Where("status = 'available'").Count(&metrics.AvailableVehicles)

	// Pending Complaints
	u.db.WithContext(ctx).Model(&entity.Complaint{}).Where("status = 'pending'").Count(&metrics.PendingComplaints)

	// Open Support Requests
	u.db.WithContext(ctx).Model(&entity.SupportRequest{}).Where("status = 'open'").Count(&metrics.OpenSupportReqs)

	return &metrics, nil
}
