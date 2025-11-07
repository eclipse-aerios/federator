package models

type ApiHealth struct {
	Status               string           `json:"status,omitempty"`
	OrionLdStatus        string           `json:"orionLdStatus,omitempty"`
	Domain               string           `json:"domain,omitempty"`
	DomainStatus         string           `json:"domainStatus"`
	PeerFederatorDomain  string           `json:"peerFederatorDomain"`
	PeerFederatorStatus  string           `json:"peerFederatorStatus"`
	IsEntrypoint         bool             `json:"isEntrypoint,omitempty"`
	Message              string           `json:"message"`
	DetailedErrorMessage string           `json:"detailedErrorMessage,omitempty"`
	FederatedDomains     FederatedDomains `json:"federatedDomains,omitempty"`
}

type FederatedDomains struct {
	Total int    `json:"total,omitempty"`
	Names string `json:"names,omitempty"`
}
