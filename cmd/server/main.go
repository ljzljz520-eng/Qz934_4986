package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"membership13/httpapi"
	"membership13/service"
	"membership13/store"
)

func main() {
	path := flag.String("db", "membership.db", "bbolt database path")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()
	st, err := store.Open(*path)
	if err != nil {
		log.Printf("open store: %v", err)
		os.Exit(1)
	}
	defer st.Close()
	svc := service.New(st)
	log.Printf("membership service listening on %s", *addr)
	if err := http.ListenAndServe(*addr, httpapi.New(svc).Routes()); err != nil {
		log.Printf("server stopped: %v", err)
	}
}
