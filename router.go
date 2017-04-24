package main

import (
	"os"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
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
		TraceFatal("Variable OPENSHIFT_GO_IP is not defined")
		os.Exit(-1)
	}

	if apiPort = os.Getenv("OPENSHIFT_GO_PORT"); apiPort == "" {
		TraceFatal("Variable OPENSHIFT_GO_PORT is not defined")
		os.Exit(-1)
	}

	if apiProtocol = os.Getenv("FRAGRANCES_API_PROTOCOL"); apiProtocol == "" {
		TraceFatal("Variable FRAGRANCES_API_PROTOCOL is not defined")
		os.Exit(-1)
	}

	if apiDNS = os.Getenv("OPENSHIFT_APP_DNS"); apiDNS == "" {
		TraceFatal("Variable OPENSHIFT_APP_DNS is not defined")
		os.Exit(-1)
	}

	baseUrl = apiProtocol + "://" + apiDNS + API_PATH
}

// NewRouter ...
func NewRouter() *mux.Router {
	root := mux.NewRouter()

	logger := NewLogger()
	logger.SetDateFormat("02.01.2006 15:04:05.000")

	recovery := negroni.NewRecovery()

	publicRouter := mux.NewRouter().StrictSlash(true)
	for _, route := range publicRoutes {
		publicRouter.Methods(route.Method).Path(API_PATH + route.Pattern).Name(route.Name).Handler(route.Endpoint)
		root.Path(API_PATH + route.Pattern).Handler(negroni.New(
			recovery,
			logger,
			negroni.Wrap(publicRouter)))
	}

	privateRouter := mux.NewRouter().PathPrefix(API_PATH).Subrouter().StrictSlash(true)
	for _, route := range privateRoutes {
		privateRouter.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.Endpoint)
	}
	root.PathPrefix(API_PATH).Handler(negroni.New(
		recovery,
		logger,
		negroni.HandlerFunc(ValidateAccessToken),
		negroni.Wrap(privateRouter)))

	return root
}
