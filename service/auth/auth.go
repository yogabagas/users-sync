package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type EntityAttr struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	UserID      int    `json:"user_id"`
	UserPassID  string `json:"userpass_id"`
	NIK         string `json:"nik"`
}

type Entity struct {
	ID         string     `json:"id"`
	Attributes EntityAttr `json:"attributes"`
	CreatedAt  string     `json:"created_at"`
	Active     bool       `json:"active"`
}

type EntityData struct {
	Data []Entity `json:"data"`
}

type UpdateNIK struct {
	NIK string `json:"nik"`
}

func Process(ctx context.Context, userID int, username string) (*Entity, error) {
	userEntity, err := getEntity(ctx, userID, username)
	if err != nil {
		return nil, err
	}

	if userEntity != nil && userEntity.Attributes.NIK == "" {
		err = updateEntityAttr(ctx, userEntity.ID, &UpdateNIK{NIK: username})
		if err != nil {
			return nil, err
		}
	}

	return userEntity, nil
}

func getEntity(ctx context.Context, userID int, username string) (*Entity, error) {
	url := fmt.Sprintf("https://api.sicepat.io/v1/auth/entity?attributes.user_id=%d", userID)
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("get entity failed")
	}

	var entity []Entity
	err = json.NewDecoder(resp.Body).Decode(&entity)
	if err != nil {
		return nil, err
	}

	if len(entity) == 0 {
		return nil, errors.New("user not found")
	}

	for _, v := range entity {
		if v.Attributes.UserPassID == username {
			return &v, nil
		}
	}

	return nil, nil
}

func updateEntityAttr(ctx context.Context, entityID string, data *UpdateNIK) error {
	url := fmt.Sprintf("https://api.sicepat.io/v1/auth/entity/%s/attributes", entityID)

	req, err := http.NewRequest(http.MethodPut, url, ConvertStructToIOReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", ctx.Value("token").(string))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid token")
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("update nik failed")
	}

	return nil
}

func ConvertStructToIOReader(req interface{}) *bytes.Reader {
	reqByte, _ := json.Marshal(req)
	return bytes.NewReader(reqByte)
}
