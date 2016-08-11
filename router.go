package main

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
)

const (
	API_PATH = "/api/v1"
)

var (
	apiHost     = ""
	apiPort     = ""
	apiDNS      = ""
	apiProtocol = ""
	baseUrl     = ""
)

func init() {
	if apiHost = os.Getenv("OPENSHIFT_GO_IP"); apiHost == "" {
		log.Fatalln("Variable OPENSHIFT_GO_IP is not defined")
		os.Exit(-1)
	}

	if apiPort = os.Getenv("OPENSHIFT_GO_PORT"); apiPort == "" {
		log.Fatalln("Variable OPENSHIFT_GO_PORT is not defined")
		os.Exit(-1)
	}

	if apiProtocol = os.Getenv("FRAGRANCES_API_PROTOCOL"); apiProtocol == "" {
		log.Fatalln("Variable FRAGRANCES_API_PROTOCOL is not defined")
		os.Exit(-1)
	}

	if apiDNS = os.Getenv("OPENSHIFT_APP_DNS"); apiDNS == "" {
		log.Fatalln("Variable OPENSHIFT_APP_DNS is not defined")
		os.Exit(-1)
	}

	baseUrl = apiProtocol + "://" + apiDNS + API_PATH
}

func redirector(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if apiProtocol == "https" && r.TLS == nil {
		http.Redirect(w, r, "https://"+apiDNS+r.URL.String(), http.StatusMovedPermanently)
		return
	}
	next(w, r)
}

// NewRouter ...
func NewRouter() *mux.Router {
	root := mux.NewRouter()

	googleValidationRouter := mux.NewRouter().PathPrefix(API_PATH + "/users").Subrouter().StrictSlash(true)
	for _, route := range validateByGoogle {
		googleValidationRouter.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.Endpoint)
	}
	root.PathPrefix(API_PATH + "/users").Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.HandlerFunc(redirector),
		negroni.HandlerFunc(ValidateTokenByGoogle),
		negroni.Wrap(googleValidationRouter)))

	dbValidationRouter := mux.NewRouter().PathPrefix(API_PATH).Subrouter().StrictSlash(true)
	for _, route := range validateByDatabase {
		dbValidationRouter.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.Endpoint)
	}
	root.PathPrefix(API_PATH).Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.HandlerFunc(redirector),
		negroni.HandlerFunc(ValidateTokenByDatabase),
		negroni.Wrap(dbValidationRouter)))

	return root
}
