package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_foreign_keys=on",
		strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()),
	)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	models := []any{
		&entity.User{},
		&entity.RefreshToken{},
		&entity.OTP{},
		&entity.Role{},
		&entity.UserRole{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.Trip{},
		&entity.TripPlan{},
		&entity.Offer{},
		&entity.Vehicle{},
		&entity.Booking{},
		&entity.BookingPlan{},
		&entity.Review{},
		&entity.Notification{},
		&entity.AITripRequest{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func mustCreateRole(t *testing.T, db *gorm.DB, name string) entity.Role {
	t.Helper()

	role := entity.Role{
		Name:        name,
		Description: name + " role",
	}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role %s: %v", name, err)
	}

	return role
}

func mustCreateUser(t *testing.T, db *gorm.DB, fullName, email, role string, verified bool) entity.User {
	t.Helper()

	user := entity.User{
		FullName:     fullName,
		Email:        email,
		HashPassword: "hashed-password",
		Role:         role,
		IsVerified:   verified,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}

	return user
}

func mustCreateOffer(t *testing.T, db *gorm.DB, code string, discount float64) entity.Offer {
	t.Helper()

	offer := entity.Offer{
		Code:            code,
		Title:           code,
		DiscountPercent: discount,
		ExpiryDate:      time.Now().Add(24 * time.Hour),
		Active:          true,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatalf("create offer %s: %v", code, err)
	}

	return offer
}

func mustCreateTrip(t *testing.T, db *gorm.DB, price float64, endDate time.Time) entity.Trip {
	t.Helper()

	trip := entity.Trip{
		From:        "Delhi",
		To:          "Goa",
		StartDate:   time.Now().Add(-24 * time.Hour),
		EndDate:     endDate,
		Duration:    3,
		TripType:    "Family",
		BudgetLevel: "Medium",
		Price:       price,
		Members:     2,
		Children:    0,
		HotelType:   "3 Star",
		Transport:   "Car",
		Status:      "active",
	}
	if err := db.Create(&trip).Error; err != nil {
		t.Fatalf("create trip: %v", err)
	}

	return trip
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func TestCreateOTPAndVerifyOTP(t *testing.T) {
	db := newTestDB(t)
	user := mustCreateUser(t, db, "Alice", "alice@example.com", "user", false)

	otp, err := CreateOTP(db, context.Background(), user.Email, "signup", 5)
	if err != nil {
		t.Fatalf("create otp: %v", err)
	}
	if otp == "" {
		t.Fatal("expected otp value")
	}

	ok, err := VerifyOTP(db, context.Background(), user.Email, otp, "signup")
	if err != nil {
		t.Fatalf("verify otp: %v", err)
	}
	if !ok {
		t.Fatal("expected otp verification to succeed")
	}

	var updated entity.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if !updated.IsVerified {
		t.Fatal("expected signup OTP to mark user verified")
	}

	var otpCount int64
	if err := db.Model(&entity.OTP{}).
		Where("email = ? AND purpose = ?", user.Email, "signup").
		Count(&otpCount).Error; err != nil {
		t.Fatalf("count otp rows: %v", err)
	}
	if otpCount != 0 {
		t.Fatalf("expected otp rows to be consumed, got %d", otpCount)
	}
}

func TestRefreshTokenRotationReplacesOldToken(t *testing.T) {
	db := newTestDB(t)
	user := mustCreateUser(t, db, "Bob", "bob@example.com", "user", true)

	plain, hashed, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	expiresAt := time.Now().Add(time.Hour)
	if err := SaveRefreshToken(db, context.Background(), user.ID, hashed, expiresAt); err != nil {
		t.Fatalf("save refresh token: %v", err)
	}

	rotatedUser, newPlain, _, err := RotateRefreshToken(db, context.Background(), plain)
	if err != nil {
		t.Fatalf("rotate refresh token: %v", err)
	}
	if rotatedUser.ID != user.ID {
		t.Fatalf("expected rotated token to belong to user %d, got %d", user.ID, rotatedUser.ID)
	}
	if newPlain == "" || newPlain == plain {
		t.Fatal("expected rotation to issue a new refresh token")
	}

	if _, err := ValidateRefreshToken(db, context.Background(), plain); err == nil {
		t.Fatal("expected original refresh token to be invalid after rotation")
	}

	var tokenCount int64
	if err := db.Model(&entity.RefreshToken{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error; err != nil {
		t.Fatalf("count refresh tokens: %v", err)
	}
	if tokenCount != 1 {
		t.Fatalf("expected one active refresh token, got %d", tokenCount)
	}
}

func TestRotateRefreshTokenRejectsBlockedUser(t *testing.T) {
	db := newTestDB(t)
	user := mustCreateUser(t, db, "Cara", "cara@example.com", "user", true)

	plain, hashed, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	if err := SaveRefreshToken(db, context.Background(), user.ID, hashed, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("save refresh token: %v", err)
	}

	if err := db.Model(&entity.User{}).Where("id = ?", user.ID).Update("is_blocked", true).Error; err != nil {
		t.Fatalf("block user: %v", err)
	}

	if _, _, _, err := RotateRefreshToken(db, context.Background(), plain); err == nil || !strings.Contains(strings.ToLower(err.Error()), "blocked") {
		t.Fatalf("expected blocked refresh error, got %v", err)
	}

	var tokenCount int64
	if err := db.Model(&entity.RefreshToken{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error; err != nil {
		t.Fatalf("count refresh tokens: %v", err)
	}
	if tokenCount != 1 {
		t.Fatalf("blocked refresh attempt should roll back, got %d tokens", tokenCount)
	}
}

func TestToggleUserBlockRevokesRefreshTokens(t *testing.T) {
	db := newTestDB(t)
	user := mustCreateUser(t, db, "Dana", "dana@example.com", "user", true)

	plainOne, hashedOne, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token one: %v", err)
	}
	plainTwo, hashedTwo, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token two: %v", err)
	}
	_ = plainOne
	_ = plainTwo

	if err := db.Create(&entity.RefreshToken{
		UserID:    user.ID,
		Token:     hashedOne,
		ExpiredAt: time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("insert refresh token one: %v", err)
	}
	if err := db.Create(&entity.RefreshToken{
		UserID:    user.ID,
		Token:     hashedTwo,
		ExpiredAt: time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("insert refresh token two: %v", err)
	}

	adminUsecase := NewAdminUsecase(repository.NewUserRepository(db), repository.NewRoleRepository(db), db)
	name, blocked, err := adminUsecase.ToggleUserBlock(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("toggle user block: %v", err)
	}
	if name != user.FullName {
		t.Fatalf("expected user name %q, got %q", user.FullName, name)
	}
	if !blocked {
		t.Fatal("expected user to be blocked")
	}

	var tokenCount int64
	if err := db.Model(&entity.RefreshToken{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error; err != nil {
		t.Fatalf("count refresh tokens: %v", err)
	}
	if tokenCount != 0 {
		t.Fatalf("expected refresh tokens to be revoked, got %d", tokenCount)
	}

	var updated entity.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if !updated.IsBlocked {
		t.Fatal("expected user to be blocked in database")
	}
}

func TestAssignRoleToUserSyncsUserRolesAndCachedRole(t *testing.T) {
	db := newTestDB(t)
	userRole := mustCreateRole(t, db, "user")
	agencyRole := mustCreateRole(t, db, "agency")
	user := mustCreateUser(t, db, "Erin", "erin@example.com", "user", true)

	if err := db.Create(&entity.UserRole{
		UserID:    user.ID,
		RoleID:    userRole.ID,
		IsPrimary: true,
	}).Error; err != nil {
		t.Fatalf("seed primary role: %v", err)
	}

	roleUsecase := NewRoleUsecase(repository.NewRoleRepository(db), repository.NewUserRepository(db), db)
	if err := roleUsecase.AssignRoleToUser(context.Background(), user.ID, agencyRole.ID); err != nil {
		t.Fatalf("assign role: %v", err)
	}

	var updated entity.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.Role != "agency" {
		t.Fatalf("expected cached user role to be agency, got %s", updated.Role)
	}

	var links []entity.UserRole
	if err := db.Where("user_id = ?", user.ID).Find(&links).Error; err != nil {
		t.Fatalf("load user roles: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one join row, got %d", len(links))
	}
	if links[0].RoleID != agencyRole.ID {
		t.Fatalf("expected agency role id %d, got %d", agencyRole.ID, links[0].RoleID)
	}

	roles, err := roleUsecase.GetUserRoles(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user roles: %v", err)
	}
	if len(roles) != 1 || roles[0].Name != "agency" {
		t.Fatalf("expected agency role from lookup, got %#v", roles)
	}
}

func TestAITripReviewApprovalIsTransactional(t *testing.T) {
	db := newTestDB(t)
	requester := mustCreateUser(t, db, "Frank", "frank@example.com", "user", true)
	admin := mustCreateUser(t, db, "Admin", "admin@example.com", "admin", true)

	notificationUsecase := NewNotificationUsecase(repository.NewNotificationRepository(db))
	aiUsecase := NewAITripRequestUsecase(
		repository.NewAITripRequestRepository(db),
		repository.NewTripRepository(db),
		repository.NewUserRepository(db),
		notificationUsecase,
		db,
	)

	request := entity.AITripRequest{
		UserID:        requester.ID,
		From:          "Delhi",
		To:            "Goa",
		Days:          3,
		TripType:      "Family",
		BudgetLevel:   "Medium",
		Members:       2,
		Children:      0,
		HotelType:     "3 Star",
		Transport:     "Car",
		Prompt:        "Plan a Goa trip",
		GeneratedPlan: "Day 1 itinerary",
		Status:        entity.AITripStatusPending,
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("create ai trip request: %v", err)
	}

	reviewed, err := aiUsecase.ReviewRequest(context.Background(), admin.ID, request.ID, true, "approved")
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}
	if reviewed.Status != entity.AITripStatusApproved {
		t.Fatalf("expected approved request, got %s", reviewed.Status)
	}
	if reviewed.TripID == nil {
		t.Fatal("expected approved request to reference created trip")
	}

	var tripCount int64
	if err := db.Model(&entity.Trip{}).Count(&tripCount).Error; err != nil {
		t.Fatalf("count trips: %v", err)
	}
	if tripCount != 1 {
		t.Fatalf("expected one trip to be created, got %d", tripCount)
	}

	var notificationCount int64
	if err := db.Model(&entity.Notification{}).
		Where("ai_trip_request_id = ? AND type = ?", request.ID, "ai_request").
		Count(&notificationCount).Error; err != nil {
		t.Fatalf("count ai notifications: %v", err)
	}
	if notificationCount != 1 {
		t.Fatalf("expected one ai notification, got %d", notificationCount)
	}

	if _, err := aiUsecase.ReviewRequest(context.Background(), admin.ID, request.ID, true, "approved twice"); err == nil || !strings.Contains(strings.ToLower(err.Error()), "already been reviewed") {
		t.Fatalf("expected duplicate approval to fail, got %v", err)
	}
}

func TestBookTripRoundsMoneyAndCreatesAdminNotification(t *testing.T) {
	db := newTestDB(t)
	user := mustCreateUser(t, db, "Grace", "grace@example.com", "user", true)
	mustCreateUser(t, db, "Admin", "admin-booking@example.com", "admin", true)
	trip := mustCreateTrip(t, db, 12.345, time.Now().Add(72*time.Hour))
	offer := mustCreateOffer(t, db, "SAVE10", 10)

	notificationUsecase := NewNotificationUsecase(repository.NewNotificationRepository(db))
	bookingUsecase := NewBookingUsecase(
		repository.NewBookingRepository(db),
		repository.NewTripRepository(db),
		repository.NewUserRepository(db),
		repository.NewOfferRepository(db),
		db,
		notificationUsecase,
	)

	booking, err := bookingUsecase.BookTrip(context.Background(), trip.ID, user.ID, 2, offer.Code, nil, nil)
	if err != nil {
		t.Fatalf("book trip: %v", err)
	}

	if !floatEquals(booking.BaseAmount, 24.70) {
		t.Fatalf("expected rounded base amount 24.70, got %.2f", booking.BaseAmount)
	}
	if !floatEquals(booking.FinalAmount, 22.23) {
		t.Fatalf("expected rounded final amount 22.23, got %.2f", booking.FinalAmount)
	}

	var userNotification entity.Notification
	if err := db.Where("user_id = ? AND type = ?", user.ID, "booking_created").First(&userNotification).Error; err != nil {
		t.Fatalf("load user booking notification: %v", err)
	}

	var adminNotification entity.Notification
	if err := db.Where("type = ? AND booking_id = ?", "admin_booking", booking.ID).First(&adminNotification).Error; err != nil {
		t.Fatalf("load admin booking notification: %v", err)
	}
	if !adminNotification.IsAdmin {
		t.Fatal("expected admin booking notification to be flagged for admins")
	}
}
