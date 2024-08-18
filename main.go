package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/process"
)

// Configuration struct to hold process names
type Config struct {
	Processes []string `json:"processes"`
}

// Define metrics
var (
	processConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_network_connections",
			Help: "Number of active network connections for specified processes",
		},
		[]string{"process_name", "state"},
	)
	processExists = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_exists",
			Help: "Indicates if the process is running (1) or not (0)",
		},
		[]string{"process_name"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(processConnections)
	prometheus.MustRegister(processExists)
}

func getConnections(config Config) {
	// Reset previous metrics
	processConnections.Reset()
	processExists.Reset()

	// Get list of all processes
	allProcs, err := process.Processes()
	if err != nil {
		log.Printf("Error fetching processes: %v", err)
		return
	}

	for _, proc := range allProcs {
		name, err := proc.Name()
		if err != nil {
			continue
		}

		// Check if the process name matches any of the target processes
		for _, pname := range config.Processes {
			if strings.Contains(name, pname) {
				// Mark process as running
				processExists.WithLabelValues(pname).Set(1)

				conns, err := proc.Connections()
				if err != nil {
					log.Printf("Error fetching connections for process %s: %v", name, err)
					continue
				}

				// Count connections by state
				connStateCount := make(map[string]int)
				for _, conn := range conns {
					connStateCount[conn.Status]++
				}

				// Update metrics
				for state, count := range connStateCount {
					processConnections.WithLabelValues(pname, state).Set(float64(count))
				}
			} else {
				// If process not found, set metric to 0
				processExists.WithLabelValues(pname).Set(0)
			}
		}
	}
}

func loadConfig(filename string) (Config, error) {
	var config Config

	// Read config file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse JSON config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func main() {
	// Accept configuration file via command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./exporter config.json")
		os.Exit(1)
	}
	configFile := os.Args[1]

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Get port from environment variable, default to 9042 if not set
	port := os.Getenv("EXPORTER_PORT")
	if port == "" {
		port = "9042"
	}

	// Start the Prometheus HTTP server
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("Starting server on port %s...", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	// Periodically collect and update metrics
	for {
		getConnections(config)
		time.Sleep(10 * time.Second)
	}
}