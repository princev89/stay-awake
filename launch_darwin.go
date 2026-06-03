//go:build darwin
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func SetLaunchAtLogin(enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	label := "com.stayawake.app"
	plistPath := filepath.Join(home, "Library", "LaunchAgents", label+".plist")
	
	if !enabled {
		if _, err := os.Stat(plistPath); err == nil {
			return os.Remove(plistPath)
		}
		return nil
	}
	
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Create LaunchAgents directory if not exists
	err = os.MkdirAll(filepath.Dir(plistPath), 0755)
	if err != nil {
		return err
	}
	
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, label, execPath)
	
	return os.WriteFile(plistPath, []byte(plistContent), 0644)
}
