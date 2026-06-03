cask "stay-awake" do
  version "1.0.3"
  sha256 :no_check # Or replace with the actual SHA256 hash of your release zip file (e.g. "a1b2c3d4...")

  url "https://github.com/princev89/stay-awake/releases/download/v#{version}/Stay.Awake.zip"
  name "Stay Awake"
  desc "Keep your Mac awake while builds, scripts, and long-running tasks execute"
  homepage "https://github.com/princev89/stay-awake"

  # GUI App configuration
  app "Stay Awake.app"

  # Clean up app configurations and launch agents upon uninstall
  zap trash: [
    "~/Library/Application Support/StayAwake",
    "~/Library/LaunchAgents/com.stayawake.app.plist",
  ]
end
