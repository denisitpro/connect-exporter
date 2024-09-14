package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Configuration struct to hold process names, port, and address
type Config struct {
	Processes    []string `toml:"processes"`
	ExporterPort string   `toml:"exporter_port"`
	ExporterAddr string   `toml:"exporter_addr"`
}

// Define metrics for tracking connection states
var (
	connectionState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_connections_state_total",
			Help: "Number of connections grouped by process, protocol, and state.",
		},
		[]string{"process_name", "state"},
	)
)

// Variables for versioning and configuration
var (
	version    = "dev"
	buildTime  = "unknown"
	commitHash = "none"
	debug      = false
	config     Config
)

// Custom logging function with a timestamp
func logRequest(r *http.Request) {
	timestamp := time.Now().Format(time.RFC3339)
	clientIP := r.RemoteAddr
	requestedURI := r.RequestURI
	log.Printf("DEBUG: [%s] Client IP: %s Requested URI: %s", timestamp, clientIP, requestedURI)
}

// Function to run ss command and filter by process name
func parseSSOutput(processName string) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ss -tanup | grep %s | awk '{print $2}' | sort | uniq -c", processName))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error getting ss output for process %s: %v", processName, err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting ss command for process %s: %v", processName, err)
		return
	}

	scanner := bufio.NewScanner(stdout)

	// Debug: Print which process is being processed
	if debug {
		log.Printf("Processing ss output for process: %s", processName)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			log.Println(line)
		}

		// Split line into count and state (e.g., "5 ESTAB")
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		// Extract count and state
		count := parts[0]
		state := parts[1]

		// Update metrics
		connectionState.WithLabelValues(processName, state).Set(stringToFloat(count))
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for ss command for process %s: %v", processName, err)
	}
}

// Helper function to convert string to float64 for Prometheus metrics
func stringToFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("Error converting string to float64: %v", err)
		return 0
	}
	return value
}

// Load configuration from a TOML file
func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}
	return config, nil
}

// Function to update metrics when requested
func updateMetrics(w http.ResponseWriter, r *http.Request) {
	// Log the request if debug is enabled
	if debug {
		logRequest(r)
	}

	// Reset metrics before each request
	connectionState.Reset()

	// Update metrics for each process in the config
	for _, processName := range config.Processes {
		parseSSOutput(processName)
	}

	// Serve the metrics
	promhttp.Handler().ServeHTTP(w, r)
}

// Custom 404 handler to return the metrics URL in the response
func custom404Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain")
	metricsURL := fmt.Sprintf("http://%s:%s/metrics", config.ExporterAddr, config.ExporterPort)
	w.Write([]byte(fmt.Sprintf("404 - Page not found. Please use the metrics endpoint: %s\n", metricsURL)))
}

// Function to check if address is IPv6
func isIPv6(addr string) bool {
	return strings.Contains(addr, ":") && !strings.Contains(addr, ".")
}

func main() {
	// Command-line flags
	configFile := flag.String("c", "config.toml", "Path to the config file")
	showVersion := flag.Bool("version", false, "Show version information")
	enableDebug := flag.Bool("v2", false, "Enable debug logging")
	flag.Parse()

	// Enable debug logging if specified
	debug = *enableDebug

	// Show version info and exit
	if *showVersion {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Commit Hash: %s\n", commitHash)
		os.Exit(0)
	}

	// Load configuration
	var err error
	config, err = loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Set default port and address if not provided in config
	if config.ExporterPort == "" {
		config.ExporterPort = "9042" // Default port
	}
	if config.ExporterAddr == "" {
		config.ExporterAddr = "0.0.0.0" // Default to all interfaces
	}

	// If address is IPv6, wrap it in square brackets
	addr := config.ExporterAddr
	if isIPv6(config.ExporterAddr) {
		addr = "[" + config.ExporterAddr + "]"
	}

	// Construct final address with port
	fullAddr := addr + ":" + config.ExporterPort

	// Register metrics
	prometheus.MustRegister(connectionState)

	// HTTP server for metrics
	http.HandleFunc("/metrics", updateMetrics)

	// Custom handler for 404 errors
	http.HandleFunc("/", custom404Handler)

	// Start HTTP server with proper address and port
	log.Printf("Starting server at %s/metrics", fullAddr)
	log.Fatal(http.ListenAndServe(fullAddr, nil))
}