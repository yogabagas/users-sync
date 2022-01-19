package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Authz struct {
	UserUUID   string
	UserID     string
	ClientName string
	RoleName   string
}

type (
	Data struct {
		Data *UsersResponse `json:"data"`
	}
	UsersResponse struct {
		Users []*User `json:"users"`
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

const (
	endpointAuthz = "https://api.sicepat.io/v2/authz/management"
)

func AuthzGetUserID(req *Authz) (data *Data, err error) {

	client := &http.Client{}

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?userID=%s", endpointAuthz, req.UserID), nil)
	if err != nil {
		return
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	newData := new(Data)

	decodeResponse(resp.Body, newData)

	return newData, nil

}


func AuthzGetClientRoleID

func decodeResponse(b io.Reader, v interface{}) {

	if err := json.NewDecoder(b).Decode(&v); err != nil {
		return
	}

}
