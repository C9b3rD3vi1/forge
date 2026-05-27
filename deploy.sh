#!/bin/bash

# Go to project directory
cd /home/user/goapps/forge

# Pull latest changes from GitHub
git pull origin main

# Build Go binary
go build -o forge

# Stop old process (if running)
pkill forge || true

# Start the app in background
nohup ./forge >app.log 2>&1 &
