package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {

	// Read command line arguments
	logDirectory := flag.String("logs", "/var/log/caddy", "Path to the directory where Caddy's logs are stored")
	geoDatabase := flag.String("geo", "/etc/maxmind/country.mmdb", "Path to your .mmdb file")
	listeningPort := flag.String("port", "5734", "Port on which the program should listen")
	webDirectory := flag.String("web", "web/dist", "Path to the directory of the web interface")
	flag.Parse()

	// For caching parse results in serialized form
	var jsonCache []byte
	var parseWait sync.WaitGroup

	// Handler for serving web files
	http.Handle("/", http.FileServer(http.Dir(*webDirectory)))

	// Handle function for serving statistics parsed from the logs in JSON format
	http.HandleFunc("/stats", func(writer http.ResponseWriter, request *http.Request) {
		parseWait.Wait()
		if jsonCache == nil {
			parseWait.Add(1)
			func() {
				defer parseWait.Done()
				stats, err := parseLogs(*logDirectory, *geoDatabase)
				if err != nil {
					http.Error(writer, "failed to parse logs", http.StatusInternalServerError)
					return
				}
				jsonCache, err = json.MarshalIndent(stats, "", "  ")
				if err != nil {
					jsonCache = nil
					http.Error(writer, "failed to convert json", http.StatusInternalServerError)
					return
				} else {
					time.AfterFunc(10*time.Second, func() { jsonCache = nil })
				}
			}()
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonCache)
	})

	// Start listening
	listenAddress := ":" + *listeningPort
	log.Print("Started listening on port " + *listeningPort + "...")
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
