#!/bin/bash

# VMS Database Setup Script
# This script helps set up and test the MariaDB database connection

set -e

echo "🐳 Starting VMS Database Setup..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_status "Docker is running ✓"

# Navigate to project directory
cd "$(dirname "$0")"

# Start database service
print_status "Starting MariaDB database service..."
docker-compose up -d db

# Wait for database to be ready
print_status "Waiting for database to be ready..."
sleep 10

# Check if database is running
if docker-compose ps db | grep -q "Up"; then
    print_success "MariaDB database is running ✓"
else
    print_error "Failed to start MariaDB database"
    exit 1
fi

# Test database connection
print_status "Testing database connection..."
if docker-compose exec -T db mysql -u vms_user -pvms_password vms_db -e "SELECT 1;" > /dev/null 2>&1; then
    print_success "Database connection successful ✓"
else
    print_error "Database connection failed"
    print_status "Checking database logs..."
    docker-compose logs db
    exit 1
fi

# Run Go API to test full integration
print_status "Testing Go API with database connection..."
cd backend/golang-api

# Build and run the API briefly to test database connection
timeout 10s go run main.go > api_test.log 2>&1 &
API_PID=$!

sleep 3

if kill -0 $API_PID 2>/dev/null; then
    print_success "Go API started successfully ✓"
    kill $API_PID
else
    print_warning "Go API failed to start - check api_test.log for details"
    cat api_test.log
fi

print_success "Database setup completed!"
print_status "You can now run: docker-compose up -d"
print_status "And test the API at: http://localhost:8080/api/health"

# Cleanup
rm -f api_test.log