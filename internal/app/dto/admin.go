package dto

import "time"

type AppState struct {
	IsSuperAdmin bool `json:"is_super_admin"`
}

type AdminUsersRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type AdminUser struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	OpenID       string    `json:"open_id"`
	Phone        string    `json:"phone"`
	Avatar       string    `json:"avatar"`
	CreatedAt    time.Time `json:"created_at"`
	IsBlocked    bool      `json:"is_blocked"`
	IsSuperAdmin bool      `json:"is_super_admin"`
}

type AdminUsersResponse struct {
	Items    []AdminUser `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type BlockUserRequest struct {
	UserID uint `json:"user_id"`
}

type BlockUserResponse struct {
	IsBlocked          bool `json:"is_blocked"`
	DisconnectNotified bool `json:"disconnect_notified"`
}

type AdminMatchStats struct {
	Date     string           `json:"date"`
	Timezone string           `json:"timezone"`
	Total    int64            `json:"total"`
	ByType   map[string]int64 `json:"by_type"`
}
