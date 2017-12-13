package main

import (
	"flag"
	catalog "linksmart.eu/localconnect/core/catalog/resource"
	"log"
	"os"
	"os/signal"
)

var (
	confPath = flag.String("conf", "conf/device-gateway.json", "Device gateway configuration file path")
)

func main() {
	log.SetPrefix("[device-gateway] ")
	log.SetFlags(log.Ltime)

	flag.Parse()
	if *confPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := loadConfig(*confPath)
	if err != nil {
		log.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Publish device data to MQTT (if require)
	mqttPublisher := newMQTTPublisher(config)

	// Start the agent programs and establish internal communication
	agentManager := newAgentManager(config)
	if mqttPublisher != nil {
		go mqttPublisher.start()
		agentManager.setPublishingChannel(mqttPublisher.dataInbox())
	}
	go agentManager.start()

	// Expose device's resources via REST (include statics and local catalog)
	restServer := newRESTfulAPI(config, agentManager.DataRequestInbox())
	catalogStorage := catalog.NewMemoryStorage()
	go restServer.start(catalogStorage)

	// Register devices in the local catalog and run periodic remote catalog updates (if required)
	go registerDevices(config, catalogStorage)

	/*
		// Announce serice using DNS-SD
		dnsRegistration, err := dnsRegisterService(config)
		if err != nil {
			log.Printf("Failed to perform DNS-SD registration: %v\n", err.Error())
		}

		or

		consider this:
		if config.DnssdEnabled {
			parts := strings.Split(config.Endpoint, ":")
			port, _ := strconv.Atoi(parts[1])
			_, err := discovery.DnsRegisterService(config.Name, catalog.DnssdServiceType, port)
			if err != nil {
				log.Printf("Failed to perform DNS-SD registration: %v\n", err.Error())
			}
		}
	*/

	// Ctrl+C handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt)
	for sig := range handler {
		if sig == os.Interrupt {
			log.Println("Caught interrupt signal...")
			break
		}
	}

	agentManager.stop()
	if mqttPublisher != nil {
		mqttPublisher.stop()
	}
	/*
		if dnsRegistration != nil {
			dnsRegistration.Stop()
		}
	*/
	// Remove registratoins from configured remote catalogs
	if len(config.Catalog) > 0 {
		unregisterDevices(config, catalogStorage)
	}

	log.Println("Stopped")
	os.Exit(0)
}
