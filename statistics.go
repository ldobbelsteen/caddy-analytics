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
	LogDirectory         string              `json:"logDirectory"`
	LogSizeBytes         int64               `json:"logSizeBytes"`
	ParseDurationSeconds float64             `json:"parseDurationSeconds"`
	FirstStampUnix       int64               `json:"firstStampUnix"`
	LastStampUnix        int64               `json:"lastStampUnix"`
	HostCounters         map[string]*counter `json:"hostCounters"`
}

func newStatistics() *statistics {
	return &statistics{
		HostCounters: map[string]*counter{},
	}
}

type counter struct {
	TotalMetrics         metrics            `json:"totalMetrics"`
	HourlyMetrics        map[int64]*metrics `json:"hourlyMetrics"`
	NonBotVisitorDevices struct {
		Mobile  int `json:"mobile"`
		Other   int `json:"other"`
		Tablet  int `json:"tablet"`
		Desktop int `json:"desktop"`
	} `json:"nonBotVisitorDevices"`
	NonBotVisitorBrowsers      map[string]int            `json:"nonBotVisitorBrowsers"`
	NonBotVisitorSystems       map[string]int            `json:"nonBotVisitorSystems"`
	NonBotVisitorPrefLanguages map[string]int            `json:"nonBotVisitorPrefLanguages"`
	NonBotVisitorCountries     map[string]int            `json:"nonBotVisitorCountries"`
	NonBotVisitorEncodings     map[string]int            `json:"nonBotVisitorEncodings"`
	BotVisitors                map[string]int            `json:"botVisitors"`
	RequestsByProtocol         map[string]int            `json:"requestsByProtocol"`
	RequestsByMethod           map[string]int            `json:"requestsByMethod"`
	RequestsByCrypto           map[string]map[string]int `json:"requestsByCrypto"`
	ResponseByContent          map[string]int            `json:"responseByContent"`
	ResponseByLocation         map[string]*statusCounter `json:"responseByLocation"`
}

func newCounter() *counter {
	return &counter{
		TotalMetrics:               metrics{ObservedBots: map[visitor]bool{}, ObservedNonBots: map[visitor]bool{}},
		HourlyMetrics:              map[int64]*metrics{},
		NonBotVisitorBrowsers:      map[string]int{},
		NonBotVisitorSystems:       map[string]int{},
		NonBotVisitorPrefLanguages: map[string]int{},
		NonBotVisitorCountries:     map[string]int{},
		NonBotVisitorEncodings:     map[string]int{},
		BotVisitors:                map[string]int{},
		RequestsByProtocol:         map[string]int{},
		RequestsByMethod:           map[string]int{},
		RequestsByCrypto:           map[string]map[string]int{},
		ResponseByContent:          map[string]int{},
		ResponseByLocation:         map[string]*statusCounter{},
	}
}

type metrics struct {
	Requests        int              `json:"requests"`
	Latency         float64          `json:"latency"`
	SentBytes       int              `json:"sentBytes"`
	NonBotVisitors  int              `json:"nonBotVisitors"`
	ObservedNonBots map[visitor]bool `json:"-"`
	ObservedBots    map[visitor]bool `json:"-"`
}

func newMetrics() *metrics {
	return &metrics{
		ObservedNonBots: map[visitor]bool{},
		ObservedBots:    map[visitor]bool{},
	}
}

type visitor struct {
	IP           string
	RawUserAgent string
}

