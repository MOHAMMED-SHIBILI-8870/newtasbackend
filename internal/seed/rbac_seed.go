package seed

import (
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/usecase"
	"strings"
)

func SeedRBAC() {
	db := config.DB

	roles := []entity.Role{
		{Name: "admin", Description: "Platform administrator"},
		{Name: "agency", Description: "Travel agency operator"},
		{Name: "guide", Description: "Trip guide"},
		{Name: "driver", Description: "Trip driver"},
		{Name: "support", Description: "Support staff"},
		{Name: "user", Description: "Standard customer"},
	}

	permissions := []entity.Permission{
		{Key: "manage_users", Name: "Manage Users", Description: "Create and manage users and roles"},
		{Key: "manage_bookings", Name: "Manage Bookings", Description: "Manage bookings and reservations"},
		{Key: "manage_tracking", Name: "Manage Tracking", Description: "Update and view live trip tracking"},
		{Key: "manage_chat", Name: "Manage Chat", Description: "Access and moderate support chat"},
		{Key: "manage_offers", Name: "Manage Offers", Description: "Create and manage offers and coupons"},
		{Key: "manage_fleet", Name: "Manage Fleet", Description: "Create and manage vehicles"},
		{Key: "manage_reviews", Name: "Manage Reviews", Description: "Moderate reviews and ratings"},
		{Key: "manage_complaints", Name: "Manage Complaints", Description: "View and resolve complaints"},

		// chat permission

		{Key: "chat.read", Name: "Read Chats", Description: "Can view chats"},
		{Key: "chat.send", Name: "Send Messages", Description: "Can send messages"},
		{Key: "chat.delete", Name: "Delete Messages", Description: "Can delete messages"},
	}

	for _, role := range roles {
		item := role
		if err := db.Where("LOWER(name) = LOWER(?)", role.Name).FirstOrCreate(&item).Error; err != nil {
			continue
		}
	}

	for _, permission := range permissions {
		item := permission
		if err := db.Where("LOWER(key) = LOWER(?)", permission.Key).FirstOrCreate(&item).Error; err != nil {
			continue
		}
	}

	assignments := map[string][]string{
		"admin":   {"manage_users", "manage_bookings", "manage_tracking", "manage_chat", "manage_offers", "manage_fleet", "manage_reviews", "manage_complaints","chat.read","chat.send","chat.delete"},
		"agency":  {"manage_bookings", "manage_offers", "manage_fleet","chat.read","chat.send"},
		"guide":   {"manage_bookings", "manage_tracking", "manage_chat","chat.read","chat.send"},
		"driver":  {"manage_tracking","chat.read","chat.send"},
		"support": {"manage_chat", "manage_complaints","chat.read","chat.send"},
		"user":    {"chat.read","chat.send"},
	}

	for roleName, keys := range assignments {
		var role entity.Role
		if err := db.Where("LOWER(name) = LOWER(?)", roleName).First(&role).Error; err != nil {
			continue
		}

		for _, key := range keys {
			var permission entity.Permission
			if err := db.Where("LOWER(key) = LOWER(?)", key).First(&permission).Error; err != nil {
				continue
			}

			link := entity.RolePermission{RoleID: role.ID, PermissionID: permission.ID}
			_ = db.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&link).Error
		}
	}

	var users []entity.User
	if err := db.Find(&users).Error; err != nil {
		return
	}

	for _, user := range users {
		roleName := usecase.NormalizeRole(user.Role)
		var role entity.Role
		if err := db.Where("LOWER(name) = LOWER(?)", roleName).First(&role).Error; err != nil {
			continue
		}

		_ = db.Where("user_id = ?", user.ID).Delete(&entity.UserRole{}).Error
		_ = db.Create(&entity.UserRole{
			UserID:    user.ID,
			RoleID:    role.ID,
			IsPrimary: true,
		}).Error
	}
}

func normalizeSeedName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
