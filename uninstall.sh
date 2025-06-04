#!/bin/bash

# --- Configuration ---
APP_USER="gitomatically"
APP_GROUP="gitomatically"
APP_DIR="/opt/gitomatically"
APP_BINARY="gitomatically"
ENV_FILE_NAME=".env"
CONFIG_FILE_NAME="config.yaml"
SERVICE_NAME="gitomatically.service"

# --- Delete group and user ---
if id -g "$APP_GROUP" >/dev/null 2>&1; then
    sudo groupdel "$APP_GROUP"
    echo "Group '$APP_GROUP' is deleted."
else
    echo "Group '$APP_GROUP' is not exists."
fi

if id -u "$APP_USER" >/dev/null 2>&1; then
    sudo userdel "$APP_USER"
    echo "User '$APP_USER' is deleted."
else
    echo "User '$APP_USER' is not exists."
fi

# --- Remove files ---
sudo rm -rf "$APP_DIR"
sudo rm /etc/systemd/system/"$SERVICE_NAME"
echo "Directory '$APP_DIR' and '$SERVICE_NAME' is deleted"

# --- Remove configuration for systemctl ---
sudo systemctl stop "$SERVICE_NAME"
echo "Stopping '$SERVICE_NAME' service!"
sudo systemctl disable "$SERVICE_NAME"
echo "Disabling '$SERVICE_NAME' service!"
sudo systemctl daemon-reload
echo "Reload systemctl. Uninstall gitomatically success."