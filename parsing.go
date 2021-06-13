package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

type logEntry struct {
	Stamp    float64 `json:"ts"`
	Status   uint16  `json:"status"`
	Duration float64 `json:"duration"`
	Size     int64   `json:"size"`
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

// Parse all logs in the log directory and return the statistics
func parseLogs(logDir string, geoFile string) (*statistics, error) {

	// Create statistics instance
	stats := newStatistics()
	stats.Directory = logDir

	// Validate log directory
	info, err := os.Stat(logDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("stat " + logDir + ": not a directory")
	}

	// Find log files with log extension
	logFiles, _ := filepath.Glob(filepath.Join(logDir, "*.log*"))

	// Start the timer
	startTime := time.Now()

	// Read all files one by one
	for _, logFile := range logFiles {

		// Open the log file
		file, err := os.Open(logFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Get log file size in bytes
		info, err := file.Stat()
		if err != nil {
			return nil, err
		}
		stats.SizeBytes += info.Size()

		// Create line scanner and decompress if the file is gzipped
		scanner := bufio.NewScanner(file)
		if filepath.Ext(logFile) == ".gz" {
			decompressed, err := gzip.NewReader(file)
			if err != nil {
				return nil, err
			}
			scanner = bufio.NewScanner(decompressed)
		}

		// Scan line by line and add them to the statistics instance
		for scanner.Scan() {
			var line logEntry
			err := json.Unmarshal(scanner.Bytes(), &line)
			if err != nil {
				return nil, err
			}
			err = addToStats(&line, stats)
			if err != nil {
				return nil, err
			}
		}

		file.Close()
	}

	// Open geolocation database
	geo, err := maxminddb.Open(geoFile)
	if err != nil {
		return nil, err
	}
	defer geo.Close()

	// Get countries of all visitors observed
	for _, counter := range stats.Hosts {
		for visitor := range counter.Total.ObservedUsers {
			var info struct {
				Country struct {
					Names map[string]string `maxminddb:"names"`
				} `maxminddb:"country"`
			}
			ip := net.ParseIP(visitor.IPAddress)
			err := geo.Lookup(ip, &info)
			if err != nil {
				return nil, err
			}
			country := info.Country.Names["en"]
			if country == "" {
				country = "Unknown"
			}
			counter.Visitors.Countries[country]++
		}
	}

	// Save parse duration and return result
	stats.ParseDuration = time.Since(startTime).Seconds()
	return stats, nil
}
