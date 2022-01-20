package masterdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Data struct {
	NIK        string `json:"nik"`
	Nama       string `json:"nama"`
	Role       string `json:"role"`
	Direktorat string `json:"direktorat"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	NIK      string `json:"nik"`
	Username string `json:""`
}

type UserResponse struct {
	Users []User `json:"users"`
}

type UserData struct {
	Data UserResponse `json:"data"`
}

func SearchUserByNIK(ctx context.Context, nik string) (*User, error) {
	url := fmt.Sprintf("https://api.sicepat.io/v1/masterdata/users?q=%s&limit=%d", nik, 50)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", ctx.Value("token").(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid token")
	}

	var users UserData
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if len(users.Data.Users) > 0 {
		for _, v := range users.Data.Users {
			if v.Username == nik || v.NIK == nik {
				return &v, nil
			}
		}
	}

	return nil, errors.New("nik not found")
}
