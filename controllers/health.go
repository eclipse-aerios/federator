package controllers

import (
	"log"
	"net/http"
	"strings"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
	"github.com/eclipse-aerios/federator/services"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	orionSvc     services.OrionldSvc
	federatorSvc services.FederatorSvc
}

func (h *HealthController) Status(c *gin.Context) {
	// Local checks
	/*** 1. Check if Orion-LD is reachable and healthy */
	log.Println("Checking the health of Orion-LD...")
	isOrionHealthy, err := h.orionSvc.IsOrionHealthy()
	if !isOrionHealthy {
		returnUnhealthyStatus(c, config.UNHEALTHY_STATUS, "", "", "Orion-LD of the domain is unhealthy", err.Error())
		return
	}

	domain, err := h.orionSvc.GetLocalDomainEntity("simplified", "domainStatus", "")
	if err != nil {
		log.Println(err)
		returnUnhealthyStatus(c, config.HEALTHY_STATUS, "", "", "Cannot retrieve local domain", err.Error())
		return
	}
	domainStatus := strings.ReplaceAll(domain.DomainStatus, "urn:ngsi-ld:DomainStatus:", "")

	// TODO retrieve domains only if the local domain has a functional status
	// if domain.DomainStatus == config.FUNCTIONAL_DOMAIN_STATUS {
	// }

	domains, domainsCount, err := h.orionSvc.GetDomainEntities("simplified", true, "domainStatus", "", "", "")
	if err != nil {
		log.Println("Error when retrieving domains")
		returnUnhealthyStatus(c, config.HEALTHY_STATUS, domainStatus, "", "Cannot retrieve continuum domains", err.Error())
		return
	}
	var domainsNames string
	for i, d := range domains {
		domainsNames += models.GetNgsiLdEntityIdValue("Domain", d.Id)
		if i < len(domains)-1 {
			domainsNames += ","
		}
	}

	// External checks
	/*** 2. Check if the peer Federator API is reachable and healthy */
	if !config.IS_ENTRYPOINT {
		log.Println("Checking the health of the peer federator...")
		isPeerFederatorHealthy, _, err := h.federatorSvc.CheckFederatorHealth(config.PEER_FEDERATOR_URL)
		if !isPeerFederatorHealthy {
			returnUnhealthyStatus(c, config.HEALTHY_STATUS, domainStatus, config.UNHEALTHY_STATUS, "The peer federator is unhealty", err.Error())
			return
		}
	}

	apiHealth := &models.ApiHealth{
		Status:              config.HEALTHY_STATUS,
		OrionLdStatus:       config.HEALTHY_STATUS,
		Domain:              config.DOMAIN_NAME,
		DomainStatus:        domainStatus,
		PeerFederatorDomain: config.PeerFederatorDomain,
		PeerFederatorStatus: config.HEALTHY_STATUS,
		IsEntrypoint:        config.IS_ENTRYPOINT,
		FederatedDomains: models.FederatedDomains{
			Total: domainsCount,
			Names: domainsNames,
		},
		Message: "The aeriOS Federator is HEALTHY",
	}
	c.JSON(http.StatusOK, apiHealth)
}

func returnUnhealthyStatus(c *gin.Context, orionLdStatus string, domainStatus string, peerFederatorStatus string, message string, errorMessage string) {
	apiHealth := models.ApiHealth{
		Status:               config.UNHEALTHY_STATUS,
		OrionLdStatus:        orionLdStatus,
		Domain:               config.DOMAIN_NAME,
		DomainStatus:         domainStatus,
		PeerFederatorDomain:  config.PeerFederatorDomain,
		PeerFederatorStatus:  peerFederatorStatus,
		IsEntrypoint:         config.IS_ENTRYPOINT,
		Message:              message,
		DetailedErrorMessage: errorMessage,
	}
	c.JSON(http.StatusInternalServerError, apiHealth)
}
