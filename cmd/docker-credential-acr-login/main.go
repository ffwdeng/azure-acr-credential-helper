package main

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/ffwdeng/azure-msi-docker-credential-helper/acr"
)

func main() {
	credentials.Serve(&acr.ACRHelper{})
}
