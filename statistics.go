package main

import (
	"crypto/tls"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mssola/user_agent"
)

type Statistics struct {
	LogDirectory  string              `json:"logDirectory"`
	LogSizeBytes  int64               `json:"logSizeBytes"`
	ParseDuration float64             `json:"parseDurationSeconds"`
	FirstStamp    float64             `json:"firstStampUnix"`
	LastStamp     float64             `json:"lastStampUnix"`
	Counters      map[string]*Counter `json:"counters"`
}

func newStatistics() *Statistics {
	return &Statistics{
		Counters: map[string]*Counter{},
	}
}

type Counter struct {
	Total   Hits           `json:"total"`
	Hourly  map[Hour]*Hits `json:"hourly"`
	Devices struct {
		Mobile int `json:"mobile"`
		Bot    int `json:"bot"`
		Other  int `json:"desktop"`
	} `json:"visitorDevice"`
	Browsers        map[string]int            `json:"visitorBrowsers"`
	Systems         map[string]int            `json:"visitorSystems"`
	Languages       map[string]int            `json:"visitorPrefLanguages"`
	Countries       map[string]int            `json:"visitorCountries"`
	EncodingSupport map[string]int            `json:"visitorEncodingSupport"`
	Protocols       map[string]int            `json:"requestProtocols"`
	Methods         map[string]int            `json:"requestMethods"`
	Crypto          map[string]map[string]int `json:"requestCrypto"`
	ContentTypes    map[string]int            `json:"responseContentTypes"`
	Locations       map[string]*StatusCounter `json:"requestLocationResponses"`
}

func newCounter() *Counter {
	return &Counter{
		Total:           Hits{Observed: map[Visitor]bool{}},
		Hourly:          map[Hour]*Hits{},
		Browsers:        map[string]int{},
		Systems:         map[string]int{},
		Languages:       map[string]int{},
		Countries:       map[string]int{},
		EncodingSupport: map[string]int{},
		Protocols:       map[string]int{},
		Methods:         map[string]int{},
		Crypto:          map[string]map[string]int{},
		ContentTypes:    map[string]int{},
		Locations:       map[string]*StatusCounter{},
	}
}

type Hits struct {
	Requests  int              `json:"requests"`
	Latency   float64          `json:"totalLatency"`
	SentBytes int              `json:"sentBytes"`
	Visitors  int              `json:"visitors"`
	Observed  map[Visitor]bool `json:"-"`
}

func newHits() *Hits {
	return &Hits{
		Observed: map[Visitor]bool{},
	}
}

type Hour struct {
	Year        int
	MonthOfYear int
	DayOfMonth  int
	HourOfDay   int
}

func (hour Hour) MarshalText() ([]byte, error) {
	year := strconv.Itoa(hour.Year)
	monthOfYear := strconv.Itoa(hour.MonthOfYear)
	dayOfMonth := strconv.Itoa(hour.DayOfMonth)
	hourOfDay := strconv.Itoa(hour.HourOfDay)
	str := year + "/" + monthOfYear + "/" + dayOfMonth + ":" + hourOfDay
	return []byte(str), nil
}

type Visitor struct {
	IP           string
	RawUserAgent string
}

type StatusCounter struct {
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
	} else {
		return host
	}
}

// Remove the http(s) prefix from a string if there is one
func stripHttpPrefix(str string) string {
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
	} else {
		return ""
	}
}

// Convert a unix timestamp to the hour it is in
func unixToHour(unix float64) Hour {
	seconds, decimals := math.Modf(unix)
	time := time.Unix(int64(seconds), int64(decimals*(1e9)))
	return Hour{
		Year:        time.Year(),
		MonthOfYear: int(time.Month()),
		DayOfMonth:  time.Day(),
		HourOfDay:   time.Hour(),
	}
}

// Get preferred language from an Accept-Language header
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
		return raw
	} else {
		return "none"
	}
}

