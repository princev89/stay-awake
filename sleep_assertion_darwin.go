//go:build darwin
package main

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/pwr_mgt/IOPMLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

static CFStringRef createCFString(const char* cstr) {
    return CFStringCreateWithCString(kCFAllocatorDefault, cstr, kCFStringEncodingUTF8);
}

static int isNullCFString(CFStringRef str) {
    return str == NULL;
}

static IOReturn createSleepAssertion(CFStringRef reason, IOPMAssertionID *assertionID) {
    return IOPMAssertionCreateWithName(kIOPMAssertionTypeNoIdleSleep, kIOPMAssertionLevelOn, reason, assertionID);
}

static void releaseSleepAssertion(IOPMAssertionID assertionID) {
    IOPMAssertionRelease(assertionID);
}
*/
import "C"
import (
	"fmt"
	"os/exec"
	"unsafe"
)

type SleepManager struct {
	assertionID   C.IOPMAssertionID
	caffeinateCmd *exec.Cmd
	active        bool
	useFallback   bool
}

func NewSleepManager() *SleepManager {
	return &SleepManager{}
}

func (s *SleepManager) Acquire(reason string) error {
	if s.active {
		return nil
	}

	// 1. Try Native Apple IOKit API
	reasonC := C.CString(reason)
	defer C.free(unsafe.Pointer(reasonC))
	
	reasonStr := C.createCFString(reasonC)
	if C.isNullCFString(reasonStr) == 0 {
		defer C.CFRelease(C.CFTypeRef(reasonStr))
		
		var assID C.IOPMAssertionID
		ret := C.createSleepAssertion(reasonStr, &assID)
		if ret == C.kIOReturnSuccess {
			s.assertionID = assID
			s.active = true
			s.useFallback = false
			return nil
		}
	}

	// 2. Fall back to caffeinate process if native API failed or is not working
	// -i: prevent system idle sleep
	// -d: prevent display sleep
	cmd := exec.Command("caffeinate", "-i", "-d")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("native assertion failed, and caffeinate fallback failed: %w", err)
	}
	s.caffeinateCmd = cmd
	s.active = true
	s.useFallback = true
	return nil
}

func (s *SleepManager) Release() error {
	if !s.active {
		return nil
	}

	if s.useFallback {
		if s.caffeinateCmd != nil && s.caffeinateCmd.Process != nil {
			_ = s.caffeinateCmd.Process.Kill()
			_ = s.caffeinateCmd.Wait()
		}
		s.caffeinateCmd = nil
	} else {
		C.releaseSleepAssertion(s.assertionID)
	}

	s.active = false
	return nil
}

func SetLidSleepDisabled(disabled bool) error {
	var val string
	if disabled {
		val = "1"
	} else {
		val = "0"
	}
	script := fmt.Sprintf(`do shell script "pmset -a disablesleep %s" with administrator privileges`, val)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}
