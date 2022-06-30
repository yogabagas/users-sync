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
	"strings"
	"sync"
)

func main() {
	config.InitDB()
	//service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		worker(ctx, 0, 100, 1)
		wg.Done()
	}()
	//
	//go func() {
	//	worker(ctx, 60, 60, 2)
	//	wg.Done()
	//}()
	//
	//go func() {
	//	worker(ctx, 200, 100, 3)
	//	wg.Done()
	//}()
	//
	//go func() {
	//	worker(ctx, 300, 100, 4)
	//	wg.Done()
	//}()

	//go func() {
	//	worker(ctx, 400, 100, 5)
	//	wg.Done()
	//}()
	//
	//go func() {
	//	worker(ctx, 500, 100, 6)
	//	wg.Done()
	//}()
	//
	//go func() {
	//	worker(ctx, 600, 100, 7)
	//	wg.Done()
	//}()
	//
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

		if masterDataUsers.ID == 0 {
			repository.UpdateStatus(ctx, repository.LogData{
				NIK:    v.NIK,
				Status: int(shared.StatusFailInAuth),
				Description: fmt.Sprintf("%s: %s (NIK: %s USERNAME: %s)", shared.StatusFailInAuth.String(), "user in auth not found",
					v.NIK, masterDataUsers.Username),
			})
			continue
		}

		entityUsers, err := auth.Process(ctx, masterDataUsers.ID, masterDataUsers.NIK, masterDataUsers.Username)
		if err != nil {
			repository.UpdateStatus(ctx, repository.LogData{
				NIK:         v.NIK,
				Status:      int(shared.StatusFailInAuth),
				Description: fmt.Sprintf("%s: %s", shared.StatusFailInAuth.String(), err.Error()),
			})
			continue
		}

		if entityUsers == nil {
			repository.UpdateStatus(ctx, repository.LogData{
				NIK:    v.NIK,
				Status: int(shared.StatusFailInAuth),
				Description: fmt.Sprintf("%s: %s (NIK: %s USERNAME: %s)", shared.StatusFailInAuth.String(), "user in auth not found",
					v.NIK, masterDataUsers.Username),
			})
			continue
		}

		log.Printf("authz processing nik:%s userID:%d username:%s \n", masterDataUsers.NIK, masterDataUsers.ID, masterDataUsers.Username)
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
			err = setUserRoles(ctx, v.Role, usersData.Data.Users[0], true)
			if err != nil {
				log.Println(err.Error())
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
				err = setUserRoles(ctx, v.Role, usersData.Data.Users[0], false)
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
	}
}

func setUserRoles(ctx context.Context, roleData string, userData authz.User, userExist bool) error {
	var clientRoleIDs []string
	var clientRole []authz.ClientRole

	mClientRoles := make(map[string]bool)

	clients, err := authz.AuthzGetClient(ctx, "hr")
	if err != nil {
		return err
	}

	clientID := ""
	for _, v := range clients.Data.Clients {
		if v.Name == "hr" {
			clientID = v.ID
		}
	}

	roles := strings.Split(roleData, ",")
	for _, role := range roles {
		role = strings.TrimSpace(role)

		listRole, err := authz.AuthzGetRole(ctx, role)
		if err != nil {
			return err
		}

		roleID := ""
		for _, v := range listRole.Data.Roles {
			if v.RoleName == role {
				roleID = v.ID
			}
		}

		clientRoleData, err := authz.AuthzGetClientRoleID(ctx, &authz.Authz{
			ClientID: clientID,
			RoleID:   roleID,
		})
		if err != nil {
			return err
		}

		crLen := len(clientRoleData.Data.ClientRoles)
		if crLen > 0 {
			if crLen > 1 {
				clientRole = clientRoleData.Data.ClientRoles
			} else {
				mClientRoles[clientRoleData.Data.ClientRoles[0].ID] = true
			}
		}

	}

	if len(clientRole) > 0 {
		for _, cr := range clientRole {
			if cr.Client.Name == "hr" {
				mClientRoles[cr.ID] = true
			}
		}
	}

	if userExist {
		userRoles, err := authz.AuthzGetUserRoles(ctx, userData.ID)
		if err != nil {
			return err
		}

		for _, v := range userRoles.Data.Permissions {
			clientRoleData, err := authz.AuthzGetClientRoleID(ctx, &authz.Authz{
				ClientID: v.Client.ID,
				RoleID:   v.Role.ID,
			})
			if err != nil {
				return err
			}
			for _, v := range clientRoleData.Data.ClientRoles {
				if _, exists := mClientRoles[v.ID]; !exists {
					mClientRoles[v.ID] = true
				}
			}
		}
	}

	for v := range mClientRoles {
		clientRoleIDs = append(clientRoleIDs, v)
	}

	if err := authz.AuthzUpdateUserRoles(ctx, clientRoleIDs, userData.UserID); err != nil {
		return err
	}

	return nil
}
