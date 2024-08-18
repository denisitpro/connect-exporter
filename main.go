package main

import (
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/process"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Configuration struct to hold process names
type Config struct {
	Processes []string `toml:"processes"`
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
	// Variables for versioning
	version   = "dev"
	buildTime = "unknown"
	commitHash = "none" // New variable for commit hash

	// Variable for logging level
	debug = false
)

func init() {
	// Register metrics
	prometheus.MustRegister(processConnections)
	prometheus.MustRegister(processExists)
}

// Custom logging function to include timestamp
func logRequest(r *http.Request) {
	timestamp := time.Now().Format(time.RFC3339)
	clientIP := r.RemoteAddr
	requestedURI := r.RequestURI
	log.Printf("DEBUG: [%s] Client IP: %s Requested URI: %s", timestamp, clientIP, requestedURI)
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
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse TOML config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func main() {
	// Command-line flags
	configFile := flag.String("c", "config.toml", "Path to the config file")
	showVersion := flag.Bool("version", false, "Show version information")
	enableDebug := flag.Bool("v2", false, "Enable debug logging")
	flag.Parse()

	// Enable debug logging if specified
	debug = *enableDebug

	// Show version info
	if *showVersion {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Commit Hash: %s\n", commitHash) // Display commit hash
		os.Exit(0)
	}

	// Check for remaining arguments after parsing
	if len(flag.Args()) > 0 {
		fmt.Println("Invalid argument:", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Get port from environment variable, default to 9042 if not set
	port := os.Getenv("EXPORTER_PORT")
	if port == "" {
		port = "9042"
	}

	// Start the Prometheus HTTP server with logging for all requests
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if debug {
			logRequest(r)
		}
		promhttp.Handler().ServeHTTP(w, r)
	})

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