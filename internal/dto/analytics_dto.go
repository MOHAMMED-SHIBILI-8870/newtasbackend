package dto

type AnalyticsDashboardResponse struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalBookings     int64   `json:"total_bookings"`
	TotalUsers        int64   `json:"total_users"`
	ActiveTrips       int64   `json:"active_trips"`
	AvailableVehicles int64   `json:"available_vehicles"`
	PendingComplaints int64   `json:"pending_complaints"`
	OpenSupportReqs   int64   `json:"open_support_requests"`
}
