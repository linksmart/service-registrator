package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	catalog "linksmart.eu/localconnect/core/catalog/resource"
	//"linksmart.eu/localconnect/core/discovery"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	CatalogBackendMemory = "memory"
	StaticLocation       = "/static"
)

var (
	confPath  = flag.String("conf", "conf/resource-catalog.json", "Device catalog configuration file path")
	staticDir = ""
)

func staticHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// serve all /static/ctx files as ld+json
	if strings.HasPrefix(req.URL.Path, "/static/ctx") {
		w.Header().Set("Content-Type", "application/ld+json")
	}
	urlParts := strings.Split(req.URL.Path, "/")
	http.ServeFile(w, req, staticDir+"/"+strings.Join(urlParts[2:], "/"))
}

func main() {
	flag.Parse()

	config, err := loadConfig(*confPath)
	if err != nil {
		log.Fatalf("Error reading config file %v:%v", *confPath, err)
	}
	staticDir = config.StaticDir

	var cat catalog.CatalogStorage

	switch config.Storage.Type {
	case CatalogBackendMemory:
		cat = catalog.NewMemoryStorage()
	}

	api := catalog.NewWritableCatalogAPI(cat, config.ApiLocation, StaticLocation)

	m := pat.New()
	// writable api
	m.Post(config.ApiLocation+"/", http.HandlerFunc(api.Add))

	m.Get(fmt.Sprintf("%s/%s/%s",
		config.ApiLocation, catalog.PatternUuid, catalog.PatternReg),
		http.HandlerFunc(api.Get))

	m.Get(fmt.Sprintf("%s/%s/%s/%s",
		config.ApiLocation, catalog.PatternUuid, catalog.PatternReg, catalog.PatternRes),
		http.HandlerFunc(api.GetResource))

	m.Get(fmt.Sprintf("%s/%s/%s/%s/%s",
		config.ApiLocation, catalog.PatternFType, catalog.PatternFPath, catalog.PatternFOp, catalog.PatternFValue),
		http.HandlerFunc(api.Filter))

	m.Put(fmt.Sprintf("%s/%s/%s",
		config.ApiLocation, catalog.PatternUuid, catalog.PatternReg),
		http.HandlerFunc(api.Update))

	m.Del(fmt.Sprintf("%s/%s/%s",
		config.ApiLocation, catalog.PatternUuid, catalog.PatternReg),
		http.HandlerFunc(api.Delete))

	m.Get(config.ApiLocation, http.HandlerFunc(api.List))

	// static
	m.Get("/static/", http.HandlerFunc(staticHandler))

	http.Handle("/", m)

	// Announce serice using DNS-SD
	/*
		if config.DnssdEnabled {
			_, err := discovery.DnsRegisterService(config.Description, catalog.DnssdServiceType, config.Port)
			if err != nil {
				log.Printf("Failed to perform DNS-SD registration: %v\n", err.Error())
			}
		}
	*/

	log.Printf("Starting standalone Device Catalog at %v:%v%v", config.BindAddr, config.BindPort, config.ApiLocation)

	// Register in Service Catalogs if configured
	if len(config.ServiceCatalog) > 0 {
		log.Println("Will now register in the configured Service Catalogs")
		registerService(config)
	}

	// Listen and Serve
	endpoint := fmt.Sprintf("%s:%s", config.BindAddr, strconv.Itoa(config.BindPort))
	log.Fatal(http.ListenAndServe(endpoint, nil))
}
