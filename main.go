package main

import (
	"context"
	"fmt"
	"log"
	"my-github/users-sync/config"
	"my-github/users-sync/repository"
	"my-github/users-sync/service/auth"
	"my-github/users-sync/service/authz"
	"my-github/users-sync/service/masterdata"
	"my-github/users-sync/shared"
)

func main() {
	config.InitDB()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	userData, err := masterdata.SearchUserByNIK(ctx, "20050160")
	if err != nil {
		go repository.InsertLog(ctx, repository.LogData{Description: shared.StatusFailInMasterData.String()})
		return
	}

	fmt.Println("MASTER DATA", userData)

	userEntity, err := auth.Process(ctx, userData.ID, userData.NIK)
	if err != nil {
		go repository.InsertLog(ctx, repository.LogData{Description: shared.StatusFailInAuth.String()})
		return
	}

	fmt.Println("AUTHENTICATION", userEntity)

	userAuthz, err := authz.AuthzGetUserID(context.Background(), &authz.Authz{
		UserID: "83233",
	})
	if err != nil {
		go repository.InsertLog(ctx, repository.LogData{Description: shared.StatusFailInAuthz.String()})
		log.Println(err)
	}

	fmt.Println("AUTHORIZATION", userAuthz)

	log.Println(&userAuthz)

	go repository.InsertLog(ctx, repository.LogData{Description: shared.StatusFinished.String()})

}
