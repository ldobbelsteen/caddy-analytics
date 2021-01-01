package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

// Format of a line in Caddy's access logs
type LogEntry struct {
	Stamp    float64 `json:"ts"`
	Status   int     `json:"status"`
	Duration float64 `json:"duration"`
	Size     int     `json:"size"`
	Request  struct {
		Address    string `json:"remote_addr"`
		Protocol   string `json:"proto"`
		Method     string `json:"method"`
		Host       string `json:"host"`
		Location   string `json:"uri"`
		Encryption struct {
			Version uint16 `json:"version"`
			Cipher  uint16 `json:"cipher_suite"`
		} `json:"tls"`
		Headers struct {
			Languages []string `json:"Accept-Language"`
			Encodings []string `json:"Accept-Encoding"`
			UserAgent []string `json:"User-Agent"`
		} `json:"headers"`
	} `json:"request"`
	Response struct {
		ContentType []string `json:"Content-Type"`
	} `json:"resp_headers"`
}

func parseLogs(logDir string, geoFile string) (Statistics, error) {

	stats := Statistics{
		Counters:     map[string]*Counter{},
		LogDirectory: logDir,
	}

	info, err := os.Stat(logDir)
	if err != nil {
		log.Print("Log directory does not exist: ", err)
		return stats, err
	}
	if !info.Mode().IsDir() {
		err = errors.New("stat " + logDir + ": not a directory")
		log.Print("Log directory is not a directory: ", err)
		return stats, err
	}

	logFiles, _ := filepath.Glob(filepath.Join(logDir, "*.log*"))

	startTime := time.Now()

	for _, logFile := range logFiles {

		file, err := os.Open(logFile)
		if err != nil {
			log.Print("Log file could not be opened: ", err)
			return stats, err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			log.Print("Log file stats could not be retrieved: ", err)
			return stats, err
		}
		stats.LogSizeBytes += info.Size()

		scanner := bufio.NewScanner(file)
		if filepath.Ext(logFile) == ".gz" {
			decompressed, err := gzip.NewReader(file)
			if err != nil {
				log.Print("Log file could not be decompressed: ", err)
				return stats, err
			}
			scanner = bufio.NewScanner(decompressed)
		}

		for scanner.Scan() {
			var line LogEntry
			err := json.Unmarshal(scanner.Bytes(), &line)
			if err != nil {
				log.Print("Log line could not be unmarshalled: ", err)
				return stats, err
			}
			addToStats(&line, &stats)
		}

		file.Close()
	}

	geo, err := maxminddb.Open(geoFile)
	if err != nil {
		log.Print("Geolocation database could not be opened: ", err)
		return stats, err
	}
	defer geo.Close()

	for _, counter := range stats.Counters {
		for visitor := range counter.Total.Observed {
			var info struct {
				Country struct {
					Names map[string]string `maxminddb:"names"`
				} `maxminddb:"country"`
			}
			ip := net.ParseIP(visitor.IP)
			err := geo.Lookup(ip, &info)
			if err != nil {
				log.Print("IP country lookup failed: ", err)
				return stats, err
			}
			country := info.Country.Names["en"]
			counter.Countries[country] += 1
		}
	}

	geo.Close()
	stats.ParseDuration = time.Since(startTime).Seconds()
	return stats, nil
}
