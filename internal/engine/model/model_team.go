package model

import (
	"github.com/bytedance/sonic"
	"github.com/observabil/arcade/internal/engine/model/entity"
)

// CreateTeamReq create team request
type CreateTeamReq struct {
	OrgId        string                 `json:"orgId" validate:"required"`
	Name         string                 `json:"name" validate:"required,min=2,max=64"`
	DisplayName  string                 `json:"displayName"`
	Description  string                 `json:"description"`
	Avatar       string                 `json:"avatar"`
	ParentTeamId string                 `json:"parentTeamId"`
	Settings     map[string]interface{} `json:"settings"`
	Visibility   int                    `json:"visibility"` // 0-private, 1-internal, 2-public
}

// UpdateTeamReq update team request
type UpdateTeamReq struct {
	Name        *string                `json:"name,omitempty"`
	DisplayName *string                `json:"displayName,omitempty"`
	Description *string                `json:"description,omitempty"`
	Avatar      *string                `json:"avatar,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Visibility  *int                   `json:"visibility,omitempty"`
	IsEnabled   *int                   `json:"isEnabled,omitempty"`
}

// TeamQueryReq query team request
type TeamQueryReq struct {
	OrgId        string `json:"orgId" form:"orgId"`
	Name         string `json:"name" form:"name"`
	ParentTeamId string `json:"parentTeamId" form:"parentTeamId"`
	Visibility   *int   `json:"visibility" form:"visibility"`
	IsEnabled    *int   `json:"isEnabled" form:"isEnabled"`
	Page         int    `json:"page" form:"page"`
	PageSize     int    `json:"pageSize" form:"pageSize"`
}

// TeamResp team response
type TeamResp struct {
	TeamId        string                 `json:"teamId"`
	OrgId         string                 `json:"orgId"`
	Name          string                 `json:"name"`
	DisplayName   string                 `json:"displayName"`
	Description   string                 `json:"description"`
	Avatar        string                 `json:"avatar"`
	ParentTeamId  string                 `json:"parentTeamId"`
	Path          string                 `json:"path"`
	Level         int                    `json:"level"`
	Settings      map[string]interface{} `json:"settings"`
	Visibility    int                    `json:"visibility"`
	IsEnabled     int                    `json:"isEnabled"`
	TotalMembers  int                    `json:"totalMembers"`
	TotalProjects int                    `json:"totalProjects"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
}

// TeamListResp team list response
type TeamListResp struct {
	Teams      []*TeamResp `json:"teams"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// EntityToTeamResp convert entity.Team to TeamResp
func EntityToTeamResp(team *entity.Team) *TeamResp {
	if team == nil {
		return nil
	}

	resp := &TeamResp{
		TeamId:        team.TeamId,
		OrgId:         team.OrgId,
		Name:          team.Name,
		DisplayName:   team.DisplayName,
		Description:   team.Description,
		Avatar:        team.Avatar,
		ParentTeamId:  team.ParentTeamId,
		Path:          team.Path,
		Level:         team.Level,
		Visibility:    team.Visibility,
		IsEnabled:     team.IsEnabled,
		TotalMembers:  team.TotalMembers,
		TotalProjects: team.TotalProjects,
		CreatedAt:     team.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     team.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// parse Settings JSON
	if len(team.Settings) > 0 {
		settings := make(map[string]interface{})
		if err := sonic.Unmarshal(team.Settings, &settings); err == nil {
			resp.Settings = settings
		}
	}

	return resp
}
