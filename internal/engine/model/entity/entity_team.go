package entity

import "gorm.io/datatypes"

// Team team entity
type Team struct {
	BaseModel
	TeamId       string         `gorm:"column:team_id" json:"teamId"`              // team unique identifier
	OrgId        string         `gorm:"column:org_id" json:"orgId"`                // organization id
	Name         string         `gorm:"column:name" json:"name"`                   // team name
	DisplayName  string         `gorm:"column:display_name" json:"displayName"`    // team display name
	Description  string         `gorm:"column:description" json:"description"`     // team description
	Avatar       string         `gorm:"column:avatar" json:"avatar"`               // team avatar
	ParentTeamId string         `gorm:"column:parent_team_id" json:"parentTeamId"` // parent team id(supports nested)
	Path         string         `gorm:"column:path" json:"path"`                   // team path(for hierarchical relationship)
	Level        int            `gorm:"column:level" json:"level"`                 // team level
	Settings     datatypes.JSON `gorm:"column:settings;type:json" json:"settings"` // team settings
	Visibility   int            `gorm:"column:visibility" json:"visibility"`       // visibility: 0-private, 1-internal, 2-public
	IsEnabled    int            `gorm:"column:is_enabled" json:"isEnabled"`        // is enabled: 0-disable, 1-enable

	// statistics fields
	TotalMembers  int `gorm:"column:total_members" json:"totalMembers"`   // total members
	TotalProjects int `gorm:"column:total_projects" json:"totalProjects"` // total projects
}

func (Team) TableName() string {
	return "t_team"
}

// TeamSettings team settings struct
type TeamSettings struct {
	DefaultRole       string `json:"default_role"`        // default role
	AllowMemberInvite bool   `json:"allow_member_invite"` // allow member invite
	RequireApproval   bool   `json:"require_approval"`    // require approval
	MaxMembers        int    `json:"max_members"`         // max members
}

// TeamVisibility team visibility enum
const (
	TeamVisibilityPrivate  = 0 // private(only members can see)
	TeamVisibilityInternal = 1 // internal(only organization members can see)
	TeamVisibilityPublic   = 2 // public(everyone can see)
)
