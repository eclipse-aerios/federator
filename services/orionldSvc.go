package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
)

type OrionldSvc struct {
	orionLdAuthSvc OrionLdAuthSvc
}

const CSR_PATH = "/ngsi-ld/v1/csourceRegistrations"
const ENTITIES_PATH = "/ngsi-ld/v1/entities"
const SOURCE_IDENTITY_PATH = "/ngsi-ld/v1/info/sourceIdentity"
const VERSION_PATH = "/version"

func (s *OrionldSvc) IsOrionHealthy() (bool, error) {
	if config.CB_HEALTH_CHECK_MODE == "endpoint" {
		log.Println("Performing an HTTP GET request to the /version endpoint...")
		fullURL := fmt.Sprintf("%s%s", config.DOMAIN_CB_URL, VERSION_PATH)
		res, err := http.Get(fullURL)
		if err != nil {
			log.Println("Error reaching the version endpoint")
			return false, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return false, errors.New(strconv.Itoa(res.StatusCode) + ": error reaching the version endpoint")
		}
		return true, nil
	} else {
		log.Println("Orion healthcheck in TCP socket mode")
		// Connect to the server
		conn, err := net.Dial("tcp", config.DOMAIN_CB_HEALTH_URL)
		if err != nil {
			config.Status = config.UNHEALTHY_STATUS
			if conn != nil {
				// Close the connection
				conn.Close()
			}
			return false, err
		}
		// Close the connection
		conn.Close()
		return true, nil

		// Send some data to the server (not needed)
		// _, err = conn.Write([]byte("The aeriOS federator is checking your health))
		// if err != nil {
		// 	fmt.Println(err)
		// 	config.Status = config.UNHEALTHY_STATUS
		// 	conn.Close()
		// 	returnUnhealthyStatus(c, "Cannot reach the Healthcheck URL of the Orion-LD of the domain")
		// 	return
		// }
	}

}

func (s *OrionldSvc) CreateDomainEntity() error {
	domainOwner, _ := models.NewMultipleRelationship(models.BuildNgsiLdEntityId("Organization", config.DOMAIN_OWNER))
	domain := &models.Domain{
		Id:           models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME),
		Type:         "Domain",
		Description:  config.DOMAIN_DESCRIPTION,
		PublicUrl:    config.DOMAIN_PUBLIC_URL,
		Owner:        domainOwner,
		IsEntrypoint: config.IS_ENTRYPOINT,
		DomainStatus: models.NewRelationship(config.FUNCTIONAL_DOMAIN_STATUS), // INITIAL_DOMAIN_STATUS
	}
	if config.DOMAIN_FEDERATOR_URL != "" {
		domain.FederatorUrl = config.DOMAIN_FEDERATOR_URL
	}

	bodyJson, err := json.Marshal(domain)
	if err != nil {
		log.Println("Failed to encode the domain in JSON")
		return err
	}

	fullURL := fmt.Sprintf("%s%s", config.DOMAIN_CB_URL, ENTITIES_PATH)
	res, err := http.Post(fullURL, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("Could not make POST request to the Orion-LD API")
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusConflict {
		return errors.New(strconv.Itoa(res.StatusCode) + ": the Domain entity is already present in the context broker")
	} else if res.StatusCode != http.StatusCreated {
		return errors.New(strconv.Itoa(res.StatusCode) + " :failed to create domain entity")
	}
	return nil
}

