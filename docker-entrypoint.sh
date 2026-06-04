#!/bin/sh
set -e

# Copy template to actual config
cp /app/config/config.yaml.template /app/config/config.yaml

# Replace placeholders with environment variable values
sed -i "s/__DB_PASSWORD__/${DB_PASSWORD:-000000}/g" /app/config/config.yaml
sed -i "s/__REDIS_PASSWORD__/${REDIS_PASSWORD:-}/g" /app/config/config.yaml
sed -i "s/__LLM_API_KEY__/${LLM_API_KEY:-}/g" /app/config/config.yaml
sed -i "s/__WX_APP_SECRET__/${WX_APP_SECRET:-}/g" /app/config/config.yaml
sed -i "s/__AMAP_KEY__/${AMAP_KEY:-}/g" /app/config/config.yaml

# Ensure media directory exists and is writable
mkdir -p /app/media/letters

# Start the server
exec /app/server
