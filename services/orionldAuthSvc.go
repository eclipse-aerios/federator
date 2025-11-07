package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
)

type OrionLdAuthSvc struct {
}

type Interceptor struct {
	core           http.RoundTripper
	orionLdAuthSvc OrionLdAuthSvc
}

const KEYCLOAK_REALM_PATH = "/auth/realms/"
const KEYCLOAK_TOKEN_PATH = "/protocol/openid-connect/token"
const KEYCLOAK_TOKEN_VALIDATION_PATH = "/protocol/openid-connect/userinfo"
const SHIM_TOKEN_PATH = "/token/cb"

func (s *OrionLdAuthSvc) GetTokenFromShim() (token string, err error) {
	log.Println("Retrieving the token from the aerios-shim module...")
	fullURL := fmt.Sprintf("%s%s", config.AERIOS_SHIM_URL, SHIM_TOKEN_PATH)
	res, err := http.Get(fullURL)
	if err != nil {
		log.Println("Error retrieving CB token")
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving the token from the aerios-shim")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	response := &models.AeriosShimToken{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Error unmarshalling response body")
		return
	}
	// config.OrionToken.AccessToken = response.Token

	return response.Token, err
}

func (s *OrionLdAuthSvc) GetTokenFromKeycloak() (token string, err error) {
	if !config.OrionToken.IsTokenExpired() {
		log.Println("The stored token is still valid")
		return config.OrionToken.AccessToken, nil
	}

	log.Println("Retrieving the token from Keycloak...")
	fullURL := fmt.Sprintf("%s%s%s%s", config.KEYCLOAK_URL, KEYCLOAK_REALM_PATH, config.KEYCLOAK_REALM, KEYCLOAK_TOKEN_PATH)
	payload := strings.NewReader("client_id=" + config.CB_OAUTH_CLIENT_ID + "&client_secret=" + config.CB_OAUTH_CLIENT_SECRET + "&grant_type=client_credentials")
	req, err := http.NewRequest(http.MethodPost, fullURL, payload)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error retrieving the CB token")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return "", errors.New(strconv.Itoa(res.StatusCode) + ": unauthorized, the Keycloak credentials are not valid")
	} else if res.StatusCode >= http.StatusBadRequest {
		return "", errors.New(strconv.Itoa(res.StatusCode) + ": not possible to retrieve the token from Keycloak")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	response := &models.KeycloakAccessToken{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Error unmarshalling response body")
		return
	}
	response.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	config.OrionToken = response

	return response.AccessToken, err
}

func (s *OrionLdAuthSvc) CheckTokenValidityInKeycloak(token string) (validToken bool, err error) {
	log.Println("Validating the token in Keycloak...")
	fullURL := fmt.Sprintf("%s%s%s%s", config.KEYCLOAK_URL, KEYCLOAK_REALM_PATH, config.KEYCLOAK_REALM, KEYCLOAK_TOKEN_VALIDATION_PATH)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error validating the token")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return false, errors.New(strconv.Itoa(res.StatusCode) + ": unauthorized, the token is not valid")
	} else {
		return true, nil
	}
}

func (s *OrionLdAuthSvc) GetAuthToken() (token string, err error) {
	if config.CB_TOKEN_MODE == "shim" {
		token, err = s.GetTokenFromShim()
	} else if config.CB_TOKEN_MODE == "keycloak" {
		token, err = s.GetTokenFromKeycloak()
	} else {
		return "", errors.New("the Authentication token retrieval mode has not been configured")
	}
	return
}

func (i *Interceptor) RoundTrip(req *http.Request) (*http.Response, error) {

	// modify before the request is sent
	accessToken, err := i.orionLdAuthSvc.GetAuthToken()
	if err != nil {
		return nil, err
	}
	log.Println("Token successfully obtained")
	// Set the Authorization header with the valid token
	// req.Header.Set("aerOS", "true")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// send the request using the DefaultTransport
	return i.core.RoundTrip(req)
}
