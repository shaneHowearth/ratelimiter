package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	ratelimiter "github.com/shanehowearth/ratelimiter/limiter/internal/ratelimiterservice"
	"github.com/shanehowearth/ratelimiter/limiter/internal/repository/postgres"
)

var rl *ratelimiter.RateLimitService

func main() {
	// Data store
	db := new(postgres.Datastore)

	db.Retry = 1
	var found bool
	db.URI, found = os.LookupEnv("DBURI")
	if !found {
		log.Fatalf("DBURI not set in ENV, cannot continue")
	}

	// limit of queries
	l, found := os.LookupEnv("LIMIT")
	if !found {
		// set a default limit of 100
		log.Print("No LIMIT set in ENV, am defaulting to '100'")
		l = "100"
	}
	limit, err := strconv.Atoi(l)
	if err != nil {
		// Bad LIMIT value in environment
		// Will die here because it's clear a value was meant to be set, but it has a typo
		log.Printf("Bad value set for LIMIT, cannot continue, have error: %v", err)
	}

	// time span limit applies to
	t, found := os.LookupEnv("TIMESPAN")
	if !found {
		// set a default limit of 1 hour
		log.Print("No TIMESPAN set in ENV, am defaulting to '1h'")
		t = "1h"
	}

	timespan, err := time.ParseDuration(t)
	if err != nil {
		// Bad TIMESPAN value in environment
		// Will die here because it's clear a value was meant to be set, but it has a typo
		log.Printf("Bad value set for TIMESPAN, cannot continue, have error: %v", err)
	}

	rl, err = ratelimiter.NewRateLimitService(db, &limit, &timespan)
	if err != nil {
		log.Fatalf("Unable to create a new ratelimitservice with error: %v", err)
	}

	// HTTP server with graceful shutdown
	log.Print("Starting HTTP server")
	portNum, found := os.LookupEnv("PORTNUM")
	if !found {
		log.Printf("No PORTNUM set, am defaulting to '80'")
		portNum = "80"
	}
	http.HandleFunc("/", rateLimitandForward)
	server := &http.Server{Addr: "0.0.0.0:" + portNum}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Panicf("Listen and serve returned error: %v", err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown returned error %v", err)
	}
}

func rateLimitandForward(w http.ResponseWriter, r *http.Request) {
	// The "/" matches anything not handled elsewhere. If it's not the root
	// then report not found.
	log.Print("rateLimitandForward")
	if r.URL.Path != "/" {
		log.Printf("Path not found %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	log.Print("Checking")
	reject, wait, err := rl.CheckReachedLimit(r.RemoteAddr)

	if err != nil {
		log.Fatalf("CheckReachedLimit returned error %v, unable to continue", err)
	}
	if reject {
		log.Printf("Rejecting with vals wait: %v", wait)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		response, jerr := json.Marshal(map[string]string{"message": fmt.Sprintf("Rate limit exceeded. Try again in %f seconds", wait)})

		if jerr != nil {
			log.Printf("Unable to marshal with error %v", jerr)
		}
		_, err = w.Write(response)
		if err != nil {
			// log the error
			log.Printf("writing response generated error: %v", err)
		}
	}
	log.Printf("Remote address: %#+v", r.RemoteAddr)
	// httputil.NewSingleHostReverseProxy(url).ServeHTTP(res, req)
	_, err = io.WriteString(w, "Hello World!")

	if err != nil {
		log.Printf("Error writing out, %v", err)
	}
}
