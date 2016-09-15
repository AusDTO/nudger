package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func MockNewRelic(bind string) {
}

func MockStatusPage(bind string, requests chan string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var p SPPayload
		body, _ := ioutil.ReadAll(r.Body)
		_ = json.Unmarshal(body, &p)
		requests <- strconv.FormatFloat(p.Data.Value, 'E', -1, 64)
	})
	log.Fatal(http.ListenAndServe(bind, nil))
}

func TestNewRelicPolling(t *testing.T) {
}

func TestStatusPagePushing(t *testing.T) {
	requests := make(chan string)

	// Send a failure after one second. If everything is working, this should not happen.
	go func() {
		time.Sleep(1 * time.Second)
		requests <- "false"
	}()

	// Setup a mock StatusPage that will received requests.
	go MockStatusPage("127.0.0.1:42224", requests)

	// Set up the dispatcher
	config := Config{
		SPBaseURL: "http://127.0.0.1:42224",
	}
	metrics := make(chan Metric)
	go Dispatch(config, metrics)

	// Then dispatch a single metric
	sample := Metric{
		SPApiKey:   "hello",
		SPPageId:   "world",
		SPMetricId: "true",
		Value:      10.123,
	}
	metrics <- sample

	request := <-requests
	switch request {
	// Test the same metric value is received
	case strconv.FormatFloat(sample.Value, 'E', -1, 64):
		t.Logf("Received: %s", strconv.FormatFloat(sample.Value, 'E', -1, 64))
	case "false":
		t.Fatal("Expected dispatch to StatusPage, got nothing after 1 second.")
	default:
		t.Fatalf("Got: '%s'", request)
	}
}
