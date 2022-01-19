package main

import (
	"context"
	"fmt"
	"log"
	"my-github/users-sync/config"
	"my-github/users-sync/repository"
	"my-github/users-sync/service"
	"my-github/users-sync/service/auth"
	"my-github/users-sync/service/authz"
	"my-github/users-sync/service/masterdata"
	"my-github/users-sync/shared"
)

func main() {
	config.InitDB()
	service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	userData, err := masterdata.SearchUserByNIK(ctx, "20050160")
	if err != nil {
		log.Println(err)
		repository.UpdateStatus(ctx, repository.LogData{
			NIK:         userData.NIK,
			Status:      int(shared.StatusFailInMasterData),
			Description: shared.StatusFailInMasterData.String(),
		})
	}

	_, err = auth.Process(ctx, userData.ID, userData.NIK)
	if err != nil {
		log.Println(err)
		repository.UpdateStatus(ctx, repository.LogData{
			NIK:         userData.NIK,
			Status:      int(shared.StatusFailInAuth),
			Description: shared.StatusFailInAuth.String(),
		})
	}

	_, err = authz.AuthzGetUserID(ctx, &authz.Authz{
		UserID: fmt.Sprint(userData.ID),
	})
	if err != nil {
		log.Println(err)
		repository.UpdateStatus(ctx, repository.LogData{
			NIK:         userData.NIK,
			Status:      int(shared.StatusFailInAuthz),
			Description: shared.StatusFailInAuthz.String(),
		})
	}

	repository.UpdateStatus(ctx, repository.LogData{Description: shared.StatusFinished.String()})
}
