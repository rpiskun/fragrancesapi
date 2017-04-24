package main

import (
	"fmt"
	_ "gopkg.in/gorp.v1"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var err error

func main() {
	dbmap := InitDb()
	defer dbmap.Db.Close()
	api := NewRouter()
	bind := fmt.Sprintf("%s:%s", apiHost, apiPort)

	go func() {
		// for pprof
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	log.Println("Server started at:", bind)
	err = http.ListenAndServe(bind, api)
	if err != nil {
		log.Fatalln("Couldn't start server at:", bind)
		os.Exit(-1)
	}
}
