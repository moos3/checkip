package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tomasen/realip"
)

/*
* This is a very simple checkip and geoip information service
 */

type key int

// GeoIP -
//This is to hold the GeoIP Lookup information that will be returned from our outbound call to freegeoip.net
type GeoIP struct {
	IP            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code"`
	ContinentName string  `json:"continent_name"`
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	RegionCode    string  `json:"region_code"`
	RegionName    string  `json:"region_name"`
	City          string  `json:"city"`
	Zipcode       string  `json:"zipcode"`
	Lat           float32 `json:"latitude"`
	Lon           float32 `json:"longitude"`
	MetroCode     int     `json:"metro_code"`
	AreaCode      int     `json:"area_code"`
}

// TzData -
// This is to hold the json from the tz db database we are looking up the lat and lon to find the timezone
type TzData struct {
	Sunrise     string  `json:"sunrise"`
	Lng         float32 `json:"lng"`
	CountryCode string  `json:"countryCode"`
	GmtOffset   int32   `json:"gmtOffset"`
	RawOffset   int32   `json:"rawOffset"`
	Sunset      string  `json:"sunset"`
	TZID        string  `json:"timezoneId"`
	DstOffset   int32   `json:"dstOffset"`
	CountryName string  `json:"countryName"`
	Time        string  `json:"time"`
	Lat         float32 `json:"lat"`
}

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	healthy    int32
	geo        GeoIP
	body       []byte
	// api key holder for the flag for ipstack
	ipStackAPIKey string
	// Timezone data holder
	tzData TzData
)

func main() {
	// Default to port 5000 on localhost
	flag.StringVar(&listenAddr, "listen-addr", ":3000", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(routes())),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Listen for CTRL+C or kill and start shutting down the app without
	// disconnecting people by not taking any new requests. ("Graceful Shutdown")
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

// Setup all your routes
func routes() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/health", healthHandler)
	return router
}

// Shows how to use templates with template functions and data
func indexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var indexHTML = `<html>
   <head><title>Current IP Check</title></head>
   <body>Current IP Address: {{ .IP }}<br/> TimeZone: {{ .TZ }}<br/></body></html>`

	ip := realip.FromRequest(r)
	geo := lookupGeoIP(ip)
	tz := lookupGeoTz(geo.Lat, geo.Lon)

	// Anonymous struct to hold template data
	data := struct {
		Country string
		City    string
		Region  string
		Lat     float32
		Lon     float32
		IP      string
		TZ      string
	}{
		IP:      ip,
		City:    geo.City,
		Country: geo.CountryName,
		Region:  geo.RegionName,
		Lat:     geo.Lat,
		Lon:     geo.Lon,
		TZ:      tz,
	}

	tmpl, err := template.New("index").Parse(indexHTML) // IRL it would be .ParseFiles("templates/index.tpl")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		fmt.Println(err)
	}
}

// lookupGeoIP this function looks up the ip address
// returns a geo object with the information on the IP address
func lookupGeoIP(a string) GeoIP {
	// Use ipstack.com to get a JSON response
	// move this to a config file.
	ipStackAPIKey = "e6b1ee21e625ce82cf20a96844308d6e"
	requestURL := "http://api.ipstack.com/" + a + "?access_key=" + ipStackAPIKey + "&output=json"
	response, err := http.Get(requestURL)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte to a GeoIP struct
	err = json.Unmarshal(body, &geo)
	if err != nil {
		fmt.Println(err)
	}

	return geo
}

// lookupGeoTZ this function looks up the ip address
// returns a geo object with the information on the IP address
func lookupGeoTz(lat float32, lon float32) string {

	// Use ipstack.com to get a JSON response
	requestURL := "http://api.geonames.org/timezoneJSON?lat=" + fmt.Sprintf("%f", lat) + "&lng=" + fmt.Sprintf("%f", lon) + "&username=moos3"
	response, err := http.Get(requestURL)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte to a GeoIP struct
	err = json.Unmarshal(body, &tzData)
	if err != nil {
		fmt.Println(err)
	}

	return tzData.TZID
}

// Prevent Content-Type sniffing
func forceTextHandler(w http.ResponseWriter, r *http.Request) {
	// https://stackoverflow.com/questions/18337630/what-is-x-content-type-options-nosniff
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"status\":\"ok\"}")
}

// Report server status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

// logging just a simple logging handler
// this generates a basic access log entry
func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// tracing for debuging a access log entry to a given request
func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
