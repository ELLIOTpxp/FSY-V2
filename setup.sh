#!/data/data/com.termux/files/usr/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${RED}"
echo "███████ ███████ ██    ██"
echo "██      ██       ██  ██       "
echo "█████   ███████   ████"
echo "██           ██    ██"
echo "██      ███████    ██"
echo -e "${NC}"
echo -e "${YELLOW}• FSY-V2 TERMUX AUTO-SETUP •${NC}"
echo -e "${YELLOW}• Powering up...${NC}\n"
echo -e "${BLUE}[*] Updating Termux packages...${NC}"
pkg update -y && pkg upgrade -y

echo -e "${BLUE}[*] Installing essential packages...${NC}"
pkg install -y \
    golang \
    python \
    python-pip \
    git \
    wget \
    curl \
    unzip \
    make \
    cmake \
    clang \
    nodejs \
    npm \
    proot \
    proot-distro \
    tor \
    openssh

echo -e "${BLUE}[*] Installing Python dependencies...${NC}"
pip install --upgrade pip
pip install \
    colorama \
    requests \
    beautifulsoup4 \
    selenium \
    colorlog \
    termcolor

echo -e "${BLUE}[*] Installing Node.js dependencies...${NC}"
npm install -g \
    puppeteer \
    axios \
    cheerio \
    request \
    colors

echo -e "${BLUE}[*] Setting up Go environment...${NC}"
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
echo -e "${BLUE}[*] Installing Go dependencies...${NC}"
go install golang.org/x/net/http2@latest
go install golang.org/x/net/publicsuffix@latest
go install go.uber.org/zap@latest
go install github.com/chromedp/chromedp@latest
echo -e "${BLUE}[*] Creating project directory...${NC}"
mkdir -p ~/FSY-V2
cd ~/FSY-V2
mkdir -p proxies headers logs cookies
echo -e "${BLUE}[*] Downloading required files...${NC}"
echo -e "${YELLOW}[*] Downloading Tool files...${NC}"
git clone 

chmod +x launcher.py

echo -e "${GREEN}[+] Setup complete!${NC}"
echo -e "${YELLOW}[*] To start the launcher:${NC}"
echo -e "${BLUE}cd ~/FSY-V2${NC}"
echo -e "${BLUE}python3 launcher.py${NC}"