package aliyundrive

import (
	"context"
	"encoding/json"
)

type GetUserInfoResponse struct {
	DomainID       string `json:"domain_id"`
	UserID         string `json:"user_id"`
	Avatar         string `json:"avatar"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
	Email          string `json:"email"`
	NickName       string `json:"nick_name"`
	Phone          string `json:"phone"`
	PhoneRegion    string `json:"phone_region"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	UserName       string `json:"user_name"`
	Description    string `json:"description"`
	DefaultDriveID string `json:"default_drive_id"`
	UserData       struct {
	} `json:"user_data"`
	DenyChangePasswordBySelf    bool        `json:"deny_change_password_by_self"`
	NeedChangePasswordNextLogin bool        `json:"need_change_password_next_login"`
	Creator                     string      `json:"creator"`
	ExpiredAt                   int         `json:"expired_at"`
	Permission                  interface{} `json:"permission"`
	DefaultLocation             string      `json:"default_location"`
	LastLoginTime               int64       `json:"last_login_time"`
}

func (c *Drive) DoGetUserInfoRequest(ctx context.Context) (*GetUserInfoResponse, error) {
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/user/get", Object{})
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
