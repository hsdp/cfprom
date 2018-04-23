// Copyright 2018 Andy Lo-A-Foe. All rights reserved.
// Use of this source code is governed by Apache-style
// license that can be found in the LICENSE file.

package main

import (
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

func main() {
	flag.Parse()

	c := &cfclient.Config{
		ApiAddress: os.Getenv("CF_API"),
		Username:   os.Getenv("CF_USERNAME"),
		Password:   os.Getenv("CF_PASSWORD"),
	}
	fmt.Println("Logging in")
	client, err := cfclient.NewClient(c)
	if err != nil {
		fmt.Printf("CF Client error: %s\n", err.Error())
		return
	}

	appEnv, err := cfenv.Current()

	if err != nil {
		fmt.Printf("Not running in CF. Exiting..\n")
		return
	}

	fmt.Printf("Fetching apps in space: %s\n", appEnv.SpaceID)

	q := url.Values{}
	q.Add("q", fmt.Sprintf("space_guid:%s", appEnv.SpaceID))
	apps, _ := client.ListAppsByQuery(q)

	app := apps[0]

	app, _ = client.GetAppByGuid(app.Guid)
	space, _ := app.Space()
	org, _ := space.Org()

	go func() {
		var i = 0
		for {
			i += 1
			if i >= 60 { // Reread full list every 60 iterations
				i = 0
				apps, _ = client.ListAppsByQuery(q)
			}
			fmt.Printf("Fetching stats of %d apps\n", len(apps))
			for _, app := range apps {
				if app.Guid == appEnv.AppID { // Skip self
					continue
				}
				stats, _ := client.GetAppStats(app.Guid)
				for i, s := range stats {
					cpuGauge.WithLabelValues(org.Name, space.Name, app.Name, i).Set(s.Stats.Usage.CPU * 100)
					memGauge.WithLabelValues(org.Name, space.Name, app.Name, i).Set(float64(s.Stats.Usage.Mem))
				}
			}
			time.Sleep(time.Duration(15 * time.Second))
		}

	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
