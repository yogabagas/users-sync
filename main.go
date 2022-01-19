package main

import (
	"context"
	"log"
	"my-github/users-sync/service"
)

func main() {

	userData, err := service.AuthzGetUserID(context.Background(), &service.Authz{
		UserID: "83233",
	})
	if err != nil {
		log.Println(err)
	}

	log.Println(&userData)

}
