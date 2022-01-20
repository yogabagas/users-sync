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
)

func main() {
	config.InitDB()
	service.Import()

	contextParent := context.Background()
	ctx := context.WithValue(contextParent, "token", shared.AuthToken)

	worker(ctx, 0, 5)
}

func worker(ctx context.Context, indexFrom, indexTo int) error {

	resp, err := repository.ReadFromLocalDB(ctx, int64(indexTo), int64(indexFrom))
	if err != nil {
		return err
	}

	for _, v := range resp {
		masterDataUsers, err := masterdata.SearchUserByNIK(ctx, v.NIK)
		if err != nil {
			repository.UpdateStatus(ctx, repository.LogData{
				NIK:         v.NIK,
				Status:      int(shared.StatusFailInMasterData),
				Description: shared.StatusFailInMasterData.String(),
			})
			return err
		}

		if masterDataUsers.ID > 0 {
			entityUsers, err := auth.Process(ctx, masterDataUsers.ID, masterDataUsers.NIK)
			if err != nil {
				repository.UpdateStatus(ctx, repository.LogData{
					NIK:         v.NIK,
					Status:      int(shared.StatusFailInAuth),
					Description: shared.StatusFailInAuth.String(),
				})
				return err
			}

			if entityUsers != nil {

				usersData, err := authz.AuthzGetUserID(ctx, &authz.Authz{
					UserID: fmt.Sprint(masterDataUsers.ID),
				})
				if err != nil {
					repository.UpdateStatus(ctx, repository.LogData{
						NIK:         v.NIK,
						Status:      int(shared.StatusFailInAuthz),
						Description: shared.StatusFailInAuthz.String(),
					})
					return err
				}

				if len(usersData.Data.Users) > 0 {

					roles := strings.Split(v.Role, ",")
					for _, role := range roles {
						log.Println(v.NIK, role)
						clientRoleData, err := authz.AuthzGetClientRoleID(ctx, &authz.Authz{
							ClientName: "hr",
							RoleName:   role,
						})
						if err != nil {
							repository.UpdateStatus(ctx, repository.LogData{
								NIK:         v.NIK,
								Status:      int(shared.StatusFailInAuthz),
								Description: shared.StatusFailInAuthz.String(),
							})
							return err
						}

						log.Println(clientRoleData.Data.ClientRoles)

						if err = authz.AuthzInsertUserRoles(ctx, &authz.Authz{
							RoleName: role,
						}, &authz.ClientRoleData{
							Data: clientRoleData.Data,
						}, &authz.UserData{
							Data: usersData.Data,
						}); err != nil {
							repository.UpdateStatus(ctx, repository.LogData{
								NIK:         v.NIK,
								Status:      int(shared.StatusFailInAuthz),
								Description: shared.StatusFailInAuthz.String(),
							})
							return err
						}
					}

				} else {

					err = authz.AuthzInsertUser(ctx, &authz.Authz{
						UserID: fmt.Sprint(masterDataUsers.ID),
					})
					if err != nil {
						repository.UpdateStatus(ctx, repository.LogData{
							NIK:         v.NIK,
							Status:      int(shared.StatusFailInAuthz),
							Description: shared.StatusFailInAuthz.String(),
						})
						return err
					}

					usersData, err := authz.AuthzGetUserID(ctx, &authz.Authz{
						UserID: fmt.Sprint(masterDataUsers.ID),
					})
					if err != nil {
						repository.UpdateStatus(ctx, repository.LogData{
							NIK:         v.NIK,
							Status:      int(shared.StatusFailInAuthz),
							Description: shared.StatusFailInAuthz.String(),
						})
						return err
					}

					if len(usersData.Data.Users) > 0 {
						roles := strings.Split(v.Role, ",")
						for _, role := range roles {

							clientRoleData, err := authz.AuthzGetClientRoleID(ctx, &authz.Authz{
								ClientName: "HR",
								RoleName:   role,
							})
							if err != nil {
								repository.UpdateStatus(ctx, repository.LogData{
									NIK:         v.NIK,
									Status:      int(shared.StatusFailInAuthz),
									Description: shared.StatusFailInAuthz.String(),
								})
								return err
							}

							if err = authz.AuthzInsertUserRoles(ctx, &authz.Authz{
								RoleName: role,
							}, &authz.ClientRoleData{
								Data: clientRoleData.Data,
							}, &authz.UserData{
								Data: usersData.Data,
							}); err != nil {
								repository.UpdateStatus(ctx, repository.LogData{
									NIK:         v.NIK,
									Status:      int(shared.StatusFailInAuthz),
									Description: shared.StatusFailInAuthz.String(),
								})
								return err
							}
						}
					}
				}
			}
		}
		repository.UpdateStatus(ctx, repository.LogData{
			NIK:         v.NIK,
			Status:      int(shared.StatusFinished),
			Description: shared.StatusFinished.String(),
		})
	}
	return nil
}