type statusCounter struct {
	ZeroXX  int `json:"0xx"`
	OneXX   int `json:"1xx"`
	TwoXX   int `json:"2xx"`
	ThreeXX int `json:"3xx"`
	FourXX  int `json:"4xx"`
	FiveXX  int `json:"5xx"`
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
	counter := stats.HostCounters[host]
	if counter == nil {
		counter = newCounter()
		stats.HostCounters[host] = counter
	}

	// Add general stats
	counter.TotalMetrics.Requests++
	counter.TotalMetrics.SentBytes += entry.Size
	counter.TotalMetrics.Latency += entry.Duration

	// Check if the visitor has not been seen yet
	ip := stripPortSuffix(entry.Request.Address)
	userAgent := getRawUserAgent(entry.Request.Headers.UserAgent)
	uniqueVisitor := visitor{ip, userAgent}
	if !counter.TotalMetrics.ObservedNonBots[uniqueVisitor] && !counter.TotalMetrics.ObservedBots[uniqueVisitor] {

		// Get visitor's browser name
		uaInfo := ua.Parse(userAgent)
		browser := uaInfo.Name
		if browser == "" {
			browser = "Unknown"
		}

		// Check if visitor is a bot and add stats if not
		if uaInfo.Bot {
			counter.BotVisitors[browser]++
			counter.TotalMetrics.ObservedBots[uniqueVisitor] = true
		} else {
			counter.NonBotVisitorBrowsers[browser]++

			// Get visitor device type
			if uaInfo.Tablet {
				counter.NonBotVisitorDevices.Tablet++
			} else if uaInfo.Mobile {
				counter.NonBotVisitorDevices.Mobile++
			} else if uaInfo.Desktop {
				counter.NonBotVisitorDevices.Desktop++
			} else {
				counter.NonBotVisitorDevices.Other++
			}

			// Get visitor operating system
			os := uaInfo.OS
			if os == "" {
				os = "Unknown"
			}
			counter.NonBotVisitorSystems[os]++

			// Get the main/preferred language of the visitor
			prefLanguage := getPreferredLanguage(entry.Request.Headers.Languages)
			counter.NonBotVisitorPrefLanguages[prefLanguage]++

			// Get all the encodings the visitor supports
			supEncodings := getSupportedEncodings(entry.Request.Headers.Encodings)
			for _, enc := range supEncodings {
				counter.NonBotVisitorEncodings[enc]++
			}

			counter.TotalMetrics.NonBotVisitors++
			counter.TotalMetrics.ObservedNonBots[uniqueVisitor] = true
		}
	}

	// Get stats counter corresponding with the hour of the timestamp
	hour := roundUnixDownToHour(entry.Stamp)
	hourly := counter.HourlyMetrics[hour]
	if hourly == nil {
		hourly = newMetrics()
		counter.HourlyMetrics[hour] = hourly
	}

	// Add general stats to the hourly counter
	hourly.Requests++
	hourly.SentBytes += entry.Size
	hourly.Latency += entry.Duration
	if !hourly.ObservedNonBots[uniqueVisitor] && !hourly.ObservedBots[uniqueVisitor] {
		uaInfo := ua.Parse(userAgent)
		if !uaInfo.Bot {
			hourly.NonBotVisitors++
			hourly.ObservedNonBots[uniqueVisitor] = true
		} else {
			hourly.ObservedBots[uniqueVisitor] = true
		}
	}

	// Add crypto protocol and cipher stats
	cipher := tls.CipherSuiteName(entry.Request.Encryption.Cipher)
	protocol := getProtocolFromVersion(int(entry.Request.Encryption.Version))
	cipherCounter := counter.RequestsByCrypto[protocol]
	if cipherCounter == nil {
		cipherCounter = map[string]int{}
		counter.RequestsByCrypto[protocol] = cipherCounter
	}
	cipherCounter[cipher]++

	// Add content type stats
	contentType := getContentType(entry.Response.ContentType)
	counter.ResponseByContent[contentType]++

	// Add location stats with the status code
	locStatusCounter := counter.ResponseByLocation[entry.Request.Location]
	if locStatusCounter == nil {
		locStatusCounter = &statusCounter{}
		counter.ResponseByLocation[entry.Request.Location] = locStatusCounter
	}
	switch int(entry.Status / 100) {
	case 0:
		locStatusCounter.ZeroXX++
	case 1:
		locStatusCounter.OneXX++
	case 2:
		locStatusCounter.TwoXX++
	case 3:
		locStatusCounter.ThreeXX++
	case 4:
		locStatusCounter.FourXX++
	case 5:
		locStatusCounter.FiveXX++
	}

	// Add HTTP method stats
	counter.RequestsByMethod[entry.Request.Method]++

	// Add HTTP protocol stats
	counter.RequestsByProtocol[entry.Request.Protocol]++

	// Change timestamp if current one lies outside the current boundaries
	stamp := int64(entry.Stamp)
	if stats.FirstStampUnix > stamp || stats.FirstStampUnix == 0 {
		stats.FirstStampUnix = stamp
	}
	if stats.LastStampUnix < stamp || stats.LastStampUnix == 0 {
		stats.LastStampUnix = stamp
	}

	return nil
}
