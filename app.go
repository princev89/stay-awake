package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var globalApp *App

type App struct {
	ctx              context.Context
	config           *AppConfig
	sleepManager     *SleepManager
	quitting         bool
	lidClosed        bool
	savedBrightness  float64
	brightnessDimmed bool
}

func NewApp() *App {
	config, err := LoadConfig()
	if err != nil {
		// Fallback to default config on error
		config = &AppConfig{
			LaunchAtLogin:  false,
			StartMinimized: false,
			AwakeState:     false,
		}
	}
	return &App{
		config:       config,
		sleepManager: NewSleepManager(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	globalApp = a
	
	// Initialise Tray menu
	InitTray()
	
	// Restore last awake state
	if a.config.AwakeState {
		_ = a.sleepManager.Acquire("Stay Awake Active")
		if a.config.LidClosePreventSleep {
			_ = SetLidSleepDisabled(true)
		}
		UpdateTray(true)
	} else {
		UpdateTray(false)
	}

	// Handle Start Minimized (launching directly into tray)
	if a.config.StartMinimized {
		go func() {
			// Give the window system a brief moment to render, then hide it
			time.Sleep(100 * time.Millisecond)
			runtime.WindowHide(a.ctx)
		}()
	}

	// Start background lid state monitor
	a.startLidMonitor()
}

func (a *App) domReady(ctx context.Context) {
	// Sync the UI with the startup state
	runtime.EventsEmit(a.ctx, "state-changed", a.config.AwakeState)
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting {
		_ = a.sleepManager.Release()
		_ = SetLidSleepDisabled(false) // Restore sleep settings on quit
		a.restoreBrightness()          // Restore display brightness on quit
		return false // Allow app to quit
	}
	// Hide window to system tray instead of closing
	runtime.WindowHide(a.ctx)
	return true // Prevent default close (don't exit)
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.sleepManager.Release()
	_ = SetLidSleepDisabled(false) // Restore sleep settings on quit
	a.restoreBrightness()          // Restore display brightness on quit
}

// BINDINGS - Exposed to Frontend

func (a *App) GetConfig() *AppConfig {
	return a.config
}

func (a *App) ToggleAwake(active bool) bool {
	a.config.AwakeState = active
	_ = SaveConfig(a.config)
	
	if active {
		_ = a.sleepManager.Acquire("Stay Awake Active")
		if a.config.LidClosePreventSleep {
			_ = SetLidSleepDisabled(true)
		}
	} else {
		_ = a.sleepManager.Release()
		_ = SetLidSleepDisabled(false)
		a.restoreBrightness()
	}
	
	UpdateTray(active)
	return active
}

func (a *App) SetLaunchAtLogin(enabled bool) bool {
	a.config.LaunchAtLogin = enabled
	_ = SaveConfig(a.config)
	_ = SetLaunchAtLogin(enabled)
	return enabled
}

func (a *App) SetStartMinimized(enabled bool) bool {
	a.config.StartMinimized = enabled
	_ = SaveConfig(a.config)
	return enabled
}

func (a *App) SetLidClosePreventSleep(enabled bool) bool {
	a.config.LidClosePreventSleep = enabled
	_ = SaveConfig(a.config)
	
	// If currently awake, apply or remove the sleep block immediately
	if a.config.AwakeState {
		if enabled {
			_ = SetLidSleepDisabled(true)
		} else {
			_ = SetLidSleepDisabled(false)
			a.restoreBrightness()
		}
	}
	return enabled
}

func (a *App) startLidMonitor() {
	go func() {
		// Wait a moment for startup to finish
		time.Sleep(1 * time.Second)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			if a.ctx == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				a.checkLidState()
			}
		}
	}()
}

func (a *App) checkLidState() {
	if !a.config.AwakeState || !a.config.LidClosePreventSleep {
		if a.brightnessDimmed {
			a.restoreBrightness()
		}
		return
	}

	closed := GetLidClosedState()
	if closed && !a.lidClosed {
		a.lidClosed = true
		a.dimBrightness()
	} else if !closed && a.lidClosed {
		a.lidClosed = false
		a.restoreBrightness()
	}
}

func (a *App) dimBrightness() {
	if a.brightnessDimmed {
		return
	}
	curr := GetDisplayBrightness()
	if curr > 0.05 {
		a.savedBrightness = curr
	} else if a.savedBrightness <= 0.05 {
		a.savedBrightness = 0.8
	}
	SetDisplayBrightness(0.0)
	a.brightnessDimmed = true
}

func (a *App) restoreBrightness() {
	if !a.brightnessDimmed {
		return
	}
	if a.savedBrightness > 0.0 {
		SetDisplayBrightness(a.savedBrightness)
	}
	a.brightnessDimmed = false
	a.lidClosed = false
}


// TRAY CONTROLS - Triggered by native Objective-C callbacks

func (a *App) ToggleAwakeFromTray() {
	newState := !a.config.AwakeState
	a.ToggleAwake(newState)
	// Update HTML interface
	runtime.EventsEmit(a.ctx, "state-changed", newState)
}

func (a *App) ShowApp() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	// Briefly make it always on top to bring it to foreground focus
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	go func() {
		time.Sleep(300 * time.Millisecond)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}()
}

func (a *App) QuitApp() {
	a.quitting = true
	runtime.Quit(a.ctx)
}

func (a *App) CheckForUpdates() map[string]interface{} {
	hasUpdate, tag, url, err := CheckForUpdate(AppVersion)
	res := map[string]interface{}{
		"currentVersion": AppVersion,
		"hasUpdate":      hasUpdate,
		"latestVersion":  tag,
		"releaseUrl":     url,
		"error":          "",
	}
	if err != nil {
		res["error"] = err.Error()
	}
	return res
}

func (a *App) OpenReleaseUrl(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}
