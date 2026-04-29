package mapper

import (
	"time"

	"backend-hotlines3/internal/feature/auth/dto"
	"backend-hotlines3/internal/feature/auth/entity"
	"backend-hotlines3/internal/models"
)

func UserToInfo(u models.User) entity.UserInfo {
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

func InfoToResponse(info entity.UserInfo) dto.UserResponse {
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

func LoginResultToResponse(r entity.LoginResult) dto.LoginResponse {
	return dto.LoginResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		User:         InfoToResponse(r.User),
	}
}

func TokenPairToRefreshResponse(p entity.TokenPair) dto.RefreshResponse {
	return dto.RefreshResponse{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
	}
}
