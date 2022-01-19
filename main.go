package main

import (
	"context"
	"fmt"
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

	fmt.Println("TEST")

	log.Println(&userData)

}
