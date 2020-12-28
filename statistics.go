package main

import (
	"crypto/tls"
	"log"
	"net"
	"strings"

	"github.com/mssola/user_agent"
)

type Statistics struct {
	LogDir        string              `json:"logDirectory"`
	LogSize       int64               `json:"logSizeBytes"`
	LogLines      int                 `json:"logLines"`
	ParseDuration float64             `json:"parseDurationSeconds"`
	FirstStamp    float64             `json:"firstStampUnix"`
	LastStamp     float64             `json:"lastStampUnix"`
	Counters      map[string]*Counter `json:"visitors"`
}

type Counter struct {
	Requests     int                   `json:"totalRequests"`
	Duration     float64               `json:"totalLatency"`
	Bytes        int                   `json:"totalSentBytes"`
	Unique       int                   `json:"uniqueVisitors"`
	Nature       map[string]int        `json:"nature"`
	Browsers     map[string]int        `json:"browsers"`
	Platforms    map[string]int        `json:"platforms"`
	Visitors     map[UniqueVisitor]int `json:"-"`
	Referers     map[string]int        `json:"referers"`
	Ciphers      map[string]int        `json:"ciphers"`
	Tls          map[string]int        `json:"tls"`
	Countries    map[string]int        `json:"countries"`
	Methods      map[string]int        `json:"methods"`
	Locations    map[string]int        `json:"locations"`
	Statuses     map[int]int           `json:"statuses"`
	Protocols    map[string]int        `json:"protocols"`
	Languages    map[string]int        `json:"preferredLanguages"`
	Encodings    map[string]int        `json:"encodings"`
	ContentTypes map[string]int        `json:"contentTypes"`
}

type UniqueVisitor struct {
	IP    string
	Agent string
}

func addToStats(stats *Statistics, data *LogLine) {

	// Get the counter corresponding to the host
	host := data.Request.Host
	counter := stats.Counters[host]

	// Create counter if it doesn't yet exist
	if counter == nil {
		stats.Counters[host] = &Counter{
			Browsers:     map[string]int{},
			Visitors:     map[UniqueVisitor]int{},
			Platforms:    map[string]int{},
			Nature:       map[string]int{},
			Tls:          map[string]int{},
			Ciphers:      map[string]int{},
			Countries:    map[string]int{},
			Methods:      map[string]int{},
			Referers:     map[string]int{},
			Locations:    map[string]int{},
			Statuses:     map[int]int{},
			Protocols:    map[string]int{},
			Languages:    map[string]int{},
			Encodings:    map[string]int{},
			ContentTypes: map[string]int{},
		}
		counter = stats.Counters[host]
	}

	// Increment log line count
	stats.LogLines += 1
	counter.Requests += 1

	// Get the pure visitor IP
	ip, _, err := net.SplitHostPort(data.Request.Address)
	if err != nil {
		log.Fatal("Failed to parse remote address: ", err)
	}

	// Get the first user agent
	var rawUserAgent string
	if len(data.Request.Headers.UserAgent) > 0 {
		rawUserAgent = data.Request.Headers.UserAgent[0]
	}

	// Record observation of current visitor
	counter.Visitors[UniqueVisitor{ip, rawUserAgent}] += 1

	// Parse and analyze raw user agent
	uaInfo := user_agent.New(rawUserAgent)
	if uaInfo.Bot() {
		counter.Nature["bot"] += 1
	} else if uaInfo.Mobile() {
		counter.Nature["mobile"] += 1
	} else {
		counter.Nature["other"] += 1
	}
	platform := uaInfo.OS()
	if platform == "" {
		platform = "Unknown"
	}
	counter.Platforms[platform] += 1
	browser, _ := uaInfo.Browser()
	if browser == "" {
		browser = "Unknown"
	}
	counter.Browsers[browser] += 1

	// Parse client's preferred language
	var language string
	if len(data.Request.Headers.Languages) > 0 {
		raw := data.Request.Headers.Languages[0]
		comma := strings.IndexRune(raw, ',')
		if comma > 0 {
			raw = raw[:comma]
		}
		dash := strings.IndexRune(raw, '-')
		if dash > 0 {
			raw = raw[:dash]
		}
		language = raw
	} else {
		language = "unspecified"
	}
	counter.Languages[language] += 1

	// Parse client's supported encodings
	if len(data.Request.Headers.Encodings) > 0 {
		raw := data.Request.Headers.Encodings[0]
		slice := strings.Split(raw, ",")
		for i := range slice {
			clean := strings.TrimSpace(slice[i])
			counter.Encodings[clean] += 1
		}
	}

	// Parse client's referer if it exists
	if len(data.Request.Headers.Referer) > 0 {
		referer := data.Request.Headers.Referer[0]
		counter.Referers[referer] += 1
	}

	// Parse response content type
	if len(data.Response.ContentType) > 0 {
		raw := data.Response.ContentType[0]
		semicolon := strings.IndexRune(raw, ';')
		if semicolon > 0 {
			raw = raw[:semicolon]
		}
		counter.ContentTypes[raw] += 1
	}

	// Analyze encryption methods
	cipher := tls.CipherSuiteName(data.Request.Encryption.Cipher)
	counter.Ciphers[cipher] += 1
	switch data.Request.Encryption.Version {
	case 0x0300:
		counter.Tls["SSL v3.0"] += 1
	case 0x0301:
		counter.Tls["TLS v1.0"] += 1
	case 0x0302:
		counter.Tls["TLS v1.1"] += 1
	case 0x0303:
		counter.Tls["TLS v1.2"] += 1
	case 0x0304:
		counter.Tls["TLS v1.3"] += 1
	default:
		counter.Tls["Unknown"] += 1
	}

	// Miscellaneous simple counters
	counter.Methods[data.Request.Method] += 1
	counter.Locations[data.Request.Location] += 1
	counter.Protocols[data.Request.Protocol] += 1
	counter.Statuses[data.Status] += 1
	counter.Duration += data.Duration
	counter.Bytes += data.Size

	// Timestamp logic
	if stats.FirstStamp > data.Stamp || stats.FirstStamp == 0 {
		stats.FirstStamp = data.Stamp
	}
	if stats.LastStamp < data.Stamp || stats.LastStamp == 0 {
		stats.LastStamp = data.Stamp
	}
}
