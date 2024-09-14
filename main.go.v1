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
	"regexp"
	"strings"
	"sync"
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
		[]string{"process_name", "protocol", "state"},
	)
	processExists = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_exists",
			Help: "Indicates if the process is running (1) or not (0)",
		},
		[]string{"process_name"},
	)
	// Variables for versioning
	version    = "dev"
	buildTime  = "unknown"
	commitHash = "none"

	// Variable for logging level
	debug = false
	mu    sync.Mutex

	// Config to store processes from config.toml
	config Config
)

// Map of numerical protocol values to their string representations
var protocolMap = map[string]string{
	"0":   "ip",
	"1":   "icmp",
	"6":   "tcp",
	"17":  "udp",
	"41":  "ipv6",
	"58":  "ipv6-icmp",
	"132": "sctp",
	"162": "ethernet-over-ip",
}

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

// Function to run netstat and capture output
func getNetstatOutput() (string, error) {
	// Run the netstat command
	cmd := exec.Command("netstat", "-tanup")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute netstat: %v", err)
	}

	return string(out), nil
}

// Helper function to check if process name is in the config list
func processInConfig(processName string) bool {
	for _, pname := range config.Processes {
		if strings.Contains(processName, pname) {
			return true
		}
	}
	return false
}

// Helper function to convert numerical protocol values to their string representations
func convertProtocol(protocol string) string {
	if protoName, found := protocolMap[protocol]; found {
		return protoName
	}
	return protocol // Return the original if not found in the map
}

// Function to parse netstat output and update metrics
func updateMetricsFromNetstat() {
	// Get netstat output
	output, err := getNetstatOutput()
	if err != nil {
		log.Printf("Error fetching netstat output: %v", err)
		return
	}

	// Process each line of the output
	scanner := bufio.NewScanner(strings.NewReader(output))
	re := regexp.MustCompile(`(\S+)\s+(\S+):(\d+)\s+(\S+):(\d+)\s+(\S+)\s+(\d+)/(\S+)`)

	// Reset previous metrics
	processConnections.Reset()

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		// Extract information from the matched line
		protocol := matches[1]    // protocol number (e.g., 6 for TCP)
		srcIP := matches[2]       // source IP
		srcPort := matches[3]     // source port
		destIP := matches[4]      // destination IP
		destPort := matches[5]    // destination port
		state := matches[6]       // connection state
		pid := matches[7]         // process PID
		processName := matches[8] // process name

		// Check if process is in config
		if !processInConfig(processName) {
			continue // Skip if not in config
		}

		// Convert numerical protocol to string representation
		protocol = convertProtocol(protocol)

		// Log for debugging
		if debug {
			log.Printf("DEBUG: protocol=%s, src=%s:%s, dest=%s:%s, state=%s, process=%s (PID=%s)", protocol, srcIP, srcPort, destIP, destPort, state, processName, pid)
		}

		// Update Prometheus metrics
		processConnections.WithLabelValues(processName, protocol, state).Inc()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading netstat output: %v", err)
	}
}

// Load configuration from the given TOML file
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

	// Load configuration
	var err error
	config, err = loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Show version info
	if *showVersion {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Commit Hash: %s\n", commitHash)
		os.Exit(0)
	}

	// Start the Prometheus HTTP server with logging for all requests
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if debug {
			logRequest(r)
		}

		// Update metrics on each request
		updateMetricsFromNetstat()
		promhttp.Handler().ServeHTTP(w, r)
	})

	port := "9042"
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}