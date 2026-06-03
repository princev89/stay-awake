//go:build !darwin
package main

type SleepManager struct {
	active bool
}

func NewSleepManager() *SleepManager {
	return &SleepManager{}
}

func (s *SleepManager) Acquire(reason string) error {
	s.active = true
	return nil
}

func (s *SleepManager) Release() error {
	s.active = false
	return nil
}

func SetLidSleepDisabled(disabled bool) error {
	return nil
}

func GetLidClosedState() bool {
	return false
}

func GetDisplayBrightness() float64 {
	return 1.0
}

func SetDisplayBrightness(brightness float64) {
}

