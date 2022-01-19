package masterdata

import (
	"encoding/json"
	"fmt"
	"my-github/users-sync/shared"
	"net/http"
)

type Data struct {
	NIK        string `json:"nik"`
	Nama       string `json:"nama"`
	Role       string `json:"role"`
	Direktorat string `json:"direktorat"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	NIK   string `json:"nik"`
}

type UserResponse struct {
	Users []User `json:"users"`
}

type UserData struct {
	Data UserResponse `json:"data"`
}

func SearchUserByNIK(nik string) (*User, error) {
	url := fmt.Sprintf("https://api.sicepat.io/v1/masterdata/users?q=%s", nik)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", shared.MasterDataToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var users UserData
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return nil, err
	}

	for _, v := range users.Data.Users {
		return &v, nil
	}

	return nil, nil
}
