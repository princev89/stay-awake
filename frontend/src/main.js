// Elements
const powerRing = document.getElementById("power-ring");
const toggleBtn = document.getElementById("toggle-btn");
const statusLabel = document.getElementById("status-label");
const statusDesc = document.getElementById("status-desc");
const launchLoginCheck = document.getElementById("launch-login-check");
const startMinimizedCheck = document.getElementById("start-minimized-check");
const lidCloseCheck = document.getElementById("lid-close-check");
const updateBanner = document.getElementById("update-banner");
const updateBannerText = document.getElementById("update-banner-text");
const updateLinkBtn = document.getElementById("update-link-btn");

let isAwake = false;

// Update the UI state based on Awake / Normal status
function updateUI(awake) {
  isAwake = awake;
  if (awake) {
    powerRing.classList.remove("state-normal");
    powerRing.classList.add("state-awake");
    statusLabel.textContent = "AWAKE";
    statusDesc.textContent = "Your Mac is locked awake. Long running tasks can continue uninterrupted.";
  } else {
    powerRing.classList.remove("state-awake");
    powerRing.classList.add("state-normal");
    statusLabel.textContent = "NORMAL";
    statusDesc.textContent = "System sleep is managed by macOS settings.";
  }
}

// Initial boot logic
async function init() {
  // Wait until Wails runtime is fully loaded
  if (!window.go || !window.go.main || !window.go.main.App) {
    setTimeout(init, 50);
    return;
  }

  try {
    // 1. Get initial configuration from Go
    const config = await window.go.main.App.GetConfig();
    updateUI(config.awakeState);
    launchLoginCheck.checked = config.launchAtLogin;
    startMinimizedCheck.checked = config.startMinimized;
    lidCloseCheck.checked = config.lidClosePreventSleep;

    // 2. Set up event listeners for settings
    launchLoginCheck.addEventListener("change", async (e) => {
      await window.go.main.App.SetLaunchAtLogin(e.target.checked);
    });

    startMinimizedCheck.addEventListener("change", async (e) => {
      await window.go.main.App.SetStartMinimized(e.target.checked);
    });

    lidCloseCheck.addEventListener("change", async (e) => {
      const targetState = e.target.checked;
      try {
        const approved = await window.go.main.App.SetLidClosePreventSleep(targetState);
        e.target.checked = approved;
      } catch (err) {
        // Revert switch on cancel or authorization error
        e.target.checked = !targetState;
        console.error("Authorization failed or canceled:", err);
      }
    });

    // 3. Set up button toggle action
    toggleBtn.addEventListener("click", async () => {
      const newState = !isAwake;
      const res = await window.go.main.App.ToggleAwake(newState);
      updateUI(res);
    });

    // 4. Listen to backend events (e.g. from system tray toggle)
    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn("state-changed", (awake) => {
        updateUI(awake);
      });
    }

    // 5. Run GitHub update check asynchronously
    try {
      const updateStatus = await window.go.main.App.CheckForUpdates();
      if (updateStatus && updateStatus.hasUpdate) {
        updateBannerText.textContent = `New Update: ${updateStatus.latestVersion}`;
        updateBanner.classList.remove("hidden");
        updateLinkBtn.addEventListener("click", () => {
          window.go.main.App.OpenReleaseUrl(updateStatus.releaseUrl);
        });
      }
    } catch (updateErr) {
      console.error("Failed to check for updates:", updateErr);
    }

  } catch (err) {
    console.error("Failed to initialize Stay Awake Go bindings:", err);
  }
}

document.addEventListener("DOMContentLoaded", init);
