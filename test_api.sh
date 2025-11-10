#!/bin/bash

# 3D Model Generation API - Automated Test Script
# This script tests all backend API endpoints

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_info "Test $TOTAL_TESTS: $test_name"

    if eval "$test_command"; then
        print_success "PASSED: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_error "FAILED: $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to check if server is running
check_server() {
    print_info "Checking if server is running at $BASE_URL..."
    if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/app/config" | grep -q "200"; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not running at $BASE_URL"
        print_warning "Please start the backend server: ./iac-test.exe"
        exit 1
    fi
}

# Function to wait for a condition
wait_for_condition() {
    local condition="$1"
    local timeout="$2"
    local message="$3"

    print_info "$message"

    for i in $(seq 1 $timeout); do
        if eval "$condition"; then
            return 0
        fi
        echo -n "."
        sleep 1
    done

    echo ""
    return 1
}

echo "=================================================="
echo "  3D Model Generation API - Automated Tests"
echo "=================================================="
echo ""

# Check if required tools are installed
command -v curl >/dev/null 2>&1 || { print_error "curl is required but not installed. Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { print_warning "jq is not installed. Some tests may not work properly."; }

# Check server
check_server

echo ""
echo "Starting API Tests..."
echo ""

# ================================================
# TEST 1: Health Check
# ================================================
run_test "Health Check - GET /app/config" \
    "curl -s -o /dev/null -w '%{http_code}' '$BASE_URL/app/config' | grep -q '200'"

# ================================================
# TEST 2: Create Text-to-3D Generation Job
# ================================================
echo ""
print_info "Creating text-to-3D generation job..."

TEXT_TO_3D_RESPONSE=$(curl -s -X POST "$BASE_URL/3dmodels/generate/text" \
    -H "Content-Type: application/json" \
    -d '{"prompt": "A test model from automated script"}')

if echo "$TEXT_TO_3D_RESPONSE" | jq -e '.data.id' > /dev/null 2>&1; then
    TEXT_JOB_ID=$(echo "$TEXT_TO_3D_RESPONSE" | jq -r '.data.id')
    print_success "Text-to-3D job created: $TEXT_JOB_ID"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_error "Failed to create text-to-3D job"
    print_error "Response: $TEXT_TO_3D_RESPONSE"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TEXT_JOB_ID=""
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# ================================================
# TEST 3: Create Image-to-3D Generation Job
# ================================================
echo ""
print_info "Creating image-to-3D generation job..."

# Create a minimal base64 image (1x1 pixel red PNG)
MINIMAL_IMAGE="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="

IMAGE_TO_3D_RESPONSE=$(curl -s -X POST "$BASE_URL/3dmodels/generate/image" \
    -H "Content-Type: application/json" \
    -d "{\"imageData\": \"$MINIMAL_IMAGE\", \"prompt\": \"Test image model\"}")

if echo "$IMAGE_TO_3D_RESPONSE" | jq -e '.data.id' > /dev/null 2>&1; then
    IMAGE_JOB_ID=$(echo "$IMAGE_TO_3D_RESPONSE" | jq -r '.data.id')
    print_success "Image-to-3D job created: $IMAGE_JOB_ID"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_error "Failed to create image-to-3D job"
    print_error "Response: $IMAGE_TO_3D_RESPONSE"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    IMAGE_JOB_ID=""
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# ================================================
# TEST 4: Poll Text-to-3D Job Status
# ================================================
if [ -n "$TEXT_JOB_ID" ]; then
    echo ""
    print_info "Polling text-to-3D job status (max 20 seconds)..."

    TEXT_JOB_COMPLETED=0
    for i in {1..10}; do
        STATUS_RESPONSE=$(curl -s "$BASE_URL/3dmodels/$TEXT_JOB_ID")
        STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status' 2>/dev/null || echo "error")
        PROGRESS=$(echo "$STATUS_RESPONSE" | jq -r '.data.progress' 2>/dev/null || echo "0")

        print_info "  Attempt $i: Status=$STATUS, Progress=$PROGRESS%"

        if [ "$STATUS" = "completed" ]; then
            print_success "Text-to-3D job completed!"
            TEXT_JOB_COMPLETED=1
            PASSED_TESTS=$((PASSED_TESTS + 1))
            break
        elif [ "$STATUS" = "failed" ]; then
            print_error "Text-to-3D job failed"
            ERROR_MSG=$(echo "$STATUS_RESPONSE" | jq -r '.data.error' 2>/dev/null || echo "Unknown error")
            print_error "Error: $ERROR_MSG"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            break
        fi

        sleep 2
    done

    if [ $TEXT_JOB_COMPLETED -eq 0 ] && [ "$STATUS" != "failed" ]; then
        print_warning "Text-to-3D job did not complete within timeout"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# ================================================
# TEST 5: Poll Image-to-3D Job Status
# ================================================
if [ -n "$IMAGE_JOB_ID" ]; then
    echo ""
    print_info "Polling image-to-3D job status (max 20 seconds)..."

    IMAGE_JOB_COMPLETED=0
    for i in {1..10}; do
        STATUS_RESPONSE=$(curl -s "$BASE_URL/3dmodels/$IMAGE_JOB_ID")
        STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status' 2>/dev/null || echo "error")
        PROGRESS=$(echo "$STATUS_RESPONSE" | jq -r '.data.progress' 2>/dev/null || echo "0")

        print_info "  Attempt $i: Status=$STATUS, Progress=$PROGRESS%"

        if [ "$STATUS" = "completed" ]; then
            print_success "Image-to-3D job completed!"
            IMAGE_JOB_COMPLETED=1
            MODEL_URL=$(echo "$STATUS_RESPONSE" | jq -r '.data.modelUrl' 2>/dev/null)
            print_info "  Model URL: $MODEL_URL"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            break
        elif [ "$STATUS" = "failed" ]; then
            print_error "Image-to-3D job failed"
            ERROR_MSG=$(echo "$STATUS_RESPONSE" | jq -r '.data.error' 2>/dev/null || echo "Unknown error")
            print_error "Error: $ERROR_MSG"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            break
        fi

        sleep 2
    done

    if [ $IMAGE_JOB_COMPLETED -eq 0 ] && [ "$STATUS" != "failed" ]; then
        print_warning "Image-to-3D job did not complete within timeout"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# ================================================
# TEST 6: Download Generated Model
# ================================================
if [ -n "$TEXT_JOB_ID" ] && [ $TEXT_JOB_COMPLETED -eq 1 ]; then
    echo ""
    print_info "Downloading generated model..."

    MODEL_FILE="test_model_$TEXT_JOB_ID.glb"
    HTTP_CODE=$(curl -s -o "$MODEL_FILE" -w "%{http_code}" "$BASE_URL/storage/3d_models/$TEXT_JOB_ID.glb")

    if [ "$HTTP_CODE" = "200" ] && [ -f "$MODEL_FILE" ]; then
        FILE_SIZE=$(wc -c < "$MODEL_FILE")
        if [ $FILE_SIZE -gt 0 ]; then
            print_success "Model downloaded successfully ($FILE_SIZE bytes)"
            PASSED_TESTS=$((PASSED_TESTS + 1))

            # Verify it's a GLB file (should start with "glTF")
            MAGIC=$(xxd -l 4 -p "$MODEL_FILE" 2>/dev/null || echo "")
            if [ "$MAGIC" = "676c5446" ]; then
                print_success "File is a valid GLB (magic: glTF)"
            else
                print_warning "File may not be a valid GLB (magic: $MAGIC)"
            fi

            # Clean up
            rm "$MODEL_FILE"
        else
            print_error "Downloaded file is empty"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        print_error "Failed to download model (HTTP $HTTP_CODE)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# ================================================
# TEST 7: List All Models
# ================================================
echo ""
run_test "List all models - POST /3dmodels/list" \
    "curl -s -X POST '$BASE_URL/3dmodels/list' -H 'Content-Type: application/json' -d '{}' | jq -e '.data' > /dev/null"

# ================================================
# TEST 8: Delete Text-to-3D Job
# ================================================
if [ -n "$TEXT_JOB_ID" ]; then
    echo ""
    run_test "Delete text-to-3D model - DELETE /3dmodels/$TEXT_JOB_ID" \
        "curl -s -o /dev/null -w '%{http_code}' -X DELETE '$BASE_URL/3dmodels/$TEXT_JOB_ID' | grep -q '200'"
fi

# ================================================
# TEST 9: Delete Image-to-3D Job
# ================================================
if [ -n "$IMAGE_JOB_ID" ]; then
    echo ""
    run_test "Delete image-to-3D model - DELETE /3dmodels/$IMAGE_JOB_ID" \
        "curl -s -o /dev/null -w '%{http_code}' -X DELETE '$BASE_URL/3dmodels/$IMAGE_JOB_ID' | grep -q '200'"
fi

# ================================================
# TEST SUMMARY
# ================================================
echo ""
echo "=================================================="
echo "  Test Summary"
echo "=================================================="
echo ""
echo "Total Tests:  $TOTAL_TESTS"
echo "Passed:       $PASSED_TESTS"
echo "Failed:       $FAILED_TESTS"
echo ""

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo "Success Rate: $SUCCESS_RATE%"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    print_success "All tests passed! ✨"
    exit 0
else
    print_error "Some tests failed. Please review the output above."
    exit 1
fi
