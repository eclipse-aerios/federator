package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-aerios/federator/models"
	"github.com/joho/godotenv"
)

const (
	SERVICE_NAME             string = "aeriOS Federator"
	API_VERSION              string = "1.0.1"
	HEALTHY_STATUS           string = "HEALTHY"
	UNHEALTHY_STATUS         string = "UNHEALTHY"
	INITIAL_DOMAIN_STATUS    string = models.NGSILD_PREFIX + "DomainStatus:Preliminary"
	DELETED_DOMAIN_STATUS    string = models.NGSILD_PREFIX + "DomainStatus:Removed"
	DISABLED_DOMAIN_STATUS   string = models.NGSILD_PREFIX + "DomainStatus:Disabled"
	FUNCTIONAL_DOMAIN_STATUS string = models.NGSILD_PREFIX + "DomainStatus:Functional"
	REGISTRATIONS_PER_DOMAIN int    = 3
	REGISTRATIONS_PREFIX     string = "urn:aerios:federation"
)

var REGISTRATIONS_TYPES []string = []string{
	"organizations",
	"infrastructure",
	"services",
}

var APP_ENV string
var APP_PORT string
var IS_ENTRYPOINT bool
var DOMAIN_NAME string
var DOMAIN_DESCRIPTION string
var DOMAIN_PUBLIC_URL string
var DOMAIN_OWNER string
var DOMAIN_CB_URL string
var DOMAIN_CB_HEALTH_URL string
var PEER_FEDERATOR_URL string
var BROKER_ID string
var LOCAL_DOMAIN *models.NewDomain
var CB_HEALTH_CHECK_MODE string
var CB_TOKEN_MODE string
var TLS_CERTIFICATE_VALIDATION bool
var AERIOS_SHIM_URL string
var CB_OAUTH_CLIENT_ID string
var CB_OAUTH_CLIENT_SECRET string
var KEYCLOAK_URL string
var KEYCLOAK_REALM string
var DOMAIN_FEDERATOR_URL string
var Status string = HEALTHY_STATUS
var OrionToken *models.KeycloakAccessToken
var PeerFederatorDomain string

// TODO replace by an init function?
func LoadEnvVars() {
	var err error
	var isAppEnvPresent bool
	var isAppPortPresent bool

	APP_ENV, isAppEnvPresent = os.LookupEnv("APP_ENV")
	if !isAppEnvPresent || APP_ENV == "" {
		APP_ENV = "development"
	}
	log.Println("Federator environment mode: " + APP_ENV)
	// DEV Load environment variables from .env file -> change the "env" variable to use a specific file -> Uncomment
	if APP_ENV != "production" {
		const envsFolder = "test/"
		const envFile = ""
		log.Println("Not in production mode, so loading env vars from the local file " + envsFolder + envFile + ".env")
		if err := godotenv.Load(envsFolder + envFile + ".env"); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	APP_PORT, isAppPortPresent = os.LookupEnv("APP_PORT")
	if !isAppPortPresent || APP_PORT == "" {
		APP_PORT = "8050"
	}
	_, isEntrypointPresent := os.LookupEnv("IS_ENTRYPOINT")
	if !isEntrypointPresent {
		log.Println("IS_ENTRYPOINT env var not present, setting to false")
		IS_ENTRYPOINT = false
	} else {
		IS_ENTRYPOINT, err = strconv.ParseBool(os.Getenv("IS_ENTRYPOINT"))
		if err != nil {
			log.Panicln("Error loading the IS_ENTRYPOINT environment variable")
		}
	}
	if IS_ENTRYPOINT {
		log.Println("ENTRYPOINT MODE")
	}

	DOMAIN_NAME = strings.ReplaceAll(os.Getenv("DOMAIN_NAME"), " ", "")
	DOMAIN_DESCRIPTION = os.Getenv("DOMAIN_DESCRIPTION")
	DOMAIN_PUBLIC_URL = os.Getenv("DOMAIN_PUBLIC_URL")
	DOMAIN_OWNER = os.Getenv("DOMAIN_OWNER")
	DOMAIN_CB_URL = os.Getenv("DOMAIN_CB_URL")
	DOMAIN_CB_HEALTH_URL = os.Getenv("DOMAIN_CB_HEALTH_URL")
	DOMAIN_FEDERATOR_URL = os.Getenv("DOMAIN_FEDERATOR_URL")
	PEER_FEDERATOR_URL = os.Getenv("PEER_FEDERATOR_URL")
	CB_HEALTH_CHECK_MODE = os.Getenv("CB_HEALTH_CHECK_MODE")

	_, isCBTokenModePresent := os.LookupEnv("CB_TOKEN_MODE")
	if !isCBTokenModePresent {
		log.Println("CB_TOKEN_MODE env var not present, setting to shim")
		CB_TOKEN_MODE = "shim"
	} else {
		CB_TOKEN_MODE = os.Getenv("CB_TOKEN_MODE")
		if CB_TOKEN_MODE != "shim" && CB_TOKEN_MODE != "keycloak" {
			log.Println("CB_TOKEN_MODE has no valid value: " + CB_TOKEN_MODE + ", so setting to shim")
			CB_TOKEN_MODE = "shim"
		}
	}
	log.Println("Context Broker Authorization token mode: " + CB_TOKEN_MODE)

	AERIOS_SHIM_URL = os.Getenv("AERIOS_SHIM_URL")
	CB_OAUTH_CLIENT_ID = os.Getenv("CB_OAUTH_CLIENT_ID")
	CB_OAUTH_CLIENT_SECRET = os.Getenv("CB_OAUTH_CLIENT_SECRET")
	KEYCLOAK_URL = os.Getenv("KEYCLOAK_URL")
	KEYCLOAK_REALM = os.Getenv("KEYCLOAK_REALM")

	_, isTlsValPresent := os.LookupEnv("TLS_CERTIFICATE_VALIDATION")
	if !isTlsValPresent {
		log.Println("TLS_CERTIFICATE_VALIDATION env var not present, setting to false")
		TLS_CERTIFICATE_VALIDATION = false
	} else {
		TLS_CERTIFICATE_VALIDATION, err = strconv.ParseBool(os.Getenv("TLS_CERTIFICATE_VALIDATION"))
		if err != nil {
			log.Panicln("Error loading the TLS_CERTIFICATE_VALIDATION environment variable")
		}
	}

	OrionToken = &models.KeycloakAccessToken{
		AccessToken: "",
		ExpiresAt:   time.Now().Add(-1 * time.Minute), // Token is expired
	}

	fmt.Println("")

	// create here the localDomain variable as a constant
	LOCAL_DOMAIN = &models.NewDomain{
		Name:         DOMAIN_NAME,
		PublicUrl:    DOMAIN_PUBLIC_URL,
		IsEntrypoint: IS_ENTRYPOINT,
	}
}
