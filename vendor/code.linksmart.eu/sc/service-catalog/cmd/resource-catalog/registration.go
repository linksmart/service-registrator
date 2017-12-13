package main

import (
	"encoding/json"
	"fmt"
	dc "linksmart.eu/localconnect/core/catalog/resource"
	sc "linksmart.eu/localconnect/core/catalog/service"
	"log"
)

const (
	registrationTemplate = `
	{
	  "meta": {
	    "serviceType": "",
	    "apiVersion": ""
	  },
	  "protocols": [
	    {
	      "type": "REST",
	      "endpoint": {
	        "url": ""
	      },
	      "methods": [
	        "GET",
	        "POST"
	      ],
	      "content-types": [
	        "application/ld+json"
	      ]
	    }
	  ],
	  "representation": {
	    "application/ld+json": {}
	  }
	}
	`
)

func registrationFromConfig(config *Config) *sc.ServiceConfig {
	serviceConfig := &sc.ServiceConfig{}

	json.Unmarshal([]byte(registrationTemplate), serviceConfig)
	serviceConfig.Name = dc.ApiCollectionType
	serviceConfig.Host = config.PublicAddr
	serviceConfig.Description = config.Description

	// meta
	serviceConfig.Meta["serviceType"] = dc.DnssdServiceType
	serviceConfig.Meta["apiVersion"] = dc.ApiVersion

	// protocols
	// port from the bind port, address from the public address
	serviceConfig.Protocols[0].Endpoint["url"] = fmt.Sprintf("http://%v:%v%v", config.PublicAddr, config.BindPort, config.ApiLocation)

	return serviceConfig
}

// Registers service in all configured catalogs
func registerService(config *Config) {
	serviceConfig := registrationFromConfig(config)

	for _, cat := range config.ServiceCatalog {
		// Ignore endpoint: discover and register
		if cat.Discover == true {
			// TODO: implement discovery of service catalog and register in it
		} else {
			// Register in the catalog specified by endpoint
			registrator := sc.NewRegistrator(cat.Endpoint)
			log.Printf("Registering in the Service Catalog at %s\n", cat.Endpoint)

			// Set TTL
			serviceConfig.Ttl = cat.Ttl

			err := registrator.RegisterService(serviceConfig, true)
			if err != nil {
				log.Printf("Error registering in Service Catalog %v: %v\n", cat.Endpoint, err)
			}
		}
	}
}
