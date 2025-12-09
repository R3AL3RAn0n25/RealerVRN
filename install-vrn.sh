#!/bin/bash
set -e
dnf install -y golang git || apt install -y golang-go git
git clone https://github.com/yourname/VRN.git /opt/VRN
cd /opt/VRN
go mod tidy
go build -o /usr/local/bin/vrn ./cmd/vrn
mkdir -p /etc/vrn
cp config.example.toml /etc/vrn/config.toml
cp vrn.service /etc/systemd/system/vrn.service
systemctl daemon-reload
systemctl enable --now vrn
echo "VRN installed & running! Check: curl localhost:8080/health"