//go:build darwin
package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

extern void GoTrayOnToggle();
extern void GoTrayOnOpen();
extern void GoTrayOnQuit();

@interface TrayHandler : NSObject
- (void)onToggle:(id)sender;
- (void)onOpen:(id)sender;
- (void)onQuit:(id)sender;
@end

@implementation TrayHandler
- (void)onToggle:(id)sender {
    GoTrayOnToggle();
}
- (void)onOpen:(id)sender {
    GoTrayOnOpen();
}
- (void)onQuit:(id)sender {
    GoTrayOnQuit();
}
@end

static NSStatusItem *statusItem = nil;
static TrayHandler *trayHandler = nil;
static NSMenuItem *toggleMenuItem = nil;

static void setupStatusItem() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem != nil) return;
        
        statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        [statusItem retain];
        
        statusItem.button.title = @"💤 Stay Awake";
        
        trayHandler = [[TrayHandler alloc] init];
        [trayHandler retain];
        
        NSMenu *menu = [[NSMenu alloc] init];
        [menu retain];
        
        toggleMenuItem = [[NSMenuItem alloc] initWithTitle:@"Stay Awake: Normal"
                                                    action:@selector(onToggle:)
                                             keyEquivalent:@""];
        [toggleMenuItem setTarget:trayHandler];
        [toggleMenuItem retain];
        [menu addItem:toggleMenuItem];
        
        NSMenuItem *openItem = [[NSMenuItem alloc] initWithTitle:@"Open App"
                                                          action:@selector(onOpen:)
                                                   keyEquivalent:@""];
        [openItem setTarget:trayHandler];
        [menu addItem:openItem];
        
        [menu addItem:[NSMenuItem separatorItem]];
        
        NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                          action:@selector(onQuit:)
                                                   keyEquivalent:@"q"];
        [quitItem setTarget:trayHandler];
        [menu addItem:quitItem];
        
        statusItem.menu = menu;
    });
}

static void updateStatusItem(int active) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem == nil) return;
        if (active) {
            statusItem.button.title = @"⚡ Stay Awake";
            [toggleMenuItem setTitle:@"Stay Awake: Awake"];
            [toggleMenuItem setState:NSControlStateValueOn];
        } else {
            statusItem.button.title = @"💤 Stay Awake";
            [toggleMenuItem setTitle:@"Stay Awake: Normal"];
            [toggleMenuItem setState:NSControlStateValueOff];
        }
    });
}
*/
import "C"

func InitTray() {
	C.setupStatusItem()
}

func UpdateTray(active bool) {
	if active {
		C.updateStatusItem(1)
	} else {
		C.updateStatusItem(0)
	}
}
