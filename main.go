// Copyright 2018 Andy Lo-A-Foe. All rights reserved.
// Use of this source code is governed by Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr     = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	cpuGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_usage",
			Help: "CPU usage",
		},
		[]string{"org", "space", "app", "instance"})
	memGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mem_usage",
			Help: "Memory usage",
		},
		[]string{"org", "space", "app", "instance"})
)

func init() {
	prometheus.MustRegister(cpuGauge)
	prometheus.MustRegister(memGauge)
}

type config struct {
	cfclient.Config
	SpaceID string
	AppID   string
}

type bootstrapRequest struct {
	Password string `json:"password"`
}

type bootstrapResponse struct {
	Status string `json:"status"`
}

func main() {
	flag.Parse()

	c := config{
		cfclient.Config{
			ApiAddress: os.Getenv("CF_API"),
			Username:   os.Getenv("CF_USERNAME"),
			Password:   os.Getenv("CF_PASSWORD"),
		},
		"",
		"",
	}
	appEnv, err := cfenv.Current()
	if err != nil {
		fmt.Printf("Not running in CF. Exiting..\n")
		return
	}
	c.AppID = appEnv.AppID
	c.SpaceID = appEnv.SpaceID

	ch := make(chan config)

	go monitor(ch)

	ch <- c // Initial config

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/bootstrap", bootstrapHandler(ch))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func bootstrapHandler(ch chan config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var b bootstrapRequest
		var resp bootstrapResponse

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&b)
		defer req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Reconfigure
		if b.Password != "" {
			c := config{
				cfclient.Config{
					ApiAddress: os.Getenv("CF_API"),
					Username:   os.Getenv("CF_USERNAME"),
					Password:   b.Password,
				},
				"",
				"",
			}
			appEnv, err := cfenv.Current()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			c.AppID = appEnv.AppID
			c.SpaceID = appEnv.SpaceID

			ch <- c // Magic

			resp.Status = "OK"
		} else {
			resp.Status = "ERROR: missing password"
		}
		js, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}

func monitor(ch chan config) {
	var loggedIn = false
	var client *cfclient.Client
	var apps []cfclient.App
	var cfg config
	var spaceName = ""
	var orgName = ""

	check := time.NewTicker(time.Second * 15)
	refresh := time.NewTicker(time.Second * 15 * 60)

	for {
		select {
		case newConfig := <-ch:
			// Configure
			fmt.Println("Logging in after receiving configuration")
			newClient, err := cfclient.NewClient(&cfg.Config)
			if err != nil {
				fmt.Printf("Error logging in: %v\n", err)
				continue
			}
			client = newClient
			fmt.Printf("Fetching apps in space: %s\n", cfg.SpaceID)
			q := url.Values{}
			q.Add("q", fmt.Sprintf("space_guid:%s", cfg.SpaceID))
			apps, _ = client.ListAppsByQuery(q)
			app := apps[0]
			app, _ = client.GetAppByGuid(app.Guid)
			space, _ := app.Space()
			org, _ := space.Org()
			spaceName = space.Name
			orgName = org.Name
			cfg = newConfig
			loggedIn = true
		case <-refresh.C:
			if cfg.Config.Password == "" {
				fmt.Println("No configuration available during refresh")
				continue
			}
			fmt.Println("Refreshing login")
			newClient, err := cfclient.NewClient(&cfg.Config)
			if err != nil {
				fmt.Printf("Error refreshing login: %v\n", err)
				loggedIn = false
				continue
			}
			client = newClient
			q := url.Values{}
			q.Add("q", fmt.Sprintf("space_guid:%s", cfg.SpaceID))
			apps, _ = client.ListAppsByQuery(q)
		case <-check.C:
			if !loggedIn {
				continue
			}
			fmt.Printf("Fetching stats of %d apps\n", len(apps))
			for _, app := range apps {
				if app.Guid == cfg.AppID { // Skip self
					continue
				}
				stats, _ := client.GetAppStats(app.Guid)
				for i, s := range stats {
					cpuGauge.WithLabelValues(orgName, spaceName, app.Name, i).Set(s.Stats.Usage.CPU * 100)
					memGauge.WithLabelValues(orgName, spaceName, app.Name, i).Set(float64(s.Stats.Usage.Mem))
				}
			}
		}
	}
}
