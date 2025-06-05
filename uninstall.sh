#!/bin/bash

# --- Configuration ---
APP_USER="gitomatically"
APP_GROUP="gitomatically"
APP_DIR="/opt/gitomatically"
APP_BINARY="gitomatically"
ENV_FILE_NAME=".env"
CONFIG_FILE_NAME="config.yaml"
SERVICE_NAME="gitomatically.service"

# --- Remove configuration for systemctl ---
sudo systemctl stop "$SERVICE_NAME"
echo "Stopping '$SERVICE_NAME' service!"
sudo systemctl disable "$SERVICE_NAME"
echo "Disabling '$SERVICE_NAME' service!"
sudo systemctl daemon-reload

# --- Remove files ---
sudo rm -rf "$APP_DIR"
sudo rm /usr/local/bin/"$APP_BINARY"
sudo rm /etc/systemd/system/"$SERVICE_NAME"
echo "Directory '$APP_DIR' and '$SERVICE_NAME' is deleted"
echo "Uninstall gitomatically success."
