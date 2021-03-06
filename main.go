// Copyright 2014-2016 Fraunhofer Institute for Applied Information Technology FIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	_ "github.com/linksmart/go-sec/auth/keycloak/obtainer"
	"github.com/linksmart/go-sec/auth/obtainer"
	"github.com/linksmart/service-catalog/v2/catalog"
	"github.com/linksmart/service-catalog/v2/client"
	uuid "github.com/satori/go.uuid"
)

var (
	confPath = flag.String("conf", "", "Path to the service configuration file")
	endpoint = flag.String("endpoint", "", "Service Catalog endpoint")
	//discover = flag.Bool("discover", false, "Use DNS-SD service discovery to find Service Catalog endpoint")
	// Authentication configuration
	authProvider    = flag.String("authProvider", "", "Authentication provider name")
	authProviderURL = flag.String("authProviderURL", "", "Authentication provider url")
	authUser        = flag.String("authUser", "", "Auth. server username")
	authPass        = flag.String("authPass", "", "Auth. server password")
	serviceID       = flag.String("serviceID", "", "Service ID at the auth. server")
)

const LINKSMART = `
╦   ╦ ╔╗╔ ╦╔═  ╔═╗ ╔╦╗ ╔═╗ ╦═╗ ╔╦╗ R
║   ║ ║║║ ╠╩╗  ╚═╗ ║║║ ╠═╣ ╠╦╝  ║
╩═╝ ╩ ╝╚╝ ╩ ╩  ╚═╝ ╩ ╩ ╩ ╩ ╩╚═  ╩
`

func main() {
	flag.Parse()
	if *confPath == "" || *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Print(LINKSMART)
	log.Printf("Starting Service Registrator")

	// requiresAuth if authProvider is specified
	var requiresAuth bool = (*authProvider != "")

	service, err := LoadConfigFromFile(*confPath)
	if err != nil {
		log.Fatal("Unable to read service configuration from file: ", err)
	}

	if service.TTL == 0 {
		log.Fatal("TTL must be larger than zero")
	}

	if service.ID == "" {
		service.ID = uuid.NewV4().String()
		log.Printf("ID not set, generated UUID: %s", service.ID)
	} else {
		log.Printf("Loaded service with ID: %s", service.ID)
	}

	var ticket *obtainer.Client
	if requiresAuth {
		// Setup ticket client
		ticket, err = obtainer.NewClient(*authProvider, *authProviderURL, *authUser, *authPass, *serviceID)
		if err != nil {
			log.Fatal(err.Error())
		}

	}

	// Launch the registration routine
	_, _, err = client.RegisterServiceAndKeepalive(*endpoint, *service, ticket)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ctrl+C / Kill handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt, os.Kill)
	<-handler
	log.Println("Shutting down...")
}

// Loads service registration from a config file
func LoadConfigFromFile(confPath string) (*catalog.Service, error) {

	f, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	var service catalog.Service
	err = json.Unmarshal(f, &service)
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %s", err)
	}

	return &service, err
}
