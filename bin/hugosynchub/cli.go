package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configFile = "config.json"

// Config holds application settings
type Config struct {
	Name       string `json:"name"`
	FolderName string `json:"folder_name"`
}

func Cli() {
	var config Config
	_ = loadConfig(&config)
	reader := bufio.NewReader(os.Stdin)
	currentMenu := "main"

	for {
		switch currentMenu {
		case "main":
			currentMenu = showMainMenu(reader, &config)
		case "settings":
			currentMenu = showSettingsMenu(reader, &config)
		case "exit":
			fmt.Println("Goodbye!")
			saveConfig(config)
			return
		}
	}
}

func showMainMenu(r *bufio.Reader, cfg *Config) string {
	for {
		fmt.Println("\n===== Main Menu =====")
		fmt.Println("1. Install Hugo")
		fmt.Println("2. Initialize SyncHub")
		fmt.Println("9. Settings/Return to Previous Menu")
		fmt.Print("Select an option: ")
		choice := readChoice(r)
		switch choice {
		case "1":
			if ensureSettings(cfg, r) {
				fmt.Println("[Placeholder] Installing Hugo...")
			}
		case "2":
			if ensureSettings(cfg, r) {
				fmt.Println("[Placeholder] Initializing SyncHub...")
			}
		case "9":
			saveConfig(*cfg)
			return "settings"
		default:
			fmt.Println("Invalid input, try again.")
		}
	}
}

func showSettingsMenu(r *bufio.Reader, cfg *Config) string {
	for {
		fmt.Println("\n===== Settings =====")
		nameDisplay := cfg.Name
		if nameDisplay == "" {
			nameDisplay = "None"
		}
		folderDisplay := cfg.FolderName
		if folderDisplay == "" {
			folderDisplay = "None"
		}
		fmt.Printf("1. Set application name (%s)\n", nameDisplay)
		fmt.Printf("2. Set application folder (%s)\n", folderDisplay)
		fmt.Println("9. Return to Previous Menu")
		fmt.Print("Select an option: ")
		choice := readChoice(r)
		switch choice {
		case "1":
			fmt.Print("Enter application name: ")
			cfg.Name = readLine(r)
			fmt.Println("Name set to", cfg.Name)
		case "2":
			fmt.Print("Enter application folder: ")
			input := readLine(r)
			// store absolute or relative path
			cfg.FolderName = filepath.Clean(input)
			fmt.Println("Folder set to", cfg.FolderName)
		case "9":
			saveConfig(*cfg)
			return "main"
		default:
			fmt.Println("Invalid input, try again.")
		}
	}
}

func ensureSettings(cfg *Config, r *bufio.Reader) bool {
	if cfg.Name == "" {
		fmt.Println("Application name is not set. Please set it now.")
		showSettingsMenu(r, cfg)
	}
	if cfg.FolderName == "" {
		fmt.Println("Application folder is not set. Please set it now.")
		showSettingsMenu(r, cfg)
	}
	return cfg.Name != "" && cfg.FolderName != ""
}

func readChoice(r *bufio.Reader) string {
	input := readLine(r)
	return strings.TrimSpace(input)
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

// getSettingsPath returns the path to ../../data/settings.json relative to the executable
func getSettingsPath() (string, error) {
	dataDir := filepath.Join(".", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "settings.json"), nil
}

func saveConfig(config interface{}) error {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return err
	}
	file, err := os.Create(settingsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(config)
}

func loadConfig(config interface{}) error {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return err
	}
	file, err := os.Open(settingsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(config)
}
