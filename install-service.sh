#!/bin/bash

set -e

SERVICE_NAME=mp3player
APP_DIR=$(pwd)/src/main
USER=$(whoami)

SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

echo "Creating systemd service..."

sudo bash -c "cat > $SERVICE_FILE" <<EOF
[Unit]
Description=Headless MP3 Player (Go)
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$APP_DIR
ExecStart=go run main.go
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

echo "Reloading systemd daemon..."
sudo systemctl daemon-reexec
sudo systemctl daemon-reload

echo "Enabling and starting $SERVICE_NAME.service..."
sudo systemctl enable --now "$SERVICE_NAME.service"

echo "Done. Status:"
sudo systemctl status "$SERVICE_NAME.service" --no-pager