#!/bin/bash

# Stay Awake - Installer Script
# Host link: https://raw.githubusercontent.com/princev89/stay-awake/main/install.sh
# Installation Command:
#   curl -fsSL https://raw.githubusercontent.com/princev89/stay-awake/main/install.sh | bash

set -e

# Color codes for pretty terminal logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}⚡ Starting Stay Awake Installer...${NC}"

# 1. Verify that the operating system is macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
  echo -e "${RED}❌ Error: Stay Awake is a macOS-only application and cannot be installed on $OSTYPE.${NC}"
  exit 1
fi

# 2. Install using Homebrew Cask if Homebrew is installed
if command -v brew &> /dev/null; then
  echo -e "${GREEN}🍺 Homebrew detected. Installing via Homebrew Cask...${NC}"
  
  # Run the combined tap & install command
  brew install --cask princev89/tap/stay-awake
  
  echo -e "${GREEN}🎉 Stay Awake installed successfully via Homebrew!${NC}"
else
  # 3. Fallback: Manual install from GitHub Releases (for non-developer users without Homebrew)
  echo -e "${YELLOW}⚠️ Homebrew not detected. Installing manually from GitHub Releases...${NC}"
  
  VERSION="1.0.1"
  DOWNLOAD_URL="https://github.com/princev89/stay-awake/releases/download/v${VERSION}/Stay.Awake.zip"
  TEMP_DIR=$(mktemp -d)
  
  echo -e "${GREEN}📥 Downloading Stay Awake v${VERSION}...${NC}"
  curl -L -s -o "$TEMP_DIR/Stay.Awake.zip" "$DOWNLOAD_URL"
  
  echo -e "${GREEN}📦 Extracting application bundle...${NC}"
  unzip -q "$TEMP_DIR/Stay.Awake.zip" -d "$TEMP_DIR"
  
  echo -e "${GREEN}🚀 Moving Stay Awake.app to your Applications folder...${NC}"
  
  # Clear existing installation if it exists
  if [ -d "/Applications/Stay Awake.app" ]; then
    rm -rf "/Applications/Stay Awake.app"
  fi
  
  mv "$TEMP_DIR/Stay Awake.app" "/Applications/"
  
  # Clean up temporary folders
  rm -rf "$TEMP_DIR"
  
  echo -e "${GREEN}🎉 Stay Awake successfully installed in /Applications!${NC}"
fi

echo -e "${GREEN}✅ Done! You can now launch Stay Awake from your Applications or via Spotlight search.${NC}"
