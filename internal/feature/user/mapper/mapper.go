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
	info := entity.UserInfo{
		ID:                 u.ID,
		Username:           u.Username,
		Role:               u.Role,
		TeamID:             u.TeamID,
		DisplayName:        u.DisplayName,
		Position:           u.Position,
		PhoneNumber:        u.PhoneNumber,
		IsActive:           u.IsActive,
		MustChangePassword: u.MustChangePassword,
		LastLogin:          &lastLogin,
		CreatedAt:          u.CreatedAt.Format(time.RFC3339),
	}
	if u.Team != nil {
		info.Team = &entity.TeamInfo{ID: u.Team.ID, Name: u.Team.Name}
	}
	return info
}

func ToResponse(info entity.UserInfo) dto.UserResponse {
	resp := dto.UserResponse{
		ID:                 info.ID,
		Username:           info.Username,
		Role:               info.Role,
		TeamID:             info.TeamID,
		DisplayName:        info.DisplayName,
		Position:           info.Position,
		PhoneNumber:        info.PhoneNumber,
		IsActive:           info.IsActive,
		MustChangePassword: info.MustChangePassword,
		LastLogin:          info.LastLogin,
		CreatedAt:          info.CreatedAt,
	}
	if info.Team != nil {
		resp.Team = &dto.TeamNested{ID: info.Team.ID, Name: info.Team.Name}
	}
	return resp
}

func ToContactResponse(info entity.ContactInfo, actor entity.Actor) dto.ContactDirectoryResponse {
	resp := dto.ContactDirectoryResponse{
		ID:          info.ID,
		Username:    info.Username,
		DisplayName: info.DisplayName,
		Position:    info.Position,
		PhoneNumber: info.PhoneNumber,
		Role:        info.Role,
		TeamID:      info.TeamID,
		IsActive:    info.IsActive,
		UpdatedAt:   info.UpdatedAt,
		Actions: dto.ContactDirectoryActions{
			CanEdit:           actor.Role == "super_admin" || actor.ID == info.ID,
			CanEditRoleOrTeam: actor.Role == "super_admin",
		},
	}
	if info.Team != nil {
		resp.Team = &dto.TeamNested{ID: info.Team.ID, Name: info.Team.Name}
	}
	return resp
}
