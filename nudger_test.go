package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func MockNewRelic(requests chan string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ApplicationResponse{}
		b, _ := json.Marshal(response)
		w.Write(b)
		parts := strings.Split(r.URL.String(), "/")
		id := strings.Split(parts[len(parts)-1], ".")[0]
		requests <- id
	}))
	return ts
}

func MockStatusPage(requests chan string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p SPPayload
		body, _ := ioutil.ReadAll(r.Body)
		_ = json.Unmarshal(body, &p)
		requests <- strconv.FormatFloat(p.Data.Value, 'E', -1, 64)
	}))
	return ts
}

func TestNewRelicPolling(t *testing.T) {
	requests := make(chan string)

	// Send a failure after one second. If everything is working, this should not happen.
	go func() {
		time.Sleep(1 * time.Second)
		requests <- "false"
	}()

	// Setup a mock StatusPage that will received requests.
	nr := MockNewRelic(requests)

	// Then make a request
	config := Config{
		NRBaseURL: nr.URL + "/v2/applications/",
	}
	metrics := make(chan Metric)
	app := App{NRAppId: 123456}
	go PollNR(config, app, metrics)

	request := <-requests
	switch request {
	// Test the New Relic API is hit
	case strconv.Itoa(app.NRAppId):
		t.Logf("Received request for app: %d", app.NRAppId)
	case "false":
		t.Fatal("Expected request to New Relic, got nothing after 1 second.")
	default:
		t.Fatalf("Got: '%s'", request)
	}
}

func TestStatusPagePushing(t *testing.T) {
	requests := make(chan string)

	// Send a failure after one second. If everything is working, this should not happen.
	go func() {
		time.Sleep(1 * time.Second)
		requests <- "false"
	}()

	// Setup a mock StatusPage that will received requests.
	sp := MockStatusPage(requests)

	// Set up the dispatcher
	config := Config{
		SPBaseURL: sp.URL + "/v1",
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
		t.Logf("Received dispatched value: %s", strconv.FormatFloat(sample.Value, 'E', -1, 64))
	case "false":
		t.Fatal("Expected dispatch to StatusPage, got nothing after 1 second.")
	default:
		t.Fatalf("Got: '%s'", request)
	}
}
