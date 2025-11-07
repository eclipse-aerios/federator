package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/eclipse-aerios/federator/config"
	"github.com/eclipse-aerios/federator/router"
	"github.com/eclipse-aerios/federator/utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("aeriOS Federator")
	log.Println("Developed in Go, REST API using GinGonic framework")

	// Load environment variables
	config.LoadEnvVars()
	if !config.TLS_CERTIFICATE_VALIDATION {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	initialization := &utils.Initialization{}
	err := initialization.InitializeFederator()

	if err != nil {
		log.Println(err)
		log.Fatalln("Couldn't initialize the aeriOS Federator in this domain")
	}
	log.Println("aeriOS Federator successfully initialized")
	log.Println("=============================================================")

	app := router.NewRouter()
	app.Run(":" + config.APP_PORT)
}
