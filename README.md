### LinkSmart Service Registrar
Service Registrar is a tool to register and update services in [Service Catalog](https://docs.linksmart.eu/SC). 

Documentation is available in the [wiki](https://docs.linksmart.eu/display/SC/Service+Registrar).


## Compile from source
```
git clone https://code.linksmart.eu/scm/sc/service-registrar.git src/code.linksmart.eu/sc/service-registrar
export GOPATH=`pwd`
go install code.linksmart.eu/sc/service-registrar
```

## Development
The dependencies of this package are managed by [dep](https://github.com/golang/dep).
