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
	LogDirectory  string              `json:"logDirectory"`
	LogSizeBytes  int64               `json:"logSizeBytes"`
	ParseDuration float64             `json:"parseDurationSeconds"`
	FirstStamp    int64               `json:"firstStampUnix"`
	LastStamp     int64               `json:"lastStampUnix"`
	Counters      map[string]*counter `json:"counters"`
}

func newStatistics() *statistics {
	return &statistics{
		Counters: map[string]*counter{},
	}
}

type counter struct {
	Total   hits            `json:"total"`
	Hourly  map[int64]*hits `json:"hourly"`
	Devices struct {
		Mobile  int `json:"mobile"`
		Bot     int `json:"bot"`
		Other   int `json:"other"`
		Tablet  int `json:"tablet"`
		Desktop int `json:"desktop"`
	} `json:"visitorDevice"`
	Browsers        map[string]int            `json:"visitorBrowsers"`
	Systems         map[string]int            `json:"visitorSystems"`
	Languages       map[string]int            `json:"visitorPreferredLanguages"`
	Countries       map[string]int            `json:"visitorCountries"`
	EncodingSupport map[string]int            `json:"visitorEncodingSupport"`
	Protocols       map[string]int            `json:"requestProtocols"`
	Methods         map[string]int            `json:"requestMethods"`
	Crypto          map[string]map[string]int `json:"requestCrypto"`
	ContentTypes    map[string]int            `json:"responseContentTypes"`
	Locations       map[string]*statusCounter `json:"requestLocationResponses"`
}

func newCounter() *counter {
	return &counter{
		Total:           hits{Observed: map[visitor]bool{}},
		Hourly:          map[int64]*hits{},
		Browsers:        map[string]int{},
		Systems:         map[string]int{},
		Languages:       map[string]int{},
		Countries:       map[string]int{},
		EncodingSupport: map[string]int{},
		Protocols:       map[string]int{},
		Methods:         map[string]int{},
		Crypto:          map[string]map[string]int{},
		ContentTypes:    map[string]int{},
		Locations:       map[string]*statusCounter{},
	}
}

type hits struct {
	Requests  int              `json:"requests"`
	Latency   float64          `json:"totalLatency"`
	SentBytes int              `json:"sentBytes"`
	Visitors  int              `json:"visitors"`
	Observed  map[visitor]bool `json:"-"`
}

func newHits() *hits {
	return &hits{
		Observed: map[visitor]bool{},
	}
}

type visitor struct {
	IP           string
	RawUserAgent string
}

type statusCounter struct {
	Informational int `json:"informational"`
	Successful    int `json:"successful"`
	Redirection   int `json:"redirection"`
	ClientError   int `json:"clientError"`
	ServerError   int `json:"serverError"`
	Cancelled     int `json:"cancelled"`
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

// Convert a unix timestamp to the hour it is in
func roundUnixDownToHour(unix float64) int64 {
	seconds := int64(unix)
	rounded := seconds - (seconds % (60 * 60))
	return rounded
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
			case "deflate":
				clean = append(clean, "Deflate")
			case "br":
				clean = append(clean, "Brotli")
			case "gzip":
				clean = append(clean, "Gzip")
			case "snappy":
				clean = append(clean, "Snappy")
			case "sdch":
				clean = append(clean, "SDCH")
			default:
				if enc != "identity" && enc != "utf-8" {
					clean = append(clean, enc)
				}
			}
		}
		return clean
	}
	return []string{}
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
	host := stripPortSuffix(entry.Request.Host)
	counter := stats.Counters[host]
	if counter == nil {
		counter = newCounter()
		stats.Counters[host] = counter
	}

	// Add general stats
	counter.Total.Requests++
	counter.Total.SentBytes += entry.Size
	counter.Total.Latency += entry.Duration

	// Check if the visitor has not been seen yet
	ip := stripPortSuffix(entry.Request.Address)
	userAgent := getRawUserAgent(entry.Request.Headers.UserAgent)
	uniqueVisitor := visitor{ip, userAgent}
	if !counter.Total.Observed[uniqueVisitor] {
		uaInfo := ua.Parse(userAgent)
		if uaInfo.Bot {
			counter.Devices.Bot++
		} else if uaInfo.Tablet {
			counter.Devices.Tablet++
		} else if uaInfo.Mobile {
			counter.Devices.Mobile++
		} else if uaInfo.Desktop {
			counter.Devices.Desktop++
		} else {
			counter.Devices.Other++
		}
		browser := uaInfo.Name
		if browser == "" {
			browser = "Unknown"
		}
		counter.Browsers[browser]++

		os := uaInfo.OS
		if os == "" {
			os = "Unknown"
		}
		counter.Systems[os]++

		prefLanguage := getPreferredLanguage(entry.Request.Headers.Languages)
		counter.Languages[prefLanguage]++

		supEncodings := getSupportedEncodings(entry.Request.Headers.Encodings)
		for _, enc := range supEncodings {
			counter.EncodingSupport[enc]++
		}

		counter.Total.Visitors++
		counter.Total.Observed[uniqueVisitor] = true
	}

	// Get stats counter corresponding with the timestamp
	hour := roundUnixDownToHour(entry.Stamp)
	hourly := counter.Hourly[hour]
	if hourly == nil {
		hourly = newHits()
		counter.Hourly[hour] = hourly
	}

	// Add general stats to the hourly counter
	hourly.Requests++
	hourly.SentBytes += entry.Size
	hourly.Latency += entry.Duration
	if !hourly.Observed[uniqueVisitor] {
		hourly.Visitors++
		hourly.Observed[uniqueVisitor] = true
	}

	// Add crypto protocol and cipher stats
	cipher := tls.CipherSuiteName(entry.Request.Encryption.Cipher)
	protocol := getProtocolFromVersion(int(entry.Request.Encryption.Version))
	cipherCounter := counter.Crypto[protocol]
	if cipherCounter == nil {
		cipherCounter = map[string]int{}
		counter.Crypto[protocol] = cipherCounter
	}
	cipherCounter[cipher]++

	// Add content type stats
	contentType := getContentType(entry.Response.ContentType)
	counter.ContentTypes[contentType]++

	// Add location stats with the status code
	locStatusCounter := counter.Locations[entry.Request.Location]
	if locStatusCounter == nil {
		locStatusCounter = &statusCounter{}
		counter.Locations[entry.Request.Location] = locStatusCounter
	}
	switch int(entry.Status / 100) {
	case 0:
		locStatusCounter.Cancelled++
	case 1:
		locStatusCounter.Informational++
	case 2:
		locStatusCounter.Successful++
	case 3:
		locStatusCounter.Redirection++
	case 4:
		locStatusCounter.ClientError++
	case 5:
		locStatusCounter.ServerError++
	}

	// Add HTTP method stats
	counter.Methods[entry.Request.Method]++

	// Add HTTP protocol stats
	counter.Protocols[entry.Request.Protocol]++

	// Change timestamp if current one lies outside the current boundaries
	stamp := int64(entry.Stamp + 0.5)
	if stats.FirstStamp > stamp || stats.FirstStamp == 0 {
		stats.FirstStamp = stamp
	}
	if stats.LastStamp < stamp || stats.LastStamp == 0 {
		stats.LastStamp = stamp
	}

	return nil
}