// Get supported encoding/compression schemes from Accept-Encodings header
func getSupportedEncodings(slc []string) []string {
	if len(slc) > 0 {
		slc := strings.Split(slc[0], ",")
		for i := range slc {
			slc[i] = strings.TrimSpace(slc[i])
		}
		return slc
	} else {
		return []string{}
	}
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
	} else {
		return "none"
	}
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
func addToStats(entry *LogEntry, stats *Statistics) error {

	// Get counter corresponding to host
	host := stripPortSuffix(entry.Request.Host)
	counter := stats.Counters[host]
	if counter == nil {
		counter = newCounter()
		stats.Counters[host] = counter
	}

	// Add general stats
	counter.Total.Requests += 1
	counter.Total.SentBytes += entry.Size
	counter.Total.Latency += entry.Duration

	// Check if the visitor has not been seen yet
	ip := stripPortSuffix(entry.Request.Address)
	ua := getRawUserAgent(entry.Request.Headers.UserAgent)
	visitor := Visitor{ip, ua}
	if !counter.Total.Observed[visitor] {
		uaInfo := user_agent.New(ua)
		if uaInfo.Bot() {
			counter.Devices.Bot += 1
		} else if uaInfo.Mobile() {
			counter.Devices.Mobile += 1
		} else {
			counter.Devices.Other += 1
		}
		browser, _ := uaInfo.Browser()
		if browser == "" {
			browser = "Unknown"
		}
		counter.Browsers[browser] += 1

		os := uaInfo.OS()
		if os == "" {
			os = "Unknown"
		}
		counter.Systems[os] += 1

		prefLanguage := getPreferredLanguage(entry.Request.Headers.Languages)
		counter.Languages[prefLanguage] += 1

		supEncodings := getSupportedEncodings(entry.Request.Headers.Encodings)
		for _, enc := range supEncodings {
			counter.EncodingSupport[enc] += 1
		}

		counter.Total.Visitors += 1
		counter.Total.Observed[visitor] = true
	}

	// Get stats counter corresponding with the timestamp
	hour := unixToHour(entry.Stamp)
	hourly := counter.Hourly[hour]
	if hourly == nil {
		hourly = newHits()
		counter.Hourly[hour] = hourly
	}

	// Add general stats to the hourly counter
	hourly.Requests += 1
	hourly.SentBytes += entry.Size
	hourly.Latency += entry.Duration
	if !hourly.Observed[visitor] {
		hourly.Visitors += 1
		hourly.Observed[visitor] = true
	}

	// Add crypto protocol and cipher stats
	cipher := tls.CipherSuiteName(entry.Request.Encryption.Cipher)
	protocol := getProtocolFromVersion(int(entry.Request.Encryption.Version))
	cipherCounter := counter.Crypto[protocol]
	if cipherCounter == nil {
		cipherCounter = map[string]int{}
		counter.Crypto[protocol] = cipherCounter
	}
	cipherCounter[cipher] += 1

	// Add content type stats
	contentType := getContentType(entry.Response.ContentType)
	counter.ContentTypes[contentType] += 1

	// Add location stats with the status code
	statusCounter := counter.Locations[entry.Request.Location]
	if statusCounter == nil {
		statusCounter = &StatusCounter{}
		counter.Locations[entry.Request.Location] = statusCounter
	}
	switch int(entry.Status / 100) {
	case 0:
		statusCounter.Cancelled += 1
	case 1:
		statusCounter.Informational += 1
	case 2:
		statusCounter.Successful += 1
	case 3:
		statusCounter.Redirection += 1
	case 4:
		statusCounter.ClientError += 1
	case 5:
		statusCounter.ServerError += 1
	}

	// Add HTTP method stats
	counter.Methods[entry.Request.Method] += 1

	// Add HTTP protocol stats
	counter.Protocols[entry.Request.Protocol] += 1

	// Change timestamp if current one lies outside the current boundaries
	if stats.FirstStamp > entry.Stamp || stats.FirstStamp == 0 {
		stats.FirstStamp = entry.Stamp
	}
	if stats.LastStamp < entry.Stamp || stats.LastStamp == 0 {
		stats.LastStamp = entry.Stamp
	}

	return nil
}
