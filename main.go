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

	_ "code.linksmart.eu/com/go-sec/auth/keycloak/obtainer"
	"code.linksmart.eu/com/go-sec/auth/obtainer"
	"code.linksmart.eu/sc/service-catalog/catalog"
	"code.linksmart.eu/sc/service-catalog/client"
	"github.com/satori/go.uuid"
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
	if *confPath == "" {
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
	unregsiter, err := client.RegisterServiceAndKeepalive(*endpoint, *service, ticket)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ctrl+C handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt)
	for sig := range handler {
		if sig == os.Interrupt {
			log.Println("Caught interrupt signal...")
			break
		}
	}

	err = unregsiter()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Stopped")
	os.Exit(0)
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
