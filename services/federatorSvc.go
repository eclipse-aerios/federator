package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
)

type FederatorSvc struct {
	orionLdAuthSvc OrionLdAuthSvc
}

const DOMAINS_PATH = "/v1/domains"
const HEALTH_PATH = "/health"

// Notifies a new domain creation to another federator, acting as the PEER domain
// FIXME can this function just return the error? The response was only used for testing purposes...
func (f *FederatorSvc) NotifyNewDomain(newDomain *models.NewDomain, federatorUrl string) (response *models.NewDomainSpreadResponse, err error) {
	queryParams := url.Values{}
	queryParams.Add("spread", "false")
	fullURL := fmt.Sprintf("%s%s?%s", federatorUrl, DOMAINS_PATH, queryParams.Encode())

	bodyJson, err := json.Marshal(newDomain)
	if err != nil {
		log.Println("Failed to encode the domain in JSON")
		return
	}
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: f.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Could not make POST request to the Federator API")
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	response = &models.NewDomainSpreadResponse{}
	json.Unmarshal(body, response)
	if err != nil {
		log.Println("Error unmarshalling response body")
		return
	}

	if res.StatusCode != http.StatusCreated {
		log.Println("Failed to notify the creation of the domain")
		resJson, _ := json.MarshalIndent(response, "", " ")
		log.Println(string(resJson))
		err = errors.New(strconv.Itoa(res.StatusCode) + ": " + response.Message)
		return
	}
	return
}

// Spreads the LOCAL new domain creation -> only in initialization, sends the request to the PEER domain
func (f *FederatorSvc) SpreadNewLocalDomain() (response *models.NewDomainSpreadResponse, err error) {
	queryParams := url.Values{}
	queryParams.Add("spread", "true")
	fullURL := fmt.Sprintf("%s%s?%s", config.PEER_FEDERATOR_URL, DOMAINS_PATH, queryParams.Encode())

	bodyJson, err := json.Marshal(config.LOCAL_DOMAIN)
	if err != nil {
		log.Println("Failed to encode the domain in JSON")
		return
	}
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: f.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Could not make POST request to the Federator API")
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	response = &models.NewDomainSpreadResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		log.Println("Error unmarshalling response body")
		return
	}

	if res.StatusCode == http.StatusMultiStatus {
		log.Println(response.Message)
		log.Println(strings.Join(response.FailedDomains, ", "))
	} else if res.StatusCode != http.StatusCreated {
		log.Println("Failed to spread the creation of the domain")
		resJson, _ := json.MarshalIndent(response, "", " ")
		log.Println(string(resJson))
		err = errors.New(strconv.Itoa(res.StatusCode) + ": " + response.Message)
		return
	}
	return
}

// Notifies a domain deletion to another federator, acting as the PEER domain
func (f *FederatorSvc) NotifyDeletedDomain(domainId string, federatorUrl string) (err error) {
	fullURL := fmt.Sprintf("%s%s/%s", federatorUrl, DOMAINS_PATH, domainId)
	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return
	}
	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: f.orionLdAuthSvc,
		},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error deleting local Domain entity")
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(res.StatusCode) + ": failed to spread the deletion of the domain")
	}
	return
}

func (f *FederatorSvc) CheckFederatorHealth(url string) (bool, string, error) {
	log.Println("Checking the health of another Federator...")
	// build request
	fullURL := fmt.Sprintf("%s%s", url, HEALTH_PATH)
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Println("HTTP client: could not create request")
		return false, "", err
	}
	// use token interceptor
	client := &http.Client{
		Transport: &Interceptor{
			core:           http.DefaultTransport,
			orionLdAuthSvc: f.orionLdAuthSvc,
		},
	}
	// send the request
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error retrieving health info")
		return false, "", err
	}
	defer res.Body.Close()
	// check status
	if res.StatusCode == http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return false, "", err
		}
		apiHealth := &models.ApiHealth{}
		err = json.Unmarshal(body, apiHealth)
		if err != nil {
			return false, "", err
		}
		return true, apiHealth.Domain, err
	} else if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		// The token is not valid, unauthorized request to the peer federator
		log.Println("The token is not valid, unauthorized request sent to the peer federator (" + strconv.Itoa(res.StatusCode) + ")")
		return false, "", err
	} else {
		return false, "", err
	}
}
