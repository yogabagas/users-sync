package main

import (
	"context"
	"log"
	"my-github/users-sync/config"
	"my-github/users-sync/service/auth"
	"my-github/users-sync/service/authz"
	"my-github/users-sync/shared"
	"sync"
)

func main() {
	config.InitDB()
	//service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	var wg sync.WaitGroup

	wg.Add(10)

	go func() {
		worker(ctx, 5000, 500, 1)
		wg.Done()
	}()

	go func() {
		worker(ctx, 5500, 500, 2)
		wg.Done()
	}()

	go func() {
		worker(ctx, 6000, 500, 3)
		wg.Done()
	}()

	go func() {
		worker(ctx, 6500, 500, 4)
		wg.Done()
	}()

	go func() {
		worker(ctx, 7000, 500, 5)
		wg.Done()
	}()

	go func() {
		worker(ctx, 7500, 500, 6)
		wg.Done()
	}()

	go func() {
		worker(ctx, 8000, 500, 7)
		wg.Done()
	}()

	go func() {
		worker(ctx, 8500, 500, 8)
		wg.Done()
	}()

	go func() {
		worker(ctx, 9000, 500, 9)
		wg.Done()
	}()

	go func() {
		worker(ctx, 9500, 500, 10)
		wg.Done()
	}()

	wg.Wait()
	log.Println("WORK FINISHED")
}

func worker(ctx context.Context, indexFrom, indexTo, no int) {
	res, err := authz.AuthzGetUserRoles(ctx, indexTo, indexFrom)
	if err != nil {
		log.Println("error when get user roles: ", err)
		return
	}

	for i, v := range res.Data.Users {
		log.Printf("WORKER %d DATA %d", no, i)
		log.Printf("DATA: %+v", v)

		err = auth.Process(ctx, v.UserID, v.Type, v.ClientRoleIDs)
		if err != nil {
			log.Println("error auth process: ", err)
		}
	}
}
