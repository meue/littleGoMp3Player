#!/bin/bash

set -e

echo "=== Git und Go Installation auf Raspberry Pi ==="

# Update & Upgrade
echo "[1/5] Paketquellen aktualisieren..."
sudo apt update && sudo apt upgrade -y

# Git installieren
echo "[2/5] Git installieren..."
sudo apt install -y git

# Aktuelle Go-Version festlegen
GO_VERSION="1.22.4"
ARCH=$(dpkg --print-architecture)  # meist arm64 oder armhf

# Mapping von Architektur zu Go-Archiven
if [ "$ARCH" = "arm64" ]; then
    GO_ARCH="arm64"
elif [ "$ARCH" = "armhf" ]; then
    GO_ARCH="armv6l"
else
    echo "Nicht unterstützte Architektur: $ARCH"
    exit 1
fi

# Go herunterladen und entpacken
echo "[3/5] Go $GO_VERSION für $GO_ARCH wird heruntergeladen..."
wget -q https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz -O /tmp/go${GO_VERSION}.tar.gz

echo "[4/5] Entpacken und installieren..."
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go${GO_VERSION}.tar.gz

# PATH setzen (nur wenn noch nicht gesetzt)
GO_LINE='export PATH=$PATH:/usr/local/go/bin'
if ! grep -qF "$GO_LINE" ~/.profile; then
    echo "$GO_LINE" >> ~/.profile
    echo '[5/5] Go dem PATH hinzugefügt. Bitte führe `source ~/.profile` aus oder starte das Terminal neu.'
else
    echo "[5/5] Go war bereits im PATH enthalten."
fi

echo "[+] Erstelle symbolischen Link nach /usr/local/bin..."
sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

echo "✅ Installation abgeschlossen. Versionen:"
git --version
go version

# mp3 abspielen
sudo apt install mpg123

# download latest yt-dlp
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o src/main/yt-dlp
chmod a+rx src/main/yt-dlp

# mp3 converter
sudo apt install ffmpeg: