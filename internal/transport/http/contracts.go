package thttp

import "nerion/internal/entity"

// Me / Account

type updateMeRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Auth

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type resetRequestReq struct {
	Email string `json:"email"`
}

type resetPasswordReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Space

type createSpaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type renameSpaceRequest struct {
	Name string `json:"name"`
}

type deleteSpaceRequest struct {
	ConfirmName string `json:"confirm_name"`
}

type spaceResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	OwnerID    int64  `json:"owner_id"`
	TableCount int    `json:"table_count"`
}

func toSpaceResponse(sp *entity.Space, tableCount int) spaceResponse {
	return spaceResponse{
		ID:         sp.ID,
		Name:       sp.Name,
		Slug:       sp.Slug,
		OwnerID:    sp.OwnerID,
		TableCount: tableCount,
	}
}

// Member

type inviteMemberRequest struct {
	Email string `json:"email"`
}

type changeMemberRoleRequest struct {
	Role string `json:"role"`
}

type memberResponse struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"name"`
	UserEmail string `json:"email"`
	Role      string `json:"role"`
}

// Invite

type inviteInfoResponse struct {
	SpaceID   int64  `json:"space_id"`
	SpaceName string `json:"space_name"`
	Email     string `json:"email"`
}

// Schema

type createTableRequest struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	TemplateID string `json:"template_id,omitempty"`
}

type updateFieldsRequest struct {
	Fields []fieldRequest `json:"fields"`
}

type fieldRequest struct {
	Name                string   `json:"name"`
	Slug                string   `json:"slug"`
	Type                string   `json:"type"`
	Required            bool     `json:"required"`
	DefaultValue        *string  `json:"default_value,omitempty"`
	Unique              bool     `json:"unique"`
	EnumValues          []string `json:"enum_values,omitempty"`
	RelationTableID     *int64   `json:"relation_table_id,omitempty"`
	RelationCardinality *string  `json:"relation_cardinality,omitempty"`
	RelationTarget      *string  `json:"relation_target,omitempty"`
}

// PDF

type saveMappingRequest struct {
	Mappings []struct {
		Placeholder   string  `json:"placeholder"`
		SourceFieldID *int64  `json:"source_field_id"`
		Expression    *string `json:"expression"`
	} `json:"mappings"`
}

type pdfPreviewRequest struct {
	RecordID  int64  `json:"record_id"`
	TableSlug string `json:"table_slug"`
}

type pdfGenerateRequest struct {
	TemplateID int64   `json:"template_id"`
	TableSlug  string  `json:"table_slug"`
	RecordIDs  []int64 `json:"record_ids"`
}

// Lists

type createListRequest struct {
	Slug         string                   `json:"slug"`
	TableSlug    string                   `json:"table_slug"`
	FieldConfig  []entity.ListFieldConfig `json:"field_config"`
	FilterConfig map[string]any           `json:"filter_config"`
	SortConfig   []entity.ListSortConfig  `json:"sort_config"`
	RowLimit     int                      `json:"row_limit"`
	Published    bool                     `json:"published"`
}

type updateListRequest struct {
	FieldConfig  []entity.ListFieldConfig `json:"field_config"`
	FilterConfig map[string]any           `json:"filter_config"`
	SortConfig   []entity.ListSortConfig  `json:"sort_config"`
	RowLimit     *int                     `json:"row_limit"`
	Published    *bool                    `json:"published"`
}

// API Keys

type createAPIKeyRequest struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// User

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          string(u.Role),
		EmailVerified: u.EmailVerified,
	}
}
