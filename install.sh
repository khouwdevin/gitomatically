#!/bin/bash

# --- Configuration ---
APP_USER=$(whoami)
APP_GROUP=$(whoami)
APP_DIR="/opt/gitomatically"
APP_BINARY="gitomatically"
ENV_FILE_NAME=".env"
CONFIG_FILE_NAME="config.yaml"
SERVICE_NAME="gitomatically.service"

# --- Create app directory ---
sudo mkdir -p "$APP_DIR"
echo "Directory '$APP_DIR' created and permissions set."

# --- Build app ---
go build -o ./gitomatically ./main.go

# --- Copy files ---
sudo cp ./"$APP_BINARY" "$APP_DIR"/"$APP_BINARY"
sudo ln -s ./"$CONFIG_FILE_NAME" "$APP_DIR"/"$CONFIG_FILE_NAME"
sudo ln -s ./"$ENV_FILE_NAME" "$APP_DIR"/"$ENV_FILE_NAME"

# --- Change files owner ---
sudo chmod +x "$APP_DIR"/"$APP_BINARY"
sudo chown -R "$APP_USER":"$APP_GROUP" "$APP_DIR"

sudo bash -c "cat > /etc/systemd/system/$SERVICE_NAME << EOF
[Unit]
Description=Gitomatically CI/CD Service
After=network.target

[Service]
Type=simple
User=$APP_USER
Group=$APP_GROUP
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$APP_BINARY
EnvironmentFile=$APP_DIR/$ENV_FILE_NAME
Restart=always
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=gitomatically

[Install]
WantedBy=multi-user.target
EOF"
echo "Service file '$SERVICE_NAME' created."

# --- Configure systemctl ---
sudo systemctl daemon-reload
echo "Enabling service '$SERVICE_NAME' to start on boot..."
sudo systemctl enable "$SERVICE_NAME"
echo "Starting service '$SERVICE_NAME'..."
sudo systemctl start "$SERVICE_NAME"
echo "Setup complete! Checking service status:"
sudo systemctl status "$SERVICE_NAME" --no-pager
echo ""
echo "To view logs, run: sudo journalctl -u $SERVICE_NAME -f"
