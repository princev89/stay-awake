//go:build darwin
package main

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework CoreGraphics
#include <IOKit/pwr_mgt/IOPMLib.h>
#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <CoreGraphics/CoreGraphics.h>
#include <stdlib.h>
#include <dlfcn.h>

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

// Lid state detection
static int isLidClosed() {
    io_service_t rootDomain = IOServiceGetMatchingService(0, IOServiceMatching("IOPMrootDomain"));
    if (rootDomain == MACH_PORT_NULL) {
        return 0;
    }
    CFTypeRef clamshellState = IORegistryEntryCreateCFProperty(rootDomain, CFSTR("AppleClamshellState"), kCFAllocatorDefault, 0);
    IOObjectRelease(rootDomain);
    if (clamshellState == NULL) {
        return 0;
    }
    int closed = 0;
    if (CFGetTypeID(clamshellState) == CFBooleanGetTypeID()) {
        closed = (clamshellState == kCFBooleanTrue);
    }
    CFRelease(clamshellState);
    return closed;
}

// Display brightness function pointers
typedef int (*DisplayServicesGetLinearBrightness_t)(CGDirectDisplayID, float*);
typedef int (*DisplayServicesSetLinearBrightness_t)(CGDirectDisplayID, float);
typedef int (*DisplayServicesGetBrightness_t)(CGDirectDisplayID, float*);
typedef int (*DisplayServicesSetBrightness_t)(CGDirectDisplayID, float);
typedef double (*CoreDisplay_Display_GetUserBrightness_t)(CGDirectDisplayID);
typedef void (*CoreDisplay_Display_SetUserBrightness_t)(CGDirectDisplayID, double);

static void* displayServicesHandle = NULL;
static void* coreDisplayHandle = NULL;

static DisplayServicesGetLinearBrightness_t pDisplayServicesGetLinearBrightness = NULL;
static DisplayServicesSetLinearBrightness_t pDisplayServicesSetLinearBrightness = NULL;
static DisplayServicesGetBrightness_t pDisplayServicesGetBrightness = NULL;
static DisplayServicesSetBrightness_t pDisplayServicesSetBrightness = NULL;
static CoreDisplay_Display_GetUserBrightness_t pCoreDisplay_Display_GetUserBrightness = NULL;
static CoreDisplay_Display_SetUserBrightness_t pCoreDisplay_Display_SetUserBrightness = NULL;

static void initBrightnessAPIs() {
    static int initialized = 0;
    if (initialized) return;
    initialized = 1;

    // Try to load DisplayServices (private framework)
    displayServicesHandle = dlopen("/System/Library/PrivateFrameworks/DisplayServices.framework/DisplayServices", RTLD_LAZY);
    if (displayServicesHandle) {
        pDisplayServicesGetLinearBrightness = (DisplayServicesGetLinearBrightness_t)dlsym(displayServicesHandle, "DisplayServicesGetLinearBrightness");
        pDisplayServicesSetLinearBrightness = (DisplayServicesSetLinearBrightness_t)dlsym(displayServicesHandle, "DisplayServicesSetLinearBrightness");
        pDisplayServicesGetBrightness = (DisplayServicesGetBrightness_t)dlsym(displayServicesHandle, "DisplayServicesGetBrightness");
        pDisplayServicesSetBrightness = (DisplayServicesSetBrightness_t)dlsym(displayServicesHandle, "DisplayServicesSetBrightness");
    }

    // Try to load CoreDisplay (public framework but private symbols)
    coreDisplayHandle = dlopen("/System/Library/Frameworks/CoreDisplay.framework/CoreDisplay", RTLD_LAZY);
    if (coreDisplayHandle) {
        pCoreDisplay_Display_GetUserBrightness = (CoreDisplay_Display_GetUserBrightness_t)dlsym(coreDisplayHandle, "CoreDisplay_Display_GetUserBrightness");
        pCoreDisplay_Display_SetUserBrightness = (CoreDisplay_Display_SetUserBrightness_t)dlsym(coreDisplayHandle, "CoreDisplay_Display_SetUserBrightness");
    }
}

static double getDisplayBrightness() {
    initBrightnessAPIs();
    CGDirectDisplayID mainDisplay = CGMainDisplayID();

    if (pDisplayServicesGetLinearBrightness != NULL) {
        float val = 0.0f;
        if (pDisplayServicesGetLinearBrightness(mainDisplay, &val) == 0) {
            return (double)val;
        }
    }
    if (pDisplayServicesGetBrightness != NULL) {
        float val = 0.0f;
        if (pDisplayServicesGetBrightness(mainDisplay, &val) == 0) {
            return (double)val;
        }
    }
    if (pCoreDisplay_Display_GetUserBrightness != NULL) {
        return pCoreDisplay_Display_GetUserBrightness(mainDisplay);
    }

    return 0.5; // fallback
}

static void setDisplayBrightness(double brightness) {
    initBrightnessAPIs();
    CGDirectDisplayID mainDisplay = CGMainDisplayID();

    if (pDisplayServicesSetLinearBrightness != NULL) {
        pDisplayServicesSetLinearBrightness(mainDisplay, (float)brightness);
    }
    if (pDisplayServicesSetBrightness != NULL) {
        pDisplayServicesSetBrightness(mainDisplay, (float)brightness);
    }
    if (pCoreDisplay_Display_SetUserBrightness != NULL) {
        pCoreDisplay_Display_SetUserBrightness(mainDisplay, brightness);
    }
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

func GetLidClosedState() bool {
	return C.isLidClosed() != 0
}

func GetDisplayBrightness() float64 {
	return float64(C.getDisplayBrightness())
}

func SetDisplayBrightness(brightness float64) {
	C.setDisplayBrightness(C.double(brightness))
}

