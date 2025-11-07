package controllers

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
	"github.com/eclipse-aerios/federator/services"
	"github.com/gin-gonic/gin"
)

type DomainController struct {
	orionSvc     services.OrionldSvc
	federatorSvc services.FederatorSvc
}

func (d *DomainController) List(c *gin.Context) {
	domains, _, err := d.orionSvc.GetDomainEntities("simplified", true, "publicUrl,domainStatus,isEntrypoint,publicKey,owner", "", "", "")
	if err != nil {
		log.Println("Error when retrieving Domains")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve continuum domains"})
		return
	}
	c.JSON(http.StatusOK, domains)
}

func (d *DomainController) GetLocalDomain(c *gin.Context) {
	domain, err := d.orionSvc.GetLocalDomainEntity("simplified", "", "")
	if err != nil {
		log.Println("Error when retrieving Domains")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve the local domain"})
		return
	}
	c.JSON(http.StatusOK, domain)
}

func (d *DomainController) NewDomain(c *gin.Context) {
	// Check spread parameter
	spread, err := strconv.ParseBool(c.DefaultQuery("spread", "false"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Spread parameter must be boolean"})
		return
	}

	// Parse body
	newDomain := &models.NewDomain{}
	if err := c.ShouldBindJSON(&newDomain); err != nil {
		log.Println(err)
		if err == io.EOF {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The body of the request cannot be empty"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log.Println("New domain: " + newDomain.Name)

	// TODO The receiver federator can also check if this domain is already present in the continuum

	// Spread the new domain creation among the brokers of the continuum (it only must be done by the entrypoint or selected peer federator)
	if spread {
		log.Println("SPREADING MODE")
		// Check if the domain exists in the continuum
		log.Println("Checking the existence of the domain in the continuum...")
		domainExists, err := d.orionSvc.ExistsDomainInTheContinuum(newDomain.Name)
		if err != nil {
			log.Println("The existence of the new domain entity cannot be checked, so the spreading process cannot be started")
			log.Println(err)
			c.JSON(http.StatusInternalServerError, &models.NewDomainSpreadResponse{Message: "Cannot retrieve continuum domains"})
			return
		}
		if domainExists {
			log.Println("The domain already exists in the continuum")
			c.JSON(http.StatusBadRequest, &models.NewDomainSpreadResponse{Message: "The domain already exists in the continuum"})
			return
		} else {
			log.Println("The domain does not exist in the continuum")
		}
		// Create CSR in the local broker
		log.Println("Creating CSRs pointing to the new broker in the local broker...")
		newRegistrations := d.orionSvc.GenerateContextSourceRegistrations(newDomain)
		err = d.orionSvc.CreateContextSourceRegistrations(&newRegistrations)
		if err != nil {
			log.Println("Error when creating local CSRs")
			log.Println(err)
			if strings.Contains(err.Error(), "409") {
				c.JSON(http.StatusConflict, &models.NewDomainSpreadResponse{Message: "The domain has been already registered in the domain's context broker"})
			} else {
				c.JSON(http.StatusConflict, &models.NewDomainSpreadResponse{Message: "Cannot create CSRs in the domain's context broker"})
			}
			return
		}

		localDomainRegistrations := d.orionSvc.GenerateContextSourceRegistrations(config.LOCAL_DOMAIN)

		// Retrieve the filtered registrations (only aeriOS related and exclude the new broker itself) present in the local broker
		localRegistrations, err := d.orionSvc.GetAeriosContextSourceRegistrations("aeriosDomain!=\""+newDomain.Name+"\"", true)
		if err != nil {
			log.Println("Error when retrieving local CSRs")
			c.JSON(http.StatusInternalServerError, &models.NewDomainSpreadResponse{Message: "Cannot retrieve local CSRs"})
			return
		}

		// Add the registrations pointing to the local domain
		localRegistrations = append(localRegistrations, localDomainRegistrations...)

		// Get domains
		idPattern := "^(?!.*(" + models.BuildNgsiLdEntityId("Domain", newDomain.Name) + "|" + models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME) + ")).*$"
		// FIXME only functional domains
		// domainsQuery := "domainStatus==\"" + config.FUNCTIONAL_DOMAIN_STATUS + "\""
		domainsQuery := ""
		domains, _, err := d.orionSvc.GetDomainEntities("simplified", true, "publicUrl,domainStatus,isEntrypoint,federatorUrl", domainsQuery, "", idPattern)
		if err != nil {
			log.Println("Error when retrieving Domains")
			log.Println(err)
			c.JSON(http.StatusInternalServerError, &models.NewDomainSpreadResponse{Message: "Cannot retrieve continuum domains"})
			return
		}

		log.Println("Spreading the new domain...")
		// Send new Domain requests (spread=false) to notify the other brokers
		failedDomains := make([]string, 0)
		for _, domain := range domains {
			log.Println("Sending the new domain creation to domain -> " + domain.Id)
			log.Println("POST request to " + domain.FederatorUrl + " pointing to domain " + domain.Id)

			// Check if federatorUrl in domain entity, if not, use publicUrl + /federator
			var federatorUrl string
			if domain.FederatorUrl == "" {
				federatorUrl = domain.PublicUrl + "/federator"
			} else {
				federatorUrl = domain.FederatorUrl
			}
			// Notify the new domain addition to the domain federator and check the result
			_, err = d.federatorSvc.NotifyNewDomain(newDomain, federatorUrl)
			if err != nil {
				failedDomains = append(failedDomains, domain.Id)
			}
		}
		if len(domains) == 0 {
			log.Println("No domains to spread the new domain creation")
		}

		// Create and send response
		// TODO return also succeedDomains?
		response := &models.NewDomainSpreadResponse{
			NewRegistrations:       newRegistrations,
			Domains:                domains,
			NewDomainRegistrations: localRegistrations,
			FailedDomains:          failedDomains,
			Message:                "Spreading operation completed",
		}

		if len(failedDomains) > 0 {
			response.Message = "Spreading operation completed, but the domain addition has failed in some domains"
			c.JSON(http.StatusMultiStatus, response)
		} else {
			c.JSON(http.StatusCreated, response)
		}
	} else {
		log.Println("NO SPREADING MODE")
		// Create CSR in the local broker
		log.Println("Creating CSRs pointing to the new broker in the local broker...")
		newRegistrations := d.orionSvc.GenerateContextSourceRegistrations(newDomain)
		err = d.orionSvc.CreateContextSourceRegistrations(&newRegistrations)
		if err != nil {
			log.Println("Error when creating local CSRs")
			log.Println(err)
			if strings.Contains(err.Error(), "409") {
				c.JSON(http.StatusConflict, &models.NewDomainSpreadResponse{Message: "The domain has been already registered in the domain's context broker"})
			} else {
				c.JSON(http.StatusConflict, &models.NewDomainSpreadResponse{Message: "Cannot create CSRs in the domain's context broker"})
			}
			return
		}
		response := models.NewDomainSpreadResponse{
			NewRegistrations: newRegistrations,
			Message:          "New CSRs pointing to the domain '" + newDomain.Name + "' created in the domain's broker",
		}
		log.Println("New CSRs pointing to the domain '" + newDomain.Name + "' created in the domain's broker")
		c.JSON(http.StatusCreated, response)
	}
}

// ENABLED ONLY IN ENTRYPOINT (DeleteOtherDomain)
func (d *DomainController) SpreadDomainDeletion(c *gin.Context) {
	// TODO implement
	domain := c.Param("domainName")
	c.JSON(http.StatusInternalServerError, &models.NewDomainSpreadResponse{Message: "The deletion of Domain " + domain + "has been successfully spread"})
}

func (d *DomainController) Delete(c *gin.Context) {
	domain := c.Param("domainName")
	// TODO check if local domain and return a 400/403? if it hasn't been sent by the entrypoint/deleted domain itself (if yes, handle - send some code?)
	// TODO check if domain exists and return a 404
	// Delete CSR pointing to the deleted domain in the local broker
	err := d.orionSvc.DeleteAeriosDomainContextSourceRegistrations(domain)
	if err != nil {
		log.Println("Cannot delete local CSRs")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot delete local CSRs"})
		return
	}

	if domain == config.PeerFederatorDomain {
		// Select another peer federator -> default entrypoint domain?
		log.Println("Deleting the domain of the peer federator, so a new peer federator must be configured...")
		domainsQuery := "isEntrypoint==true"
		domains, _, err := d.orionSvc.GetDomainEntities("simplified", true, "publicUrl,domainStatus,isEntrypoint,federatorUrl", domainsQuery, "", "")
		if err != nil {
			log.Println("Error when retrieving Domains")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve continuum domains"})
			return
		}
		// only 1 entrypoint domain is possible
		if len(domains) != 1 {
			log.Println("Error when retrieving Domains")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve continuum domains"})
			return
		}
		if domains[0].FederatorUrl == "" {
			config.PEER_FEDERATOR_URL = domains[0].PublicUrl + "/federator"
		} else {
			config.PEER_FEDERATOR_URL = domains[0].FederatorUrl
		}
		config.PeerFederatorDomain = strings.ReplaceAll(domains[0].Id, "urn:ngsi-ld:Domain:", "")
		log.Println("The new peer federator is the entrypoint domain federator -> " + config.PeerFederatorDomain)
		// TODO this works, but what about if the federator dies? the former value from the env var will be used... -> need of an aux db
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain " + domain + " successfully deleted"})
}

func (d *DomainController) DeleteLocalDomain(c *gin.Context) {
	// TODO solve in the future -> move the entrypoint domain?
	if config.IS_ENTRYPOINT {
		log.Println("The entrypoint domain cannot be deleted")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "The entrypoint domain cannot be deleted"})
		return
	}

	// Check if status is Removed
	localDomain, err := d.orionSvc.GetLocalDomainEntity("simplified", "domainStatus", "")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve domain status"})
		return
	}
	if localDomain.DomainStatus == config.DELETED_DOMAIN_STATUS {
		c.JSON(http.StatusBadRequest, gin.H{"message": "The domain has already been removed"})
		return
	}

	// Update Domain status to Removed
	err = d.orionSvc.UpdateLocalDomainStatus(config.DELETED_DOMAIN_STATUS)
	if err != nil {
		log.Println("Cannot update domain status to Removed")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot update domain status to Removed"})
		return
	}

	// Spread the domain deletion among the brokers of the continuum (it only must be done by the entrypoint or selected peer federator)
	// FIXME only functional domains -> domainsQuery := "domainStatus==\"" + config.FUNCTIONAL_DOMAIN_STATUS + "\""
	domainsQuery := ""
	domains, _, err := d.orionSvc.GetDomainEntities("simplified", true, "publicUrl,domainStatus,federatorUrl", domainsQuery, "", "")
	if err != nil {
		log.Println("Error when retrieving Domains")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot retrieve continuum domains"})
		return
	}

	// Send delete Domain requests to notify the other brokers
	failedDomains := make([]string, 0)
	if len(domains) == 0 {
		log.Println("No domains to spread the domain deletion")
	}
	for _, domain := range domains {
		if domain.Id == models.BuildNgsiLdEntityId("Domain", config.DOMAIN_NAME) {
			continue
		}
		log.Println("Sending the domain deletion to domain " + domain.Id)
		log.Println("DELETE request to " + domain.FederatorUrl + " pointing to domain " + domain.Id)

		// If Federator URL is not included in the Domain entity, use publicUrl + "/federator"
		var federatorUrl string
		if domain.FederatorUrl == "" {
			federatorUrl = domain.PublicUrl + "/federator"
		} else {
			federatorUrl = domain.FederatorUrl
		}
		err = d.federatorSvc.NotifyDeletedDomain(config.DOMAIN_NAME, federatorUrl)
		if err != nil {
			log.Println(err)
			log.Println("Cannot contant with the domain to spread the domain deletion")
			failedDomains = append(failedDomains, domain.Id)
		}
	}

	// Delete CSR in the local broker
	err = d.orionSvc.DeleteAeriosContextSourceRegistrations()
	if err != nil {
		log.Println("Cannot delete local CSRs")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot delete local CSRs"})
		return
	}

	response := &models.DeleteDomainSpreadResponse{
		FailedDomains: failedDomains,
		Message:       "",
	}

	if len(failedDomains) > 0 {
		response.Message = "Local domain " + config.DOMAIN_NAME + " successfully deleted, but the deletion has failed in some domains"
		c.JSON(http.StatusMultiStatus, response)
	} else {
		response.Message = "Local domain " + config.DOMAIN_NAME + " successfully deleted"
		c.JSON(http.StatusCreated, response)
	}
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