func (s *OrionldSvc) CreateOrganizationEntity() error {
	organization := &models.Organization{
		Id:   models.BuildNgsiLdEntityId("Organization", config.DOMAIN_OWNER),
		Type: "Organization",
		Name: config.DOMAIN_OWNER,
	}
	bodyJson, err := json.Marshal(organization)
	if err != nil {
		log.Println("Failed to encode the domain in JSON")
		return err
	}

	fullURL := fmt.Sprintf("%s%s", config.DOMAIN_CB_URL, ENTITIES_PATH)
	res, err := http.Post(fullURL, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("Could not make POST request to the Orion-LD API")
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusConflict {
		return errors.New(strconv.Itoa(res.StatusCode) + ": the Organization entity is already present in the context broker")
	} else if res.StatusCode != http.StatusCreated {
		return errors.New(strconv.Itoa(res.StatusCode) + " :failed to create organization entity")
	}
	return nil
}

func (s *OrionldSvc) CreateContextSourceRegistrations(registrations *[]models.ContextSourceRegistration) error {
	for _, v := range *registrations {
		bodyJson, err := json.Marshal(v)
		if err != nil {
			log.Println("Failed to encode the CSR in JSON")
			return err
		}
		fullURL := fmt.Sprintf("%s%s", config.DOMAIN_CB_URL, CSR_PATH)
		res, err := http.Post(fullURL, "application/json", bytes.NewBuffer(bodyJson))
		if err != nil {
			log.Println("Could not make POST request to the Orion-LD API")
			return err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusConflict {
			return errors.New(strconv.Itoa(res.StatusCode) + ": CSR is already present in the context broker")
		} else if res.StatusCode != http.StatusCreated {
			return errors.New(strconv.Itoa(res.StatusCode) + " :failed to create CSR")
		}
	}
	return nil
}

func (s *OrionldSvc) GenerateContextSourceRegistrations(newDomain *models.NewDomain) (registrations []models.ContextSourceRegistration) {
	infracrs := models.NewInfrastructureCSR(newDomain)
	organizationsCrs := models.NewOrganizationCSR(newDomain)
	servicesCsr := models.NewServicesCSR(newDomain)
	benchmarkCsr := models.NewBenchmarkCSR(newDomain)

	registrations = append(registrations, infracrs, organizationsCrs, servicesCsr, benchmarkCsr)
	return
}

func (s *OrionldSvc) GetDomainEntities(format string, count bool, attrs string, q string, options string, idPattern string) (domains []models.DomainSimplified, resultsCount int, err error) {
	queryParams := url.Values{}
	queryParams.Add("type", "Domain")
	queryParams.Add("format", format)
	queryParams.Add("count", strconv.FormatBool(count))
	if attrs != "" {
		queryParams.Add("attrs", attrs)
	}
	if q != "" {
		queryParams.Add("q", q)
	}
	if options != "" {
		queryParams.Add("options", options)
	}
	if idPattern != "" {
		queryParams.Add("idPattern", idPattern)
	}

	log.Println("Retrieving Domain entities from the continuum...")
	fullURL := fmt.Sprintf("%s?%s", config.DOMAIN_CB_URL+ENTITIES_PATH, queryParams.Encode())
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Set("aerOS", "true")

	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: s.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error retrieving Domain entities")
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return domains, 0, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving domains")
	}

	totalDomains := res.Header.Get("NGSILD-Results-Count")
	resultsCount, err = strconv.Atoi(totalDomains)
	if err != nil {
		resultsCount = 0
	}
	log.Println("Total number of domains: " + strconv.Itoa(resultsCount))

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	_ = json.Unmarshal(body, &domains)

	return domains, resultsCount, err
}

func (s *OrionldSvc) GetLocalDomainEntity(format string, attrs string, options string) (domain *models.DomainSimplified, err error) {
	queryParams := url.Values{}
	// queryParams.Add("type", "Domain")
	queryParams.Add("local", "true")
	queryParams.Add("format", format)
	if attrs != "" {
		queryParams.Add("attrs", attrs)
	}
	if options != "" {
		queryParams.Add("options", options)
	}

	log.Println("Retrieving the local Domain entity...")
	fullURL := fmt.Sprintf("%s%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME), queryParams.Encode())
	res, err := http.Get(fullURL)
	if err != nil {
		log.Println("Error retrieving local Domain entity")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return domain, errors.New("local domain entity not found")
	} else if res.StatusCode >= 400 {
		return domain, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving local Domain entity")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	_ = json.Unmarshal(body, &domain)

	return domain, err
}

func (s *OrionldSvc) ExistsLocalDomainEntity() (exists bool, err error) {
	queryParams := url.Values{}
	queryParams.Add("type", "Domain")
	// queryParams.Add("onlyIds", strconv.FormatBool(true))
	queryParams.Add("local", "true")

	log.Println("Retrieving the local Domain entity...")
	fullURL := fmt.Sprintf("%s%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME), queryParams.Encode())
	res, err := http.Get(fullURL)
	if err != nil {
		log.Println("Error retrieving local Domain entity")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, err
	} else if res.StatusCode >= 400 {
		return false, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving local Domain entity")
	} else {
		return true, nil
	}
}

func (s *OrionldSvc) ExistsDomainInTheContinuum(domain string) (exists bool, err error) {
	queryParams := url.Values{}
	queryParams.Add("type", "Domain")
	queryParams.Add("format", "simplified")
	queryParams.Add("attrs", "isEntrypoint")

	log.Println("Retrieving the Domain entity...")
	fullURL := fmt.Sprintf("%s%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Domain", domain), queryParams.Encode())
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Set("aerOS", "true")

	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: s.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error retrieving Domain entity")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, err
	} else if res.StatusCode >= 400 {
		return false, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving Domain entity")
	} else {
		return true, nil
	}
}

func (s *OrionldSvc) ExistsOrganizationInTheContinuum(organization string) (exists bool, err error) {
	queryParams := url.Values{}
	queryParams.Add("type", "Organization")
	queryParams.Add("format", "simplified")

	log.Println("Retrieving the Organization entity...")
	fullURL := fmt.Sprintf("%s%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Organization", organization), queryParams.Encode())
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Set("aerOS", "true")

	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: s.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error retrieving Organization entity")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, err
	} else if res.StatusCode >= 400 {
		return false, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving Organization entity")
	} else {
		return true, nil
	}
}

