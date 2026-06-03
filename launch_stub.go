//go:build !darwin
package main

func SetLaunchAtLogin(enabled bool) error {
	return nil
}
