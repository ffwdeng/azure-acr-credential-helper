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
	"os"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/docker/docker-credential-helpers/credentials"
)

var (
	ErrNotImplemented      error = errors.New("not implemented")
	ErrUnsupportedRegistry error = errors.New("unsupported registry")
	ErrNoTenantIDClaim     error = errors.New("claim 'tid' not presend in token")
	ErrTenantIDNotString   error = errors.New("claim 'tid' is not a string in token")
	ErrAuthFailed          error = errors.New("authentication failed")
)

// This docker user should be used for ACR login when using JWT token as password
const DOCKER_USER string = "00000000-0000-0000-0000-000000000000"

type ACRHelper struct{}

func (az *ACRHelper) Get(serverURL string) (string, string, error) {
	logger := log.New(os.Stderr, "", 0)

	registry, err := extractRegistry(serverURL)
	if err != nil {
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	user, pass, err := getCredentials(registry)
	if err != nil {
		logger.Print(err)
		return "", "", credentials.NewErrCredentialsNotFound()
	}

	return user, pass, nil
}

func (az *ACRHelper) Add(creds *credentials.Credentials) error {
	return ErrNotImplemented
}

func (az *ACRHelper) Delete(serverURL string) error {
	return ErrNotImplemented
}

func (az *ACRHelper) List() (map[string]string, error) {
	return map[string]string{}, nil
}

// getCredentials Fetches credentials for an Azure container registry
func getCredentials(registryID string) (string, string, error) {
	// Use default azure credential provider to support both
	// logged in cli and managed identity.
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", "", fmt.Errorf("%s: %s", ErrAuthFailed.Error(), err)
	}

	// Fetch JWT token for authenticating with registry.
	token, err := cred.GetToken(
		context.Background(),
		policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}},
	)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to obtain token: %s", ErrAuthFailed.Error(), err)
	}

	// We need the tenant ID to authenticate with the registry.
	// We can pull the tenant ID out of the JWT token.
	// This assumes the registry is in the same tenant as our
	// authenticated credential.
	// TODO: support different tenant ID.
	tenantID, err := getTenantID(token.Token)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to extract tenant id from token: %s", ErrAuthFailed.Error(), err)
	}

	// Authenticate with azure container registry.
	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {registryID},
		"tenant":       {tenantID},
		"access_token": {token.Token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", registryID), formData)
	if err != nil {
		return "", "", fmt.Errorf("%s: %s", ErrAuthFailed.Error(), err)
	}

	// Decode response and return refresh token.
	var response map[string]interface{}
	err = json.NewDecoder(jsonResponse.Body).Decode(&response)
	if err != nil {
		return "", "", err
	}

	val, ok := response["refresh_token"]
	if !ok {
		return "", "", fmt.Errorf("%s: unable to get refresh token", ErrAuthFailed.Error())
	}

	rt, ok := val.(string)
	if !ok {
		return "", "", fmt.Errorf("%s: unable to cast refresh token to string", ErrAuthFailed.Error())
	}

	return DOCKER_USER, rt, nil
}

// getTenantID Fetches the Tenant ID out of JWT's claims
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
		return "", ErrNoTenantIDClaim
	}

	tid, ok := val.(string)
	if !ok {
		return "", ErrTenantIDNotString
	}

	return tid, nil
}

func extractRegistry(serverURL string) (string, error) {
	// registry must be *.azurecr.io for us to be able to authenticate
	regex := regexp.MustCompile(`[[a-zA-Z0-9\-]+\.azurecr\.io`)
	registry := regex.FindString(serverURL)

	if registry == "" {
		return "", ErrUnsupportedRegistry
	}

	return registry, nil
}
