package models

import "strings"

type ContextSourceRegistration struct {
	Id                     string        `json:"id"`
	Type                   string        `json:"type"`
	Information            []information `json:"information"`
	ContextSourceInfo      []KeyValue    `json:"contextSourceInfo"`
	Mode                   string        `json:"mode"`
	HostAlias              string        `json:"hostAlias"`
	Operations             []string      `json:"operations"`
	Endpoint               string        `json:"endpoint"`
	Management             CSRManagement `json:"management"`
	AeriosDomain           string        `json:"aeriosDomain"`
	AeriosDomainFederation bool          `json:"aeriosDomainFederation"`
}

type CSRManagement struct {
	LocalOnly bool `json:"localOnly"`
}

type information struct {
	Entities []informationEntities `json:"entities"`
}

type informationEntities struct {
	Id   string `json:"id,omitempty"`
	Type string `json:"type"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewInfrastructureCSR(newDomain *NewDomain) ContextSourceRegistration {
	return ContextSourceRegistration{
		Id:   "urn:aerios:federation:" + strings.ToLower(newDomain.Name) + ":infrastructure",
		Type: "ContextSourceRegistration",
		Mode: "inclusive",
		Information: []information{
			{
				Entities: []informationEntities{
					{
						Type: "Domain",
					},
					{
						Type: "LowLevelOrchestrator",
					},
					{
						Type: "InfrastructureElement",
					},
				},
			},
		},
		ContextSourceInfo: []KeyValue{
			{
				Key:   "Authorization",
				Value: NGSILD_PREFIX + "request",
			},
		},
		Operations: []string{
			"retrieveOps",
		},
		HostAlias: newDomain.BrokerId,
		Endpoint:  newDomain.PublicUrl + "/orionld", // Add /orionld to be aligned with current KrakenD config
		// Endpoint: newDomain.PublicUrl,
		Management: CSRManagement{
			LocalOnly: true,
		},
		AeriosDomain:           newDomain.Name,
		AeriosDomainFederation: true,
	}
}

func NewServicesCSR(newDomain *NewDomain) ContextSourceRegistration {
	return ContextSourceRegistration{
		Id:   "urn:aerios:federation:" + strings.ToLower(newDomain.Name) + ":services",
		Type: "ContextSourceRegistration",
		Mode: "inclusive",
		Information: []information{
			{
				Entities: []informationEntities{
					{
						Type: "Service",
					},
					{
						Type: "ServiceComponent",
					},
					{
						Type: "NetworkPort",
					},
					{
						Type: "InfrastructureElementRequirements",
					},
				},
			},
		},
		ContextSourceInfo: []KeyValue{
			{
				Key:   "Authorization",
				Value: NGSILD_PREFIX + "request",
			},
		},
		Operations: []string{
			"retrieveOps",
			"updateOps",
			"deleteEntity",
			"deleteAttrs",
			// "mergeEntity",
		},
		HostAlias: newDomain.BrokerId,
		Endpoint:  newDomain.PublicUrl + "/orionld", // Add /orionld to be aligned with current KrakenD config
		// Endpoint: newDomain.PublicUrl,
		Management: CSRManagement{
			LocalOnly: true,
		},
		AeriosDomain:           newDomain.Name,
		AeriosDomainFederation: true,
	}
}

func NewOrganizationCSR(newDomain *NewDomain) ContextSourceRegistration {
	return ContextSourceRegistration{
		Id:   "urn:aerios:federation:" + strings.ToLower(newDomain.Name) + ":organizations",
		Type: "ContextSourceRegistration",
		Mode: "inclusive",
		Information: []information{
			{
				Entities: []informationEntities{
					{
						Type: "Organization",
					},
				},
			},
		},
		ContextSourceInfo: []KeyValue{
			{
				Key:   "Authorization",
				Value: NGSILD_PREFIX + "request",
			},
		},
		Operations: []string{
			"retrieveOps",
		},
		HostAlias: newDomain.BrokerId,
		Endpoint:  newDomain.PublicUrl + "/orionld", // Add /orionld to be aligned with current KrakenD config
		// Endpoint: newDomain.PublicUrl,
		Management: CSRManagement{
			LocalOnly: true,
		},
		AeriosDomain:           newDomain.Name,
		AeriosDomainFederation: true,
	}
}

func NewBenchmarkCSR(newDomain *NewDomain) ContextSourceRegistration {
	return ContextSourceRegistration{
		Id:   "urn:aerios:federation:" + strings.ToLower(newDomain.Name) + ":benchmark",
		Type: "ContextSourceRegistration",
		Mode: "inclusive",
		Information: []information{
			{
				Entities: []informationEntities{
					{
						Type: "Benchmark",
					},
				},
			},
		},
		ContextSourceInfo: []KeyValue{
			{
				Key:   "Authorization",
				Value: NGSILD_PREFIX + "request",
			},
		},
		Operations: []string{
			"retrieveOps",
		},
		HostAlias: newDomain.BrokerId,
		Endpoint:  newDomain.PublicUrl + "/orionld", // Add /orionld to be aligned with current KrakenD config
		// Endpoint: newDomain.PublicUrl,
		Management: CSRManagement{
			LocalOnly: true,
		},
		AeriosDomain:           newDomain.Name,
		AeriosDomainFederation: true,
	}
}
