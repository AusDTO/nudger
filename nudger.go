package main

import (
	"bytes"
	"encoding/json"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Timeout    time.Duration
	Interval   time.Duration
	ConfigPath string
	Debug      bool
	SPBaseURL  string
	NRBaseURL  string
}

type ApplicationResponse struct {
	Application Application
}

type Application struct {
	Id                 int                `json:"id"`
	Name               string             `json:"name"`
	Reporting          bool               `json:"reporting"`
	ApplicationSummary ApplicationSummary `json:"application_summary"`
}

type ApplicationSummary struct {
	ResponseTime  float64 `json:"response_time"`
	Throughput    float64 `json:"throughput"`
	ErrorRate     float64 `json:"error_rate"`
	ApdexTarget   float64 `json:"apdex_target"`
	ApdexScore    float64 `json:"apdex_score"`
	HostCount     float64 `json:"host_count"`
	InstanceCount float64 `json:"instance_count"`
}

type App struct {
	NRApiKey  string            `json:"nr_api_key"`
	NRAppId   int               `json:"nr_app_id"`
	SPApiKey  string            `json:"sp_api_key"`
	SPPageId  string            `json:"sp_page_id"`
	SPMetrics map[string]string `json:"metrics"`
}

type Metric struct {
	SPApiKey   string  `json:"sp_api_key"`
	SPPageId   string  `json:"sp_page_id"`
	SPMetricId string  `json:"sp_metric_id"`
	Value      float64 `json:"value"`
}

type SPData struct {
	Timestamp int32   `json:"timestamp"`
	Value     float64 `json:"value"`
}

func PollNR(config Config, app App, metrics chan Metric) {
	appid := strconv.Itoa(app.NRAppId)
	parts := []string{config.NRBaseURL, appid, ".json"}
	url := strings.Join(parts, "")

	client := &http.Client{Timeout: time.Second * 5}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[error] PollNR: new request: %s\n", err)
		return
	}
	req.Header.Set("X-Api-Key", app.NRApiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[error] PollNR: client do: %s\n", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[error] PollNR: couldn't read body: %s\n", err)
		return
	}

	if config.Debug {
		log.Printf("[debug]: PollNR raw body: %s\n", body)
	}

	var sample ApplicationResponse
	err = json.Unmarshal(body, &sample)
	if err != nil {
		log.Printf("[error] PollNR: couldn't decode json: %s", err)
		return
	}
	if config.Debug {
		log.Printf("[debug]: PollNR decoded JSON: %+v\n", sample)
	}

	m := Metric{SPPageId: app.SPPageId, SPApiKey: app.SPApiKey}

	if _, ok := app.SPMetrics["response_time"]; ok {
		if config.Debug {
			log.Println("[debug]: PollNR: Fetching response_time for", appid)
		}
		m.SPMetricId = app.SPMetrics["response_time"]
		m.Value = sample.Application.ApplicationSummary.ResponseTime
		metrics <- m
	}

	if _, ok := app.SPMetrics["throughput"]; ok {
		if config.Debug {
			log.Println("[debug]: PollNR: Fetching throughput for nr_app_id", appid)
		}
		m.SPMetricId = app.SPMetrics["throughput"]
		m.Value = sample.Application.ApplicationSummary.Throughput
		metrics <- m
	}

	if _, ok := app.SPMetrics["error_rate"]; ok {
		if config.Debug {
			log.Println("[debug]: PollNR: Fetching error_rate for", appid)
		}
		m.SPMetricId = app.SPMetrics["error_rate"]
		m.Value = sample.Application.ApplicationSummary.ErrorRate
		metrics <- m
	}
}

func Setup(config Config, apps *[]App) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[error] Setup: unhandled panic when polling for checks:", r)
		}
	}()

	contents, err := ioutil.ReadFile(config.ConfigPath)
	if err != nil {
		log.Printf("[error] Setup: couldn't read contents: %s\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(contents, &apps)
	if err != nil {
		log.Printf("[error] Setup: couldn't decode apps: %s\n", err)
		log.Printf("[error] Setup: response contents: %s\n", string(contents))
		os.Exit(1)
	}

	log.Printf("[info] Setup: Tracking metrics from %d applications", len(*apps))
}

func Dispatch(config Config, metrics chan Metric) {
	for {
		metric := <-metrics
		parts := []string{config.SPBaseURL, "pages", metric.SPPageId, "metrics", metric.SPMetricId, "data.json"}
		url := strings.Join(parts, "/")
		if config.Debug {
			log.Printf("[debug] Dispatch: URL: %s", url)
		}

		payload := map[string]SPData{"data": SPData{Timestamp: int32(time.Now().Unix()), Value: metric.Value}}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[error] Dispatch: JSON marshal: %s\n", err)
			continue
		}
		if config.Debug {
			log.Printf("[debug] Dispatch: JSON marshal: %s", string(body))
		}

		client := &http.Client{Timeout: config.Timeout}
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("[error] Dispatch: new request: %s\n", err)
			continue
		}
		req.Header.Set("Authorization", "OAuth "+metric.SPApiKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[error] Dispatch: client do: %s\n", err)
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[error] Dispatch: couldn't read body: %s\n", err)
			continue
		}

		if resp.StatusCode != 201 {
			log.Printf("[error] Dispatch: StatusPage returned HTTP %d: %s\n", resp.StatusCode, string(body))
		}
	}
}

var (
	configPath = kingpin.Flag("config", "Path to Nudger config").Default("nudger.json").OverrideDefaultFromEnvar("CONFIG_PATH").String()
	debug      = kingpin.Flag("debug", "Debug mode").Default("false").OverrideDefaultFromEnvar("DEBUG").Bool()
	spBaseURL  = kingpin.Flag("statuspage-base-url", "StatusPage API base URL").Default("https://api.statuspage.io/v1").String()
	nrBaseURL  = kingpin.Flag("newrelic-base-url", "New Relic API base URL").Default("https://api.newrelic.com/v2/applications/").String()
)

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()

	config := Config{
		Interval:   time.Second * 10,
		ConfigPath: *configPath,
		Timeout:    time.Second * 5,
		Debug:      *debug,
		SPBaseURL:  *spBaseURL,
		NRBaseURL:  *nrBaseURL,
	}
	if config.Debug {
		log.Printf("[debug] Main: config: %+v\n", config)
	}

	var apps []App
	Setup(config, &apps)

	metrics := make(chan Metric)
	go Dispatch(config, metrics)

	tick := time.NewTicker(config.Interval).C
	for {
		select {
		case <-tick:
			log.Printf("[info] Main: Polling metrics for %d apps", len(apps))
			for _, a := range apps {
				go PollNR(config, a, metrics)
			}
		}
	}
}
