package service

import (
	"fmt"
	"net/http"
)

type Authz struct {
	UserUUID   string
	UserID     string
	ClientName string
	RoleName   string
}

const (
	endpointAuthz = "https://api.sicepat.io/v2/authz/management"
)

func AuthzProcess(req *Authz) {

	client := &http.Client{}

	http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users?userID=%s"))
}
