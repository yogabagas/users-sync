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
	"sync"
)

func main() {
	config.InitDB()
	//service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	var wg sync.WaitGroup

	wg.Add(5)

	go func() {
		worker(ctx, 0, 100, 1)
		wg.Done()
	}()

	go func() {
		worker(ctx, 100, 100, 2)
		wg.Done()
	}()

	go func() {
		worker(ctx, 200, 100, 3)
		wg.Done()
	}()

	go func() {
		worker(ctx, 300, 100, 4)
		wg.Done()
	}()

	go func() {
		worker(ctx, 400, 100, 5)
		wg.Done()
	}()

	wg.Wait()
	log.Println("WORK FINISHED")
}

func worker(ctx context.Context, indexFrom, indexTo, no int) {
	resp, err := repository.ReadFromLocalDB(ctx, int64(indexTo), int64(indexFrom))
	if err != nil {
		log.Println(err.Error())
		return
	}

	for index, v := range resp {
		log.Printf("WORKER %d DATA %d", no, index)
		masterDataUsers, err := masterdata.SearchUserByNIK(ctx, v.NIK)
		if err != nil {
			log.Println(err.Error())
			repository.UpdateStatus(ctx, repository.LogData{
				NIK:         v.NIK,
				Status:      int(shared.StatusFailInMasterData),
				Description: fmt.Sprintf("%s: %s", shared.StatusFailInMasterData.String(), err.Error()),
			})
			continue
		}

		if masterDataUsers.ID > 0 {
			entityUsers, err := auth.Process(ctx, masterDataUsers.ID, masterDataUsers.NIK, masterDataUsers.Username)
			if err != nil {
				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFailInAuth),
					Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuth.String(), err.Error()),
				})
				continue
			}

			if entityUsers != nil {
				log.Printf("authz processing nik:%s userID:%d username:%s \n", masterDataUsers.NIK, masterDataUsers.ID, masterDataUsers.Username)
				err = authz.AuthzInsertUserRoles(ctx, fmt.Sprint(masterDataUsers.ID))
				if err != nil {
					log.Println(err.Error())
					repository.UpdateStatus(ctx, repository.LogData{
						NIK:         v.NIK,
						Status:      int(shared.StatusFailInAuthz),
						Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
					})
					continue
				}

				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFinished),
					Description: shared.StatusFinished.String(),
				})
			} else {
				repository.UpdateStatus(ctx, repository.LogData{
					NIK:    v.NIK,
					Status: int(shared.StatusFailInAuth),
					Description: fmt.Sprintf("%s: %s (NIK: %s USERNAME: %s)", shared.StatusFailInAuth.String(), "user in auth not found",
						v.NIK, masterDataUsers.Username),
				})
			}
		}
	}
}