func (s *OrionldSvc) GetAeriosContextSourceRegistrations(csf string, count bool) (registrations []models.ContextSourceRegistration, err error) {
	queryParams := url.Values{}
	queryParams.Add("count", strconv.FormatBool(count))
	if csf == "" {
		queryParams.Add("csf", "aeriosDomainFederation==true")
	} else {
		queryParams.Add("csf", "aeriosDomainFederation==true&"+csf)
	}

	log.Println("Retrieving local aeriOS federation CSRs...")
	fullURL := fmt.Sprintf("%s?%s", config.DOMAIN_CB_URL+CSR_PATH, queryParams.Encode())
	res, err := http.Get(fullURL)
	if err != nil {
		log.Println("Error retrieving CSRs")
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return registrations, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving local CSRs")
	}

	totalCSRs := res.Header.Get("NGSILD-Results-Count")
	resultsCount, err := strconv.Atoi(totalCSRs)
	if err != nil {
		resultsCount = 0
	}
	log.Println("Total number of local CSRs: " + strconv.Itoa(resultsCount))

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	_ = json.Unmarshal(body, &registrations)

	return registrations, err
}

func (s *OrionldSvc) UpdateLocalDomainStatus(status string) (err error) {
	queryParams := url.Values{}
	queryParams.Add("local", "true")

	log.Println("Updating the local Domain entity...")

	body := models.NewRelationship(status)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to create request body")
		return
	}

	fullURL := fmt.Sprintf("%s%s/%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME), "attrs/domainStatus", queryParams.Encode())
	req, err := http.NewRequest(http.MethodPatch, fullURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error updating local Domain entity")
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return errors.New(strconv.Itoa(res.StatusCode) + ": domain entity not found")
	} else if res.StatusCode >= 400 {
		return errors.New(strconv.Itoa(res.StatusCode) + ": error updating local domain status")
	}
	return
}

func (s *OrionldSvc) DeleteLocalDomainEntity() (err error) {
	queryParams := url.Values{}
	queryParams.Add("local", "true")

	log.Println("Deleting local Domain entity ...")
	fullURL := fmt.Sprintf("%s%s/%s?%s", config.DOMAIN_CB_URL, ENTITIES_PATH, models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME), queryParams.Encode())

	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error deleting Domain entity")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return errors.New(strconv.Itoa(res.StatusCode) + ": Domain entity not found")
	} else if res.StatusCode >= 400 {
		return errors.New(strconv.Itoa(res.StatusCode) + ": error deleting local Domain entity")
	}
	return
}

func (s *OrionldSvc) DeleteContextSourceRegistration(regId string) (err error) {
	log.Println("Deleting local CSR " + regId + "...")
	fullURL := fmt.Sprintf("%s%s/%s", config.DOMAIN_CB_URL, CSR_PATH, regId)

	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error deleting CSR")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return errors.New(strconv.Itoa(res.StatusCode) + ": CSR not found")
	} else if res.StatusCode >= 400 {
		return errors.New(strconv.Itoa(res.StatusCode) + ": error deleting CSR")
	}
	return
}

func (s *OrionldSvc) DeleteAeriosContextSourceRegistrations() (err error) {
	log.Println("Retrieving local aeriOS CSRs...")
	localRegistrations, err := s.GetAeriosContextSourceRegistrations("", true)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Deleting local CSRs...")
	for _, reg := range localRegistrations {
		err = s.DeleteContextSourceRegistration(reg.Id)
		if err != nil {
			log.Println(err)
		}
	}

	return err
}

func (s *OrionldSvc) DeleteAeriosDomainContextSourceRegistrations(domain string) (err error) {
	log.Println("Retrieving local CSRs pointing to domain " + domain + "...")
	localRegistrations, err := s.GetAeriosContextSourceRegistrations("aeriosDomain==\""+domain+"\"", true)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Deleting local CSRs...")
	for _, reg := range localRegistrations {
		err = s.DeleteContextSourceRegistration(reg.Id)
		if err != nil {
			log.Println(err)
		}
	}

	return err
}

func (s *OrionldSvc) GetSourceIdentity() (sourceIdentity *models.SourceIdentity, err error) {
	log.Println("Retrieving the Source Identity of the broker...")
	fullURL := fmt.Sprintf("%s%s", config.DOMAIN_CB_URL, SOURCE_IDENTITY_PATH)
	res, err := http.Get(fullURL)
	if err != nil {
		log.Println("Error retrieving Source Identity of the broker")
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return sourceIdentity, errors.New(strconv.Itoa(res.StatusCode) + ": error retrieving Source Identity of the broker")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}
	err = json.Unmarshal(body, &sourceIdentity)
	if err != nil {
		log.Println("Error unmarshalling response body")
		return
	}

	return sourceIdentity, err
}
