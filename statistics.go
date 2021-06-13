package main

import (
	"crypto/tls"
	"net"
	"strings"

	ua "github.com/mileusna/useragent"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type statistics struct {
	Directory     string           `json:"directory"`
	SizeBytes     int64            `json:"sizeBytes"`
	ParseDuration float64          `json:"parseDuration"`
	FirstStamp    int64            `json:"firstStamp"`
	LastStamp     int64            `json:"lastStamp"`
	Hosts         map[string]*host `json:"hosts"`
}

type hour uint64

type host struct {
	Total    metrics           `json:"total"`
	Hourly   map[hour]*metrics `json:"hourly"`
	Visitors visitors          `json:"visitors"`
	Requests requests          `json:"requests"`
}

type metrics struct {
	Requests      int64                  `json:"requests"`
	Latency       float64                `json:"latency"`
	SentBytes     int64                  `json:"sentBytes"`
	Visitors      int64                  `json:"visitors"`
	ObservedUsers map[uniqueVisitor]bool `json:"-"`
	ObservedBots  map[uniqueVisitor]bool `json:"-"`
}

type uniqueVisitor struct {
	IPAddress    string
	RawUserAgent string
}

type visitors struct {
	Bots      map[string]int64 `json:"bots"`
	Devices   map[string]int64 `json:"devices"`
	Browsers  map[string]int64 `json:"browsers"`
	Systems   map[string]int64 `json:"systems"`
	Languages map[string]int64 `json:"languages"`
	Countries map[string]int64 `json:"countries"`
	Encodings map[string]int64 `json:"encodings"`
}

type requests struct {
	Protocols map[string]int64     `json:"protocols"`
	Methods   map[string]int64     `json:"methods"`
	Cryptos   map[string]int64     `json:"cryptos"`
	Ciphers   map[string]int64     `json:"ciphers"`
	Contents  map[string]int64     `json:"contents"`
	Locations map[string]*statuses `json:"locations"`
}

type statuses [6]int64

func newStatistics() *statistics {
	return &statistics{
		Hosts: make(map[string]*host),
	}
}

func newHost() *host {
	return &host{
		Total:    *newMetrics(),
		Hourly:   make(map[hour]*metrics),
		Visitors: *newVisitors(),
		Requests: *newRequests(),
	}
}

func newMetrics() *metrics {
	return &metrics{
		ObservedUsers: make(map[uniqueVisitor]bool),
		ObservedBots:  make(map[uniqueVisitor]bool),
	}
}

func newVisitors() *visitors {
	return &visitors{
		Bots:      make(map[string]int64),
		Devices:   make(map[string]int64),
		Browsers:  make(map[string]int64),
		Systems:   make(map[string]int64),
		Languages: make(map[string]int64),
		Countries: make(map[string]int64),
		Encodings: make(map[string]int64),
	}
}

func newRequests() *requests {
	return &requests{
		Protocols: make(map[string]int64),
		Methods:   make(map[string]int64),
		Cryptos:   make(map[string]int64),
		Ciphers:   make(map[string]int64),
		Contents:  make(map[string]int64),
		Locations: make(map[string]*statuses),
	}
}

func newStatuses() *statuses {
	var array [6]int64
	statuses := statuses(array)
	return &statuses
}

// Convert a unix timestamp to the hour it is in
func unixToHour(unix float64) hour {
	seconds := hour(unix)
	rounded := seconds - (seconds % (60 * 60))
	return rounded
}

// Remove the port suffix from a string if there is one
func stripPortSuffix(str string) string {
	host, _, err := net.SplitHostPort(str)
	if err != nil {
		return str
	}
	return host
}

// Remove the http(s) prefix from a string if there is one
func stripHTTPPrefix(str string) string {
	if strings.HasPrefix(str, "http://") {
		return str[7:]
	} else if strings.HasPrefix(str, "https://") {
		return str[8:]
	} else {
		return str
	}
}

// Get the first user agent from a slice of user agents
func getRawUserAgent(slc []string) string {
	if len(slc) > 0 {
		return slc[0]
	}
	return ""
}

// Get preferred language from a slice of Accept-Language headers
func getPreferredLanguage(slc []string) string {
	if len(slc) > 0 {
		raw := slc[0]
		comma := strings.IndexRune(raw, ',')
		if comma > 0 {
			raw = raw[:comma]
		}
		dash := strings.IndexRune(raw, '-')
		if dash > 0 {
			raw = raw[:dash]
		}
		tag, err := language.Parse(raw)
		if err != nil {
			return "Unknown"
		}
		lang := display.English.Tags().Name(tag)
		return lang
	}
	return "None"
}

// Get supported encoding/compression schemes from Accept-Encodings header
func getSupportedEncodings(slc []string) []string {
	if len(slc) > 0 {
		slc := strings.Split(slc[0], ",")
		var clean []string
		for _, enc := range slc {
			enc = strings.TrimSpace(enc)
			switch enc {
			case "identity": // Do nothing, everyone supports this
			case "utf-8": // Do nothing, everyone supports this
			case "gzip":
				clean = append(clean, "Gzip")
			case "deflate":
				clean = append(clean, "Deflate")
			case "br":
				clean = append(clean, "Brotli")
			case "snappy":
				clean = append(clean, "Snappy")
			case "sdch":
				clean = append(clean, "SDCH")
			default:
				clean = append(clean, enc)
			}
		}
		return clean
	}
	return make([]string, 0)
}

// Get content type from response header
func getContentType(slc []string) string {
	if len(slc) > 0 {
		raw := slc[0]
		semicolon := strings.IndexRune(raw, ';')
		if semicolon > 0 {
			raw = raw[:semicolon]
		}
		return raw
	}
	return "none"
}

// Convert crypto/tls protocol version to human-readable string
func getProtocolFromVersion(version int) string {
	switch version {
	case 0x0300:
		return "SSL v3.0"
	case 0x0301:
		return "TLS v1.0"
	case 0x0302:
		return "TLS v1.1"
	case 0x0303:
		return "TLS v1.2"
	case 0x0304:
		return "TLS v1.3"
	default:
		return "Unknown"
	}
}

// Add a log entry to an instance of statistics
func addToStats(entry *logEntry, stats *statistics) error {

	// Get counter corresponding to host
	hostname := stripPortSuffix(entry.Request.Host)
	host := stats.Hosts[hostname]
	if host == nil {
		host = newHost()
		stats.Hosts[hostname] = host
	}

	// Add general stats
	host.Total.Requests++
	host.Total.SentBytes += entry.Size
	host.Total.Latency += entry.Duration

	// Check if the visitor has not been seen yet
	ipAddress := stripPortSuffix(entry.Request.Address)
	userAgent := getRawUserAgent(entry.Request.Headers.UserAgent)
	uniqueVisitor := uniqueVisitor{ipAddress, userAgent}
	neverObservedBefore := !host.Total.ObservedUsers[uniqueVisitor] && !host.Total.ObservedBots[uniqueVisitor]
	if neverObservedBefore {
		uaInfo := ua.Parse(userAgent)

		// Get visitor's browser name
		browser := uaInfo.Name
		if browser == "" {
			browser = "Unknown"
		}

		// Check if visitor is a bot and add stats if not
		if uaInfo.Bot {
			host.Visitors.Bots[browser]++
			host.Total.ObservedBots[uniqueVisitor] = true
		} else {
			host.Visitors.Browsers[browser]++

			// Get visitor device type
			if uaInfo.Tablet {
				host.Visitors.Devices["Tablet"]++
			} else if uaInfo.Mobile {
				host.Visitors.Devices["Mobile"]++
			} else if uaInfo.Desktop {
				host.Visitors.Devices["Desktop"]++
			} else {
				host.Visitors.Devices["Other"]++
			}

			// Get visitor operating system
			os := uaInfo.OS
			if os == "" {
				os = "Unknown"
			}
			host.Visitors.Systems[os]++

			// Get the main/preferred language of the visitor
			language := getPreferredLanguage(entry.Request.Headers.Languages)
			host.Visitors.Languages[language]++

			// Get all the encodings the visitor supports
			encodings := getSupportedEncodings(entry.Request.Headers.Encodings)
			for _, enc := range encodings {
				host.Visitors.Encodings[enc]++
			}

			host.Total.Visitors++
			host.Total.ObservedUsers[uniqueVisitor] = true
		}
	}

	// Get stats counter corresponding with the hour of the timestamp
	hourStamp := unixToHour(entry.Stamp)
	hour := host.Hourly[hourStamp]
	if hour == nil {
		hour = newMetrics()
		host.Hourly[hourStamp] = hour
	}

	// Add general stats to the hourly counter
	hour.Requests++
	hour.SentBytes += entry.Size
	hour.Latency += entry.Duration
	neverObservedInHour := !hour.ObservedUsers[uniqueVisitor] && !hour.ObservedBots[uniqueVisitor]
	if neverObservedInHour {
		uaInfo := ua.Parse(userAgent)
		if !uaInfo.Bot {
			hour.Visitors++
			hour.ObservedUsers[uniqueVisitor] = true
		} else {
			hour.ObservedBots[uniqueVisitor] = true
		}
	}

	// Add crypto protocol and cipher stats
	cipher := tls.CipherSuiteName(entry.Request.Encryption.Cipher)
	host.Requests.Ciphers[cipher]++
	crypto := getProtocolFromVersion(int(entry.Request.Encryption.Version))
	host.Requests.Cryptos[crypto]++

	// Add content type stats
	content := getContentType(entry.Response.ContentType)
	host.Requests.Contents[content]++

	// Add location stats with the status code
	location := entry.Request.Location
	status := int(entry.Status / 100)
	statuses := host.Requests.Locations[location]
	if statuses == nil {
		statuses = newStatuses()
		host.Requests.Locations[location] = statuses
	}
	statuses[status]++

	// Add HTTP method stats
	host.Requests.Methods[entry.Request.Method]++

	// Add HTTP protocol stats
	host.Requests.Protocols[entry.Request.Protocol]++

	// Change timestamp if current one lies outside the current boundaries
	stamp := int64(entry.Stamp)
	if stats.FirstStamp > stamp || stats.FirstStamp == 0 {
		stats.FirstStamp = stamp
	}
	if stats.LastStamp < stamp || stats.LastStamp == 0 {
		stats.LastStamp = stamp
	}

	return nil
}
