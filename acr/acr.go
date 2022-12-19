package acr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/docker/docker-credential-helpers/credentials"
)

var notImplemented error = errors.New("not implemented")

// This docker user should be used for ACR login when using JWT token as password
const DOCKER_USER string = "00000000-0000-0000-0000-000000000000"

type ACRHelper struct{}

func (az *ACRHelper) Get(serverURL string) (string, string, error) {
	// registry must be *.azurecr.io for us to be able to authenticate
	regex := regexp.MustCompile(`[[a-zA-Z0-9\-]+\.azurecr\.io`)
	registry := regex.FindString(serverURL)

	if registry == "" {
		return "", "", fmt.Errorf("unsupported registry")
	}

	return getCredentials(registry)
}

func (az *ACRHelper) Add(creds *credentials.Credentials) error {
	return notImplemented
}

func (az *ACRHelper) Delete(serverURL string) error {
	return notImplemented
}

func (az *ACRHelper) List() (map[string]string, error) {
	return map[string]string{}, nil
}

func getCredentials(registryID string) (string, string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	token, err := cred.GetToken(
		context.Background(),
		policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to obtain token: %v", err)
	}

	tenantID, err := getTenantID(token.Token)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract tenant id from token: %v\n", err)
	}

	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {registryID},
		"tenant":       {tenantID},
		"access_token": {token.Token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", registryID), formData)
	if err != nil {
		return "", "", err
	}

	var response map[string]interface{}
	err = json.NewDecoder(jsonResponse.Body).Decode(&response)
	if err != nil {
		return "", "", err
	}

	val, ok := response["refresh_token"]
	if !ok {
		return "", "", fmt.Errorf("unable to get refresh token")
	}

	rt, ok := val.(string)
	if !ok {
		return "", "", fmt.Errorf("unable to cast refresh token to string")
	}

	return DOCKER_USER, rt, nil
}

func getTenantID(token string) (string, error) {
	parts := strings.Split(token, ".")
	claimPart := parts[1]

	claimJSON, err := base64.RawURLEncoding.DecodeString(claimPart)
	if err != nil {
		return "", err
	}

	claims := map[string]interface{}{}
	err = json.Unmarshal(claimJSON, &claims)
	if err != nil {
		return "", err
	}

	val, ok := claims["tid"]
	if !ok {
		return "", fmt.Errorf("claim 'tid' not present in token")
	}

	tid, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("claim 'tid' is not a string in token")
	}

	return tid, nil
}
