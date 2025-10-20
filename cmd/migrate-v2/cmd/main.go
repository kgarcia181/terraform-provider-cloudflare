package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/registry"
	"github.com/cloudflare/terraform-provider-cloudflare/cmd/migrate-v2/resources"
)

func main() {
	// Command line flags
	var (
		dryRun       bool
		configDir    string
		stateFile    string
		resourceList string
		verbose      bool
	)
	
	flag.BoolVar(&dryRun, "dryrun", false, "Show what changes would be made without actually modifying files")
	flag.StringVar(&configDir, "config", "", "Directory containing Terraform files to migrate (optional)")
	flag.StringVar(&stateFile, "state", "", "Terraform state file to migrate (optional)")
	flag.StringVar(&resourceList, "resource", "", "Comma-separated list of resources to migrate (optional, requires -config)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Configure slog based on verbosity
	var logLevel slog.Level
	if verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Validate flag combinations
	if resourceList != "" && configDir == "" {
		slog.Error("Invalid flag combination", "error", "-resource flag requires -config to be specified")
		os.Exit(1)
	}

	if configDir == "" && stateFile == "" {
		slog.Error("No input specified", 
			"error", "At least one of -config or -state must be specified",
			"usage", "migrate-v2 [flags]",
			"available_resources", strings.Join(resources.GetAvailableResources(), ", "))
		flag.PrintDefaults()
		os.Exit(1)
	}

	if dryRun {
		slog.Info("Running in dry-run mode - no files will be modified")
	}

	// Create registry
	reg := registry.NewStrategyRegistry()

	// Register resources based on flags
	if resourceList == "" {
		// Register all available resources
		resources.RegisterAll(reg)
		slog.Debug("Registered resources", "count", reg.Count(), "mode", "all")
	} else {
		// Register specific resources
		requestedResources := strings.Split(resourceList, ",")
		for i, r := range requestedResources {
			requestedResources[i] = strings.TrimSpace(r)
		}
		resources.RegisterFromFactories(reg, requestedResources...)
		slog.Debug("Registered resources", "count", reg.Count(), "resources", resourceList)
	}

	// Process config directory if specified
	if configDir != "" {
		slog.Info("Processing config directory", "path", configDir)
		if err := processConfigDirectory(configDir, reg, dryRun); err != nil {
			slog.Error("Failed to process config directory", "error", err)
			os.Exit(1)
		}
	}

	// Process state file if specified
	if stateFile != "" {
		slog.Info("Processing state file", "path", stateFile)
		if err := processStateFile(stateFile, reg, dryRun); err != nil {
			slog.Error("Failed to process state file", "error", err)
			os.Exit(1)
		}
	}

	if !dryRun {
		slog.Info("Migration completed successfully")
	} else {
		slog.Info("Dry-run completed - no files were modified")
	}
}

func processConfigDirectory(directory string, reg *registry.StrategyRegistry, dryRun bool) error {
	// Check if the directory exists
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %v", directory, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", directory)
	}

	// Build the transformation pipeline
	pipeline := migrate_v2.BuildPipelineWithDryRun(reg, dryRun)

	// Walk through the directory recursively
	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Error accessing path", "path", path, "error", err)
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .tf files
		if !strings.HasSuffix(strings.ToLower(path), ".tf") {
			return nil
		}

		// Process the file
		if err := processConfigFile(path, pipeline, dryRun); err != nil {
			slog.Warn("Error processing file", "file", path, "error", err)
			// Continue processing other files
		}

		return nil
	})
}

func processConfigFile(filename string, pipeline *migrate_v2.Pipeline, dryRun bool) error {
	// Read the file
	originalBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Transform the content
	result, err := pipeline.Transform(originalBytes, filename)
	if err != nil {
		return fmt.Errorf("transformation failed: %v", err)
	}

	// Check if anything changed
	if string(originalBytes) == string(result) {
		return nil
	}

	if dryRun {
		slog.Info("Would update file", "file", filename, "action", "skip")
	} else {
		// Write the result back to the file
		err = os.WriteFile(filename, result, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", filename, err)
		}
		slog.Info("Updated file", "file", filename, "action", "write")
	}

	return nil
}

func processStateFile(stateFilePath string, reg *registry.StrategyRegistry, dryRun bool) error {
	// Check if the state file exists
	info, err := os.Stat(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to access state file %s: %v", stateFilePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, expected a file", stateFilePath)
	}

	// Check if it's a .tfstate file
	if !strings.HasSuffix(strings.ToLower(stateFilePath), ".tfstate") {
		return fmt.Errorf("%s is not a .tfstate file", stateFilePath)
	}

	// Read the state file
	originalBytes, err := os.ReadFile(stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %v", err)
	}

	// Build state transformation pipeline
	pipeline := migrate_v2.BuildStatePipeline(reg, dryRun)

	// Transform the state
	result, err := pipeline.TransformState(originalBytes, stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to transform state: %v", err)
	}

	if dryRun {
		slog.Info("Would update state file", "file", stateFilePath, "size", info.Size(), "action", "skip")
	} else {
		// Write the transformed state back
		err = os.WriteFile(stateFilePath, result, 0644)
		if err != nil {
			return fmt.Errorf("failed to write state file: %v", err)
		}
		slog.Info("Updated state file", "file", stateFilePath, "action", "write")
	}

	return nil
}
