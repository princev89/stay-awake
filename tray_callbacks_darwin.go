//go:build darwin
package main

import "C"

//export GoTrayOnToggle
func GoTrayOnToggle() {
	if globalApp != nil {
		globalApp.ToggleAwakeFromTray()
	}
}

//export GoTrayOnOpen
func GoTrayOnOpen() {
	if globalApp != nil {
		globalApp.ShowApp()
	}
}

//export GoTrayOnQuit
func GoTrayOnQuit() {
	if globalApp != nil {
		globalApp.QuitApp()
	}
}
