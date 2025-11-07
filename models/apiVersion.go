package models

type ApiVersion struct {
	Version     string `json:"version"`
	BuildTime   string `json:"buildTime"`
	CommitHash  string `json:"commitHash"`
	ServiceName string `json:"serviceName"`
}
