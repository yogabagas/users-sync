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
		ClientRoles []*ClientRole `json:"client_roles"`
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
		ID    string       `json:"id"`
		Input []*UserRoles `json:"input"`
	}

	UserRoles struct {
		BranchID     string `json:"branch_id"`
		ClientRoleID string `json:"client_role_id"`
	}
)

const (
	clientApp            = "hr"
	endpointAuthzStaging = "https://api.s.sicepat.io/v2/authz/management"
	endpointAuthzProd    = "https://api.sicepat.io/v2/authz/management"
)

func AuthzGetUserID(ctx context.Context, req *Authz) (userData UserData, err error) {

	client := &http.Client{}

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?userID=%s", endpointAuthzStaging, req.UserID), nil)
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
	url := fmt.Sprintf("%s/client-roles?client=%s&role=%s", endpointAuthzStaging, clientApp, req.RoleName)
	fmt.Println(url)
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
	url := fmt.Sprintf("%s/users", endpointAuthzStaging)
	fmt.Println(url)
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

func AuthzInsertUserRoles(ctx context.Context, req *Authz, clientRoleData *ClientRoleData, userData *UserData) error {

	client := &http.Client{}

	var clientRoleID string
	for _, v := range clientRoleData.Data.ClientRoles {
		if v.Client.Name != clientApp && v.Role.Name != req.RoleName {
			continue
		}
		clientRoleID = v.ID
	}

	// log.Printf("%+v", userData)

	// var request *InputUserRole
	// for _, user := range userData.Data.Users {
	request := &InputUserRole{
		ID: userData.Data.Users[0].ID,
		Input: []*UserRoles{
			{
				BranchID:     "4de69001-4e56-4c67-a074-0fce84bd43cd",
				ClientRoleID: clientRoleID,
			},
		},
	}
	// }

	log.Printf("%+v", request)

	toByte, _ := json.Marshal(request)

	url := fmt.Sprintf("%s/users/%s/roles", endpointAuthzStaging, userData.Data.Users[0].ID)
	fmt.Println(url)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(toByte))
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

func decodeResponse(b io.Reader, v interface{}) error {
	if err := json.NewDecoder(b).Decode(&v); err != nil {
		return err
	}
	return nil
}
