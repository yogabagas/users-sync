package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.sicepat.tech/platform/golib/log"
)

type Authz struct {
	UserUUID string
	UserID   string
	ClientID string
	RoleID   string
}

type (
	UserData struct {
		Data UsersResponse `json:"data"`
	}
	UsersResponse struct {
		Users []User `json:"users"`
	}

	User struct {
		ID        string    `json:"id"`
		UserID    string    `json:"user_id"`
		Type      string    `json:"type"`
		IsDeleted bool      `json:"is_deleted"`
		CreatedBy string    `json:"created_by"`
		UpdatedBy string    `json:"updated_by"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type (
	ClientRoleData struct {
		Data ClientRoleResponse `json:"data"`
	}
	ClientRoleResponse struct {
		ClientRoles []ClientRole `json:"client_roles"`
	}
	ClientRole struct {
		ID     string `json:"id"`
		Client struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"client"`
		Role struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		IsDeleted bool      `json:"is_deleted"`
		CreatedBy string    `json:"created_by"`
		UpdatedBy string    `json:"updated_by"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type (
	InsertUser struct {
		Input UserRequest `json:"input"`
	}

	UserRequest struct {
		UserID string `json:"user_id"`
		Type   string `json:"type"`
	}
)

type (
	UpdateUserRoleRequest struct {
		Input InputUserRole `json:"input"`
	}
	InputUserRole struct {
		UserID        string   `json:"user_id"`
		UserType      string   `json:"user_type"`
		Branch        string   `json:"branch"`
		ClientRoleIDs []string `json:"client_role_ids"`
		IsActive      bool     `json:"is_active"`
		Upsert        bool     `json:"upsert"`
	}

	ClientResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	RoleResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	BranchResp struct {
		Name string `json:"name"`
	}
	Permission struct {
		ID         string     `json:"id"`
		Client     ClientResp `json:"client"`
		Role       RoleResp   `json:"role"`
		BranchResp BranchResp `json:"branch_resp"`
	}

	PermissionResp struct {
		Permissions []Permission `json:"permissions"`
	}

	UserRoleResponse struct {
		Data PermissionResp `json:"data"`
	}
)

type (
	Role struct {
		ID        string    `bson:"_id" json:"id"`
		RoleID    int       `bson:"roleID" json:"role_id"`
		RoleName  string    `bson:"roleName" json:"role_name"`
		IsDeleted bool      `bson:"isDeleted" json:"is_deleted"`
		CreatedAt time.Time `bson:"createdAt" json:"created_at"`
		CreatedBy string    `bson:"createdBy" json:"created_by"`
		UpdatedAt time.Time `bson:"updatedAt" json:"updated_at"`
		UpdatedBy string    `bson:"updatedBy" json:"updated_by"`
		IsActive  bool      `bson:"isActive" json:"is_active"`
	}
	ListRoleResponse struct {
		Roles []Role `json:"roles"`
	}
	ListRoleData struct {
		Data ListRoleResponse `json:"data"`
	}
)

type (
	Client struct {
		ID        string    `bson:"_id" json:"id"`
		Name      string    `bson:"name" json:"name"`
		IsDeleted bool      `bson:"isDeleted" json:"is_deleted"`
		CreatedAt time.Time `bson:"createdAt" json:"created_at"`
		CreatedBy string    `bson:"createdBy" json:"created_by"`
		UpdatedAt time.Time `bson:"updatedAt" json:"updated_at"`
		UpdatedBy string    `bson:"updatedBy" json:"updated_by"`
		IsActive  bool      `bson:"isActive" json:"is_active"`
	}
	ListClientResponse struct {
		Clients []Client `json:"clients"`
	}
	ListClientData struct {
		Data ListClientResponse `json:"data"`
	}
)

const (
	clientApp              = "hr"
	endpointAuthzV2Staging = "https://api.s.sicepat.io/v2/authz/management"
	endpointAuthzV2Prod    = "https://api.sicepat.io/v2/authz/management"
	endpointAuthzV1Staging = "https://api.s.sicepat.io/v1/authz"
	endpointAuthzV1Prod    = "https://api.sicepat.io/v1/authz"
)

func AuthzGetUserID(ctx context.Context, req *Authz) (userData UserData, err error) {

	client := &http.Client{}

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?userID=%s", endpointAuthzV2Prod, req.UserID), nil)
	if err != nil {
		return
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = decodeResponse(resp.Body, &userData)

	return

}

func AuthzGetClientRoleID(ctx context.Context, req *Authz) (clientRoleData ClientRoleData, err error) {

	client := &http.Client{}

	url := fmt.Sprintf("%s/client-roles?clientID=%s&roleID=%s", endpointAuthzV2Prod, req.ClientID, req.RoleID)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = decodeResponse(resp.Body, &clientRoleData)

	return

}

func AuthzInsertUser(ctx context.Context, req *Authz) error {

	client := &http.Client{}

	request := &InsertUser{
		Input: UserRequest{
			UserID: req.UserID,
			Type:   "internal",
		},
	}
	toByte, _ := json.Marshal(request)

	requestBody := bytes.NewBuffer(toByte)
	url := fmt.Sprintf("%s/users", endpointAuthzV2Prod)
	httpReq, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func AuthzGetUserRoles(ctx context.Context, userID string) (data UserRoleResponse, err error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s/users/%s/roles", endpointAuthzV2Prod, userID)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("ERR", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = decodeResponse(resp.Body, &data)
	return
}

func AuthzGetRole(ctx context.Context, roleName string) (data ListRoleData, err error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s/roles?name=%s&isActive=true", endpointAuthzV2Prod, roleName)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("ERR", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = decodeResponse(resp.Body, &data)
	return
}

func AuthzGetClient(ctx context.Context, clientName string) (data ListClientData, err error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s/clients?name=%s&isActive=true", endpointAuthzV2Prod, clientName)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("ERR", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = decodeResponse(resp.Body, &data)
	return
}

func AuthzUpdateUserRoles(ctx context.Context, clientRoleIDs []string, userID string) error {

	client := &http.Client{}

	request := &UpdateUserRoleRequest{
		Input: InputUserRole{
			UserID:        userID,
			UserType:      "internal",
			Branch:        "Default",
			ClientRoleIDs: clientRoleIDs,
			IsActive:      true,
			Upsert:        false,
		},
	}

	log.Printf("REQ: %+v", request)

	toByte, _ := json.Marshal(request)

	url := fmt.Sprintf("%s/user-roles", endpointAuthzV2Prod)
	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(toByte))
	if err != nil {
		log.Println("ERR", err.Error())
		return err
	}
	httpReq.Header.Set("Authorization", ctx.Value("token").(string))

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func decodeResponse(b io.Reader, v interface{}) error {
	return json.NewDecoder(b).Decode(&v)
}
