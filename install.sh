#!/bin/bash

# --- Configuration ---
APP_USER="gitomatically"
APP_GROUP="gitomatically"
APP_DIR="/opt/gitomatically"
APP_BINARY="gitomatically"
ENV_FILE_NAME=".env"
CONFIG_FILE_NAME="config.yaml"
SERVICE_NAME="gitomatically.service"

# --- Add group and user ---
if ! id -g "$APP_GROUP" >/dev/null 2>&1; then
    sudo groupadd --system "$APP_GROUP"
    echo "Group '$APP_GROUP' created."
else
    echo "Group '$APP_GROUP' already exists."
fi

if ! id -u "$APP_USER" >/dev/null 2>&1; then
    sudo useradd --system --no-create-home --gid "$APP_GROUP" "$APP_USER"
    echo "User '$APP_USER' created."
else
    echo "User '$APP_USER' already exists."
fi

if getent group docker >/dev/null 2>&1; then
    if groups "$APP_USER" | grep -q '\bdocker\b'; then
        echo "User '$APP_USER' is already a member of the 'docker' group."
    else
        sudo usermod -aG docker "$APP_USER"
        echo "User '$APP_USER' added to 'docker' group."
        echo "NOTE: For changes to take full effect, you might need to restart the system or at least the systemd service after setup."
    fi
else
    echo "'docker' group does not exist. Is Docker installed? Skipping adding user to docker group."
fi

# --- Create app directory ---
sudo mkdir -p "$APP_DIR"
echo "Directory '$APP_DIR' created and permissions set."

# --- Build app ---
go build -o ./gitomatically ./main.go

# --- Copy files ---
sudo chmod +x "$APP_DIR"/gitomatically
sudo cp ./"$CONFIG_FILE_NAME" "$APP_DIR"/"$CONFIG_FILE_NAME"
sudo cp ./"$ENV_FILE_NAME" "$APP_DIR"/"$ENV_FILE_NAME"

# --- Change files owner ---
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

sudo systemctl daemon-reload
echo "Enabling service '$SERVICE_NAME' to start on boot..."
sudo systemctl enable "$SERVICE_NAME"
echo "Starting service '$SERVICE_NAME'..."
sudo systemctl start "$SERVICE_NAME"
echo "Setup complete! Checking service status:"
sudo systemctl status "$SERVICE_NAME" --no-pager
echo ""
echo "To view logs, run: sudo journalctl -u $SERVICE_NAME -f"