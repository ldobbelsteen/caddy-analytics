package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

//go:embed build/web
var webFiles embed.FS

func main() {

	// Read command line arguments
	logDirectory := flag.String("logs", "/var/log/caddy", "Path to the directory where Caddy's logs are stored")
	maxmindKey := flag.String("geo", "", "MaxMind license key for downloading the geolocation database")
	listeningPort := flag.Int("port", 5734, "Port on which the program should serve the web interface")
	cacheTime := flag.Int("cache", 10, "Number of seconds to cache parse results before discarding")
	flag.Parse()

	// Override arguments with environment variables if they exist
	if logs := os.Getenv("LOGS"); logs != "" {
		*logDirectory = logs
	}
	if geo := os.Getenv("GEO"); geo != "" {
		*maxmindKey = geo
	}
	if port := os.Getenv("PORT"); port != "" {
		if port, err := strconv.Atoi(port); err != nil {
			*listeningPort = port
		}
	}
	if cache := os.Getenv("CACHE"); cache != "" {
		if cache, err := strconv.Atoi(cache); err != nil {
			*cacheTime = cache
		}
	}

	// Validate arguments
	if *cacheTime < 1 {
		log.Fatal("Invalid cache time!")
	}
	if *listeningPort < 1024 || *listeningPort > 65535 {
		log.Fatal("Invalid port number!")
	}
	if *maxmindKey == "" {
		log.Fatal("No MaxMind license key specified!")
	}

	// Download/update geolocation database
	database, err := fetchGeolocationDatabase(*maxmindKey)
	if err != nil {
		log.Fatal("Failed to download/update geolocation database: ", err)
	}

	// For caching parse results in serialized form
	var jsonCache []byte
	var parseWait sync.WaitGroup

	// Handler for serving web files
	webFiles, err := fs.Sub(webFiles, "build/web")
	if err != nil {
		log.Fatal("Failed to open web files root: ", err)
	}
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// Handle function for serving statistics parsed from the logs in JSON format
	http.HandleFunc("/data", func(writer http.ResponseWriter, request *http.Request) {
		parseWait.Wait()
		if jsonCache == nil {
			parseWait.Add(1)
			stats, err := parseLogs(*logDirectory, database)
			if err != nil {
				log.Print("Failed to parse logs: ", err)
				http.Error(writer, "failed to parse logs", http.StatusInternalServerError)
				parseWait.Done()
				return
			}
			jsonCache, err = json.MarshalIndent(stats, "", "  ")
			if err != nil {
				jsonCache = nil
				log.Print("Failed to convert to JSON: ", err)
				http.Error(writer, "failed to convert json", http.StatusInternalServerError)
				parseWait.Done()
				return
			}
			time.AfterFunc(time.Duration(*cacheTime)*time.Second, func() { jsonCache = nil })
			parseWait.Done()
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(jsonCache)
	})

	// Start listening
	portString := strconv.Itoa(*listeningPort)
	listenAddress := ":" + portString
	log.Print("Started listening on port " + portString + "...")
	err = http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
