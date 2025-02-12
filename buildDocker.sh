#!/bin/bash

# Variables
IMAGE_NAME="forum"
CONTAINER_NAME="forum-container"
PORT="8000"

# Step 1: Build the Docker image
echo "Building Docker image..."
docker build -t $IMAGE_NAME .

# Check if the build was successful
if [ $? -ne 0 ]; then
  echo "Docker build failed. Exiting."
  exit 1
fi

# Step 2: Stop and remove any existing container with the same name
echo "Stopping and removing existing container..."
docker stop $CONTAINER_NAME > /dev/null 2>&1
docker rm $CONTAINER_NAME > /dev/null 2>&1

# Step 3: Run the Docker container
echo "Starting Docker container..."
docker run -d --name $CONTAINER_NAME -p $PORT:8000 $IMAGE_NAME

# Check if the container started successfully
if [ $? -eq 0 ]; then
  echo "Container started successfully!"
  echo "Application is running on http://localhost:$PORT"
else
  echo "Failed to start container."
  exit 1
fi