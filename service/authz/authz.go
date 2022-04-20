package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.sicepat.tech/platform/golib/log"
)

type Authz struct {
	UserUUID   string
	UserID     string
	ClientName string
	RoleName   string
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
	InputUserRole struct {
		UserID        string   `json:"user_id"`
		Branch        string   `json:"branch"`
		UserType      string   `json:"user_type"`
		ClientRoleIDs []string `json:"client_role_ids"`
	}

	InsertUserRoleRequest struct {
		Input InputUserRole `json:"input"`
	}

	ClientResp struct {
		Name string `json:"name"`
	}

	RoleResp struct {
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

	ReadUserRolesResponse struct {
		UserID        string   `bson:"userID" json:"user_id"`
		Type          string   `bson:"type" json:"type"`
		ClientRoleIDs []string `bson:"clientRoleIDs" json:"client_role_ids"`
	}

	PermissionResp struct {
		Users []ReadUserRolesResponse `json:"users"`
	}

	UserRoleResponse struct {
		Data PermissionResp `json:"data"`
	}
)

const (
	clientApp              = "masterdata"
	endpointAuthzV2Staging = "https://api.s.sicepat.io/v2/authz/management"
	endpointAuthzV2Prod    = "https://api.sicepat.io/v2/authz/management"
	endpointAuthzV1Staging = "https://api.s.sicepat.io/v1/authz"
	endpointAuthzV1Prod    = "https://api.sicepat.io/v1/authz"
)

func AuthzGetUserID(ctx context.Context, req *Authz) (userData UserData, err error) {

	client := &http.Client{}

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?userID=%s", endpointAuthzV2Staging, req.UserID), nil)
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

	req.RoleName = strings.ReplaceAll(req.RoleName, " ", "%20")
	url := fmt.Sprintf("%s/client-roles?client=%s&role=%s", endpointAuthzV2Prod, clientApp, req.RoleName)
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
	url := fmt.Sprintf("%s/users", endpointAuthzV2Staging)
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

func AuthzGetUserRoles(ctx context.Context, limit, offset int) (data UserRoleResponse, err error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s/user-roles?limit=%d&offset=%d", endpointAuthzV2Prod, limit, offset)
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

func AuthzInsertUserRoles(ctx context.Context, userID string) error {

	client := &http.Client{}

	request := InsertUserRoleRequest{
		Input: InputUserRole{
			UserID:        userID,
			Branch:        "Default",
			UserType:      "internal",
			ClientRoleIDs: []string{"a067a2b1-f652-469b-873b-9cd8ab020931"},
		}}

	log.Printf("REQ: %+v", request)

	toByte, _ := json.Marshal(request)

	url := fmt.Sprintf("%s/user-roles", endpointAuthzV2Prod)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(toByte))
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

	var res interface{}
	err = decodeResponse(resp.Body, &res)
	if err != nil {
		return err
	}

	log.Printf("%+v", res)
	return nil
}

func decodeResponse(b io.Reader, v interface{}) error {
	return json.NewDecoder(b).Decode(&v)
}
