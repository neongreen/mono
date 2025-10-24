package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/tomlcp"
)

func main() {
	fmt.Println("=== Configuration Manager Example ===\n")

	// Create a sample config file
	sampleConfig := `# Application Configuration
# This file is auto-generated. Please edit with care.

[app]
name = "MyApp"           # Application name
version = "1.0.0"        # Current version
environment = "development"

# Server settings
[server]
host = "0.0.0.0"        # Bind address
port = 8080             # HTTP port
workers = 4             # Number of worker threads

# Database configuration
[database]
driver = "postgres"
host = "localhost"
port = 5432
name = "myapp_dev"
user = "postgres"
pool_size = 10          # Connection pool size

# Logging configuration
[logging]
level = "info"          # Log level: debug, info, warn, error
format = "json"         # Log format: json, text
output = "stdout"       # Output: stdout, stderr, file
`

	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", "tomlcp-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")

	// Write initial config
	err = os.WriteFile(configPath, []byte(sampleConfig), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created config file at: %s\n\n", configPath)

	// Read and modify the configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := tomlcp.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	// Scenario 1: Update to production settings
	fmt.Println("Scenario 1: Switching to production environment")
	fmt.Println(strings.Repeat("-", 50))

	doc.Set("app.version", "1.1.0")
	doc.Set("app.environment", "production")
	doc.Set("server.workers", 8)
	doc.Set("logging.level", "warn")
	doc.Set("database.name", "myapp_prod")
	doc.Set("database.pool_size", 20)

	// Save the updated config
	err = os.WriteFile(configPath, doc.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated configuration (comments preserved):")
	fmt.Println(doc.String())

	// Scenario 2: Add new features
	fmt.Println("\nScenario 2: Adding cache and monitoring configuration")
	fmt.Println(strings.Repeat("-", 50))

	doc.Set("cache.enabled", true)
	doc.Set("cache.type", "redis")
	doc.Set("cache.host", "localhost")
	doc.Set("cache.port", 6379)

	doc.Set("monitoring.enabled", true)
	doc.Set("monitoring.metrics_port", 9090)
	doc.Set("monitoring.health_check_interval", 30)

	fmt.Println("Configuration with new sections:")
	fmt.Println(doc.String())

	// Scenario 3: Read and display current settings
	fmt.Println("\nScenario 3: Reading current configuration")
	fmt.Println(strings.Repeat("-", 50))

	appName, _ := doc.Get("app.name")
	version, _ := doc.Get("app.version")
	env, _ := doc.Get("app.environment")
	serverPort, _ := doc.Get("server.port")
	dbName, _ := doc.Get("database.name")
	logLevel, _ := doc.Get("logging.level")

	fmt.Printf("Application: %s v%s (%s)\n", appName, version, env)
	fmt.Printf("Server: %s:%v\n", "0.0.0.0", serverPort)
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Log Level: %s\n", logLevel)

	// Check if cache is enabled
	if doc.Has("cache.enabled") {
		cacheEnabled, _ := doc.Get("cache.enabled")
		fmt.Printf("Cache: %v\n", cacheEnabled)
	}

	// Scenario 4: Remove deprecated settings
	fmt.Println("\n\nScenario 4: Removing deprecated settings")
	fmt.Println(strings.Repeat("-", 50))

	// Suppose we want to remove the old logging.output setting
	if doc.Has("logging.output") {
		fmt.Println("Removing deprecated 'logging.output' setting...")
		doc.Delete("logging.output")
	}

	// Save final configuration
	err = os.WriteFile(configPath, doc.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nFinal configuration saved:")
	finalData, _ := os.ReadFile(configPath)
	fmt.Println(string(finalData))

	fmt.Printf("\nConfiguration file saved to: %s\n", configPath)
	fmt.Println("\nNote: All original comments have been preserved throughout all modifications!")
}
