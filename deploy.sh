#!/bin/bash

# Manual deployment script for forum application

set -e

STACK_NAME="forum-app"
REGION="${1:-us-east-1}"

echo "Deploying forum application to AWS..."

# Deploy CloudFormation stack
aws cloudformation deploy \
  --template-file cloudformation/forum-infrastructure.yml \
  --stack-name $STACK_NAME \
  --capabilities CAPABILITY_IAM \
  --region $REGION

# Get instance details
INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region $REGION)

INSTANCE_IP=$(aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --query 'Stacks[0].Outputs[?OutputKey==`InstancePublicIP`].OutputValue' \
  --output text \
  --region $REGION)

echo "Instance deployed: $INSTANCE_ID at $INSTANCE_IP"
echo "Application will be available at: http://$INSTANCE_IP:8000"
echo ""
echo "To connect via SSM:"
echo "aws ssm start-session --target $INSTANCE_ID"