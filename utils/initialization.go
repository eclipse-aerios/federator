package utils

import (
	"errors"
	"log"
	"strconv"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/services"
)

type Initialization struct {
	orionldSvc   services.OrionldSvc
	federatorSvc services.FederatorSvc
}

func (i *Initialization) InitializeFederator() (err error) {
	log.Println("Initializing the aeriOS Federator...")

	// Check Orion health
	isOrionHealthy, err := i.orionldSvc.IsOrionHealthy()
	if err != nil {
		log.Println(err)
	}
	if !isOrionHealthy {
		log.Panicln("Orion-LD instance of the domain is unhealthy, so the aeriOS Federator cannot be started.")
	}

	// Get information of the local context broker
	log.Println("Retrieving configuration of the Domain Context Broker (NGSI-LD Source Identity)...")
	brokerInfo, err := i.orionldSvc.GetSourceIdentity()
	if err != nil {
		return err
	}
	log.Println("CB Context Source Alias: " + brokerInfo.ContextSourceAlias)
	config.BROKER_ID = brokerInfo.ContextSourceAlias
	config.LOCAL_DOMAIN.BrokerId = brokerInfo.ContextSourceAlias
	// Do we need to check the status of the domain?
	noNewDomain, err := i.orionldSvc.ExistsLocalDomainEntity()
	if err != nil {
		log.Println("The existence of the local domain entity cannot be checked, so the aeriOS Federator cannot be started.")
		return err
	}
	// TODO check if the domain exists in the continuum -> panic

	// Check peer federator health
	log.Println("Checking the health of the peer federator...")
	if config.IS_ENTRYPOINT {
		// TODO check in the future
		log.Println("The entrypoint domain doesn't need a peer federator right now...")
	} else {
		isPeerFederatorHealthy, peerFederatorDomain, err := i.federatorSvc.CheckFederatorHealth(config.PEER_FEDERATOR_URL)
		if err != nil {
			log.Println("Impossible to check the peer federator health")
			return err
		} else if !isPeerFederatorHealthy {
			return errors.New("the peer federator is unhealthy")
		}
		// Set the domain of the peer federator
		config.PeerFederatorDomain = peerFederatorDomain
		log.Println("The peer federator belongs to the Domain " + config.PeerFederatorDomain)
	}

	if noNewDomain {
		log.Println("The Domain is already present in the Orion-LD of the Domain. This is not a new domain")
	} else {
		log.Println("The Domain is not present yet in the Orion-LD of the Domain. NEW DOMAIN ADDITION TO THE CONTINUUM")

		// Create Domain entity in Orion
		log.Println("Creating the Domain entity in Orion-LD")
		err := i.orionldSvc.CreateDomainEntity()
		if err != nil {
			return err
		} else {
			log.Println("Domain entity created")
		}

		// Check if isEntrypoint to not start the spreading process
		if !config.IS_ENTRYPOINT {
			log.Println("Spreading the creation of the new domain across the continuum...")

			// Spread this new domain creation to the Federator of the entrypoint domain (or other peer) -> SPREADING PROCESS
			spreadResponse, err := i.federatorSvc.SpreadNewLocalDomain()
			if err != nil {
				log.Println("Cannot contant with the peer domain to spread the new domain creation")
				// TODO implement a logic here to handle this error...
				// take some actions on the created domain entity -> delete it or mark with a new status (federationfailed?)
				deleteEntityError := i.orionldSvc.DeleteLocalDomainEntity()
				if deleteEntityError != nil {
					log.Println(deleteEntityError)
				}
				return err
			} else {
				log.Println("The creation of the new domain has been successfully spread")
				// fmt.Printf("%+v\n", *spreadResponse)

				// CSRs from the peer federator are returned as response, so create them the local broker
				log.Println("Creating CSRs pointing to the other brokers of the continuum")
				i.orionldSvc.CreateContextSourceRegistrations(&spreadResponse.NewDomainRegistrations)

				log.Println("Total number of domains (excluding the peer federator domain): " + strconv.Itoa(len(spreadResponse.Domains)))
				for i := 0; i < len(spreadResponse.Domains); i++ {
					log.Println(spreadResponse.Domains[i].Id + " " + spreadResponse.Domains[i].Description)
				}
				log.Println("Total number of FAILED domains: " + strconv.Itoa(len(spreadResponse.FailedDomains)))
				if len(spreadResponse.FailedDomains) > 0 {
					for _, d := range spreadResponse.FailedDomains {
						log.Println(d)
					}
				}
			}
		} else {
			log.Println("This Federator belongs to the Entrypoint Domain")
		}
	}

	// Create the Organization entity of the Domain owner in the continuum
	log.Println("Checking the existence of the Organization entity of the Domain owner in the continuum...")
	noNewOrganization, orgErr := i.orionldSvc.ExistsOrganizationInTheContinuum(config.DOMAIN_OWNER)
	if orgErr != nil {
		log.Println("The existence of the organization entity cannot be checked, so creating it locally...")
	}
	if !noNewOrganization {
		log.Println("The Organization entity is not present in the continuum, so creating it...")
		orgErr = i.orionldSvc.CreateOrganizationEntity()
		if orgErr != nil {
			log.Println("Error creating the Organization entity in the continuum: " + orgErr.Error())
		} else {
			log.Println("Organization entity created successfully")
		}
	}

	return err
}
