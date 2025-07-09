package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"go.roman.zone/mullvad-web-controller/api"
)

//go:embed content/*.html content/static/*
var content embed.FS

func main() {
	host := flag.String("host", "0.0.0.0", "Host to listen on")
	port := flag.Int("port", 8666, "Port to listen on")
	devMode := flag.Bool("dev", false, "Whether to actually send requests or not")
	flag.Parse()

	actualContent, _ := fs.Sub(content, "content")

	http.Handle("/", http.FileServer(http.FS(actualContent)))

	http.HandleFunc("/api/relays", api.NewRelayLocationsHandler(time.Hour*24).Handle)
	http.HandleFunc("/api/relay/location", api.NewRelayLocationChangeHandler(*devMode).Handle)
	http.HandleFunc("/api/ip", api.HandleIPRetrieval)

	fmt.Printf("Server listening on http://%s:%d\n", *host, *port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
