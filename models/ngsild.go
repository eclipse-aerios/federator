package models

import (
	"errors"
	"strings"
)

const NGSILD_PREFIX = "urn:ngsi-ld:"

type Property struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Relationship struct {
	Type   string `json:"type"`
	Object string `json:"object"`
}

type MultipleRelationship struct {
	Type   string   `json:"type"`
	Object []string `json:"object"`
}

func NewRelationship(object string) Relationship {
	return Relationship{
		Type:   "Relationship",
		Object: object,
	}
}

func NewMultipleRelationship(objects ...string) (MultipleRelationship, error) {
	if len(objects) >= 1 {
		return MultipleRelationship{
			Type:   "Relationship",
			Object: objects,
		}, nil
	} else {
		return MultipleRelationship{}, errors.New("at least one object is required to create a relationship attribute")
	}
}

func BuildNgsiLdEntityId(entityType string, value string) string {
	return NGSILD_PREFIX + entityType + ":" + value
}

func GetNgsiLdEntityIdValue(entityType string, value string) string {
	return strings.ReplaceAll(value, NGSILD_PREFIX+entityType+":", "")
}

type SourceIdentity struct {
	Id                  string              `json:"id"`
	Type                string              `json:"type"`
	ContextSourceUptime string              `json:"contextSourceUptime"`
	ContextSourceTimeAt string              `json:"contextSourceTimeAt"`
	ContextSourceAlias  string              `json:"contextSourceAlias"`
	ContextSourceExtras ContextSourceExtras `json:"contextSourceExtras"`
}

type ContextSourceExtras struct {
	OrionLDVersion string `json:"Orion-LD version"`
	Branch         string `json:"branch"`
	CoreContext    string `json:"Core Context"`
}
