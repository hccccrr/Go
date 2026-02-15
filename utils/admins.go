package utils

import (
	"context"
)

// TelegramClient interface for Telegram operations
type TelegramClient interface {
	GetAdmins(ctx context.Context, chatID int64) ([]int64, error)
	GetUserPermissions(ctx context.Context, chatID int64, userID int64) (*UserPermissions, error)
	GetEntity(ctx context.Context, chatID int64) (*ChatEntity, error)
}

// Database interface for authorized users
type AdminDatabase interface {
	GetAllAuthUsers(ctx context.Context, chatID int64) ([]int64, error)
}

// UserPermissions represents user's permissions in a chat
type UserPermissions struct {
	IsAdmin    bool
	ManageCall bool // Voice chat management permission
}

// ChatEntity represents a chat
type ChatEntity struct {
	ID    int64
	Title string
	Type  string // "user", "group", "supergroup", "channel"
}

// AdminManager handles admin and authorization operations
type AdminManager struct {
	client TelegramClient
	db     AdminDatabase
}

// NewAdminManager creates a new AdminManager
func NewAdminManager(client TelegramClient, db AdminDatabase) *AdminManager {
	return &AdminManager{
		client: client,
		db:     db,
	}
}

// GetAdmins retrieves all admins in a chat
func (am *AdminManager) GetAdmins(ctx context.Context, chatID int64) ([]int64, error) {
	return am.client.GetAdmins(ctx, chatID)
}

// GetAuthUsers retrieves all authorized users (admins + custom auth users)
func (am *AdminManager) GetAuthUsers(ctx context.Context, chatID int64) ([]int64, error) {
	authUsers := []int64{}

	// Get admins first
	admins, err := am.client.GetAdmins(ctx, chatID)
	if err == nil {
		authUsers = append(authUsers, admins...)
	}

	// Get custom authorized users from database
	customAuth, err := am.db.GetAllAuthUsers(ctx, chatID)
	if err == nil && customAuth != nil {
		// Merge without duplicates
		authMap := make(map[int64]bool)
		for _, id := range authUsers {
			authMap[id] = true
		}
		for _, id := range customAuth {
			if !authMap[id] {
				authUsers = append(authUsers, id)
			}
		}
	}

	return authUsers, nil
}

// GetUserRights checks if user has manage voice chats permission
func (am *AdminManager) GetUserRights(ctx context.Context, chatID int64, userID int64) (bool, error) {
	perms, err := am.client.GetUserPermissions(ctx, chatID, userID)
	if err != nil {
		return false, err
	}

	// Check if user is admin
	if !perms.IsAdmin {
		return false, nil
	}

	// Check for manage_call permission (voice chat management)
	if perms.ManageCall {
		return true, nil
	}

	// If manage_call not available, allow if user is admin
	return perms.IsAdmin, nil
}

// GetUserType returns user type: "admin", "auth", or "user"
func (am *AdminManager) GetUserType(ctx context.Context, chatID int64, userID int64) (string, error) {
	admins, err := am.GetAdmins(ctx, chatID)
	if err != nil {
		return "user", err
	}

	// Check if user is admin
	for _, adminID := range admins {
		if adminID == userID {
			return "admin", nil
		}
	}

	// Check if user is in auth list
	authUsers, err := am.GetAuthUsers(ctx, chatID)
	if err != nil {
		return "user", err
	}

	for _, authID := range authUsers {
		if authID == userID {
			return "auth", nil
		}
	}

	return "user", nil
}

// IsAdmin checks if user is an admin
func (am *AdminManager) IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error) {
	userType, err := am.GetUserType(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	return userType == "admin", nil
}

// IsAuthUser checks if user is authorized (admin or custom auth)
func (am *AdminManager) IsAuthUser(ctx context.Context, chatID int64, userID int64) (bool, error) {
	userType, err := am.GetUserType(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	return userType == "admin" || userType == "auth", nil
}

// CanManageVC checks if user can manage voice chats
func (am *AdminManager) CanManageVC(ctx context.Context, chatID int64, userID int64) (bool, error) {
	// First check if user is authorized
	isAuth, err := am.IsAuthUser(ctx, chatID, userID)
	if err != nil {
		return false, err
	}
	if !isAuth {
		return false, nil
	}

	// Then check specific permissions
	return am.GetUserRights(ctx, chatID, userID)
}

// Helper function to check if user ID exists in slice
func contains(slice []int64, item int64) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
