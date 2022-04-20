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

//func worker(ctx context.Context, indexFrom, indexTo, no int) {
//	resp, err := repository.ReadFromLocalDB(ctx, int64(indexTo), int64(indexFrom))
//	if err != nil {
//		log.Println(err.Error())
//		return
//	}
//
//	for index, v := range resp {
//		log.Printf("WORKER %d DATA %d", no, index)
//		masterDataUsers, err := masterdata.SearchUserByNIK(ctx, v.NIK)
//		if err != nil {
//			log.Println(err.Error())
//			repository.UpdateStatus(ctx, repository.LogData{
//				NIK:         v.NIK,
//				Status:      int(shared.StatusFailInMasterData),
//				Description: fmt.Sprintf("%s: %s", shared.StatusFailInMasterData.String(), err.Error()),
//			})
//			continue
//		}
//
//		if masterDataUsers.ID > 0 {
//			entityUsers, err := auth.Process(ctx, masterDataUsers.ID, masterDataUsers.NIK, masterDataUsers.Username)
//			if err != nil {
//				repository.UpdateStatus(ctx, repository.LogData{
//					NIK:         v.NIK,
//					Status:      int(shared.StatusFailInAuth),
//					Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuth.String(), err.Error()),
//				})
//				continue
//			}
//
//			if entityUsers != nil {
//				log.Printf("authz processing nik:%s userID:%d username:%s \n", masterDataUsers.NIK, masterDataUsers.ID, masterDataUsers.Username)
//				err = authz.AuthzInsertUserRoles(ctx, fmt.Sprint(masterDataUsers.ID))
//				if err != nil {
//					log.Println(err.Error())
//					repository.UpdateStatus(ctx, repository.LogData{
//						NIK:         v.NIK,
//						Status:      int(shared.StatusFailInAuthz),
//						Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
//					})
//					continue
//				}
//
//				repository.UpdateStatus(ctx, repository.LogData{
//					NIK:         v.NIK,
//					Status:      int(shared.StatusFinished),
//					Description: shared.StatusFinished.String(),
//				})
//			} else {
//				repository.UpdateStatus(ctx, repository.LogData{
//					NIK:    v.NIK,
//					Status: int(shared.StatusFailInAuth),
//					Description: fmt.Sprintf("%s: %s (NIK: %s USERNAME: %s)", shared.StatusFailInAuth.String(), "user in auth not found",
//						v.NIK, masterDataUsers.Username),
//				})
//			}
//		}
//	}
//}

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
