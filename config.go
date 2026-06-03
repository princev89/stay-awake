package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	LaunchAtLogin        bool `json:"launchAtLogin"`
	StartMinimized       bool `json:"startMinimized"`
	AwakeState           bool `json:"awakeState"`
	LidClosePreventSleep bool `json:"lidClosePreventSleep"`
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "StayAwake")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

func LoadConfig() (*AppConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultConfig := &AppConfig{
			LaunchAtLogin:        false,
			StartMinimized:       false,
			AwakeState:           false,
			LidClosePreventSleep: false,
		}
		_ = SaveConfig(defaultConfig)
		return defaultConfig, nil
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func SaveConfig(config *AppConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
