package controllers

import (
	"net/http"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/models"
	"github.com/gin-gonic/gin"
)

var buildTime string
var commitHash string

type VersionController struct {
}

func (v *VersionController) Version(c *gin.Context) {
	apiVersion := models.ApiVersion{
		Version:     config.API_VERSION,
		BuildTime:   buildTime,
		CommitHash:  commitHash,
		ServiceName: config.SERVICE_NAME,
	}
	c.JSON(http.StatusOK, apiVersion)
}
