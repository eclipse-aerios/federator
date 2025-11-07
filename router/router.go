package router

import (
	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/controllers"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	if config.APP_ENV == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	health := &controllers.HealthController{}
	version := new(controllers.VersionController)

	router.GET("/health", health.Status)
	router.GET("/version", version.Version)
	// router.Use(middlewares.AuthMiddleware())

	v1 := router.Group("v1")
	{
		domainsGroup := v1.Group("domains")
		{
			dc := new(controllers.DomainController)
			domainsGroup.GET("/", dc.List)
			domainsGroup.GET("/local", dc.GetLocalDomain)
			domainsGroup.POST("", dc.NewDomain)
			if config.IS_ENTRYPOINT {
				domainsGroup.DELETE("/:domainName/spread", dc.SpreadDomainDeletion)
			}
			domainsGroup.DELETE("/local", dc.DeleteLocalDomain)
			domainsGroup.DELETE("/:domainName", dc.Delete)
			// TODO PATCH to enable/disable a domain (e.g. untrusted status)
			// -> also add /local and /spread such as DELETE
			// domainsGroup.PATCH("/:domainName", dc.Enable)
			// domainsGroup.PATCH("/:domainName", dc.Disable)
		}
	}
	return router

}
