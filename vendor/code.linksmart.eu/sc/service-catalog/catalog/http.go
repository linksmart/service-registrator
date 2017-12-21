// Copyright 2014-2016 Fraunhofer Institute for Applied Information Technology FIT

package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"strings"

	"code.linksmart.eu/sc/service-catalog/utils"
	"github.com/gorilla/mux"
)

type HttpAPI struct {
	controller  *Controller
	id          string
	description string
	version     string
}

// NewHTTPAPI creates a RESTful HTTP API
func NewHTTPAPI(controller *Controller, id, description, version string) *HttpAPI {
	return &HttpAPI{
		controller:  controller,
		id:          id,
		description: description,
		version:     version,
	}
}

// Collection is the paginated list of services
type Collection struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Services    []Service `json:"services"`
	Page        int       `json:"page"`
	PerPage     int       `json:"per_page"`
	Total       int       `json:"total"`
}

// API Index: Lists services
func (a *HttpAPI) List(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error parsing the query:", err.Error())
		return
	}
	page, perPage, err := utils.ParsePagingParams(
		req.Form.Get(utils.GetParamPage), req.Form.Get(utils.GetParamPerPage), MaxPerPage)
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error parsing query parameters:", err.Error())
		return
	}

	services, total, err := a.controller.list(page, perPage)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	coll := &Collection{
		ID:          a.id,
		Description: a.description,
		Services:    services,
		Page:        page,
		PerPage:     perPage,
		Total:       total,
	}

	b, err := json.Marshal(coll)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.Write(b)
}

// Filters services
func (a *HttpAPI) Filter(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	path := params["path"]
	op := params["op"]
	value := params["value"]

	err := req.ParseForm()
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error parsing the query:", err.Error())
		return
	}
	page, perPage, err := utils.ParsePagingParams(
		req.Form.Get(utils.GetParamPage), req.Form.Get(utils.GetParamPerPage), MaxPerPage)
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error parsing query parameters:", err.Error())
		return
	}

	services, total, err := a.controller.filter(path, op, value, page, perPage)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	coll := &Collection{
		ID:          a.id,
		Description: a.description,
		Services:    services,
		Page:        page,
		PerPage:     perPage,
		Total:       total,
	}

	b, err := json.Marshal(coll)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.Write(b)
}

// Retrieves a service
func (a *HttpAPI) Get(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	s, err := a.controller.get(params["id"])
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			a.ErrorResponse(w, http.StatusNotFound, err.Error())
			return
		default:
			a.ErrorResponse(w, http.StatusInternalServerError, "Error retrieving the service:", err.Error())
			return
		}
	}

	b, err := json.Marshal(s)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.Write(b)
}

func (a *HttpAPI) createService(w http.ResponseWriter, s *Service) {
	addedS, err := a.controller.add(*s)
	if err != nil {
		switch err.(type) {
		case *ConflictError:
			a.ErrorResponse(w, http.StatusConflict, "Error creating the registration:", err.Error())
			return
		case *BadRequestError:
			a.ErrorResponse(w, http.StatusBadRequest, "Invalid service registration:", err.Error())
			return
		default:
			a.ErrorResponse(w, http.StatusInternalServerError, "Error creating the registration:", err.Error())
			return
		}
	}

	b, err := json.Marshal(addedS)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.Header().Set("Location", fmt.Sprintf("/%s", addedS.ID))
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

// Adds a service
func (a *HttpAPI) Post(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Body.Close()

	var s Service
	if err := json.Unmarshal(body, &s); err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error processing the request:", err.Error())
		return
	}

	if s.ID != "" {
		a.ErrorResponse(w, http.StatusBadRequest, "Creating a service with defined ID is not possible using a POST request.")
		return
	}

	a.createService(w, &s)
	return
}

// Updates an existing service (Response: StatusOK)
// or creates a new one with the given id (Response: StatusCreated)
func (a *HttpAPI) Put(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var s Service
	if err := json.Unmarshal(body, &s); err != nil {
		a.ErrorResponse(w, http.StatusBadRequest, "Error processing the request:", err.Error())
		return
	}

	updatedS, err := a.controller.update(params["id"], s)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			// Create a new service with the given id
			s.ID = params["id"]
			a.createService(w, &s)
			return
		case *ConflictError:
			a.ErrorResponse(w, http.StatusConflict, "Error updating the service:", err.Error())
			return
		case *BadRequestError:
			a.ErrorResponse(w, http.StatusBadRequest, "Invalid service registration:", err.Error())
			return
		default:
			a.ErrorResponse(w, http.StatusInternalServerError, "Error updating the service:", err.Error())
			return
		}
	}

	b, err := json.Marshal(updatedS)
	if err != nil {
		a.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Deletes a service
func (a *HttpAPI) Delete(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	err := a.controller.delete(params["id"])
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			a.ErrorResponse(w, http.StatusNotFound, err.Error())
			return
		default:
			a.ErrorResponse(w, http.StatusInternalServerError, "Error deleting the service:", err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.WriteHeader(http.StatusOK)
}

// a.ErrorResponse writes error to HTTP ResponseWriter
func (a *HttpAPI) ErrorResponse(w http.ResponseWriter, code int, msgs ...string) {
	msg := strings.Join(msgs, " ")
	e := &Error{
		code,
		msg,
	}
	if code >= 500 {
		logger.Println("ERROR:", msg)
	}
	b, _ := json.Marshal(e)
	w.Header().Set("Content-Type", "application/json;version="+a.version)
	w.WriteHeader(code)
	w.Write(b)
}
