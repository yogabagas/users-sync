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
	"strings"
	"sync"
)

func main() {
	config.InitDB()
	service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	var wg sync.WaitGroup

	wg.Add(9)

	go func() {
		worker(ctx, 0, 300, 1)
		wg.Done()
	}()

	go func() {
		worker(ctx, 300, 300, 2)
		wg.Done()
	}()

	go func() {
		worker(ctx, 600, 300, 3)
		wg.Done()
	}()

	go func() {
		worker(ctx, 900, 300, 4)
		wg.Done()
	}()

	go func() {
		worker(ctx, 1200, 300, 5)
		wg.Done()
	}()

	go func() {
		worker(ctx, 1500, 300, 6)
		wg.Done()
	}()

	go func() {
		worker(ctx, 1800, 300, 7)
		wg.Done()
	}()

	go func() {
		worker(ctx, 2100, 300, 8)
		wg.Done()
	}()

	go func() {
		worker(ctx, 2400, 300, 9)
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
			entityUsers, err := auth.Process(ctx, masterDataUsers.ID, masterDataUsers.Username)
			if err != nil {
				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFailInAuth),
					Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuth.String(), err.Error()),
				})
				continue
			}

			if entityUsers != nil {

				usersData, err := authz.AuthzGetUserID(ctx, &authz.Authz{
					UserID: fmt.Sprint(masterDataUsers.ID),
				})
				if err != nil {
					log.Println(err.Error())
					repository.UpdateStatus(ctx, repository.LogData{
						NIK:         v.NIK,
						Status:      int(shared.StatusFailInAuthz),
						Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
					})
					continue
				}

				if len(usersData.Data.Users) > 0 {
					err = setUserRoles(ctx, v.Role, usersData.Data.Users[0].UserID, usersData.Data.Users[0].ID)
					if err != nil {
						log.Println(err)
						repository.UpdateStatus(ctx, repository.LogData{
							NIK:         v.NIK,
							Status:      int(shared.StatusFailInAuthz),
							Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
						})
						continue
					}

				} else {
					err = authz.AuthzInsertUser(ctx, &authz.Authz{
						UserID: fmt.Sprint(masterDataUsers.ID),
					})
					if err != nil {
						log.Println(err.Error())
						repository.UpdateStatus(ctx, repository.LogData{
							NIK:         v.NIK,
							Status:      int(shared.StatusFailInAuthz),
							Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
						})
						continue
					}

					usersData, err = authz.AuthzGetUserID(ctx, &authz.Authz{
						UserID: fmt.Sprint(masterDataUsers.ID),
					})
					if err != nil {
						log.Println(err.Error())
						repository.UpdateStatus(ctx, repository.LogData{
							NIK:         v.NIK,
							Status:      int(shared.StatusFailInAuthz),
							Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
						})
						continue
					}

					if len(usersData.Data.Users) > 0 {
						err = setUserRoles(ctx, v.Role, usersData.Data.Users[0].UserID, usersData.Data.Users[0].ID)
						if err != nil {
							log.Println(err.Error())
							repository.UpdateStatus(ctx, repository.LogData{
								NIK:         v.NIK,
								Status:      int(shared.StatusFailInAuthz),
								Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), err.Error()),
							})
							continue
						}
					}
				}

				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFinished),
					Description: shared.StatusFinished.String(),
				})
			} else {
				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFailInAuth),
					Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuthz.String(), "user in auth not found"),
				})
			}
		}
	}
}

func setUserRoles(ctx context.Context, roleData string, userID string, userUUID string) error {
	var clientRoleIDs []string
	var clientRole []authz.ClientRole

	roles := strings.Split(roleData, ",")
	for _, role := range roles {
		role = strings.TrimSpace(role)
		clientRoleData, err := authz.AuthzGetClientRoleID(ctx, &authz.Authz{
			ClientName: "hr",
			RoleName:   role,
		})
		if err != nil {
			return err
		}

		crLen := len(clientRoleData.Data.ClientRoles)
		if crLen > 0 {
			if crLen > 1 {
				clientRole = clientRoleData.Data.ClientRoles
			} else {
				clientRoleIDs = append(clientRoleIDs, clientRoleData.Data.ClientRoles[0].ID)
			}
		}

	}

	if len(clientRole) > 0 {
		for _, cr := range clientRole {
			if cr.Client.Name == "hr" {
				clientRoleIDs = append(clientRoleIDs, cr.ID)
			}
		}
	}

	if err := authz.AuthzInsertUserRoles(ctx, clientRoleIDs, userID); err != nil {
		return err
	}

	return nil
}
