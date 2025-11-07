package models

type Domain struct {
	Id           string               `json:"id"`
	Type         string               `json:"type"`
	Description  string               `json:"description,omitempty"`
	PublicUrl    string               `json:"publicUrl,omitempty"`
	Owner        MultipleRelationship `json:"owner,omitempty"`
	IsEntrypoint bool                 `json:"isEntrypoint"` // TODO solve boolean marshalling (if omitempty, when false, it is omited)
	DomainStatus Relationship         `json:"domainStatus,omitempty"`
	FederatorUrl string               `json:"federatorUrl,omitempty"`
	PublicKey    string               `json:"publicKey"`
}

type DomainSimplified struct {
	Id           string   `json:"id"`
	Type         string   `json:"type"`
	Description  string   `json:"description,omitempty"`
	PublicUrl    string   `json:"publicUrl,omitempty"`
	Owner        []string `json:"owner,omitempty"`
	IsEntrypoint bool     `json:"isEntrypoint,omitempty"`
	DomainStatus string   `json:"domainStatus,omitempty"`
	FederatorUrl string   `json:"federatorUrl,omitempty"`
	PublicKey    string   `json:"publicKey"`
}

type NewDomain struct {
	Name         string `json:"name" binding:"required"`
	PublicUrl    string `json:"publicUrl" binding:"required"`
	IsEntrypoint bool   `json:"isEntrypoint"` // TODO check binding:"required"
	BrokerId     string `json:"brokerId" binding:"required"`
}

type NewDomainSpreadResponse struct {
	NewRegistrations       []ContextSourceRegistration `json:"newRegistrations,omitempty"`
	Domains                []DomainSimplified          `json:"domains,omitempty"`
	NewDomainRegistrations []ContextSourceRegistration `json:"newDomainRegistrations,omitempty"`
	FailedDomains          []string                    `json:"failedDomains,omitempty"`
	Message                string                      `json:"message,omitempty"`
}

type DeleteDomainSpreadResponse struct {
	FailedDomains []string `json:"failedDomains,omitempty"`
	Message       string   `json:"message,omitempty"`
}
