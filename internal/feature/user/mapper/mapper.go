package mapper

import (
	"time"

	"backend-hotlines3/internal/feature/user/dto"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(u models.User) entity.UserInfo {
	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format(time.RFC3339)
	}
	return entity.UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		TeamID:    u.TeamID,
		IsActive:  u.IsActive,
		LastLogin: &lastLogin,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func ToResponse(info entity.UserInfo) dto.UserResponse {
	return dto.UserResponse{
		ID:        info.ID,
		Username:  info.Username,
		Role:      info.Role,
		TeamID:    info.TeamID,
		IsActive:  info.IsActive,
		LastLogin: info.LastLogin,
		CreatedAt: info.CreatedAt,
	}
}
