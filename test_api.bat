@echo off
REM 3D Model Generation API - Automated Test Script (Windows)
REM This script tests all backend API endpoints

setlocal enabledelayedexpansion

REM Configuration
set "BASE_URL=http://localhost:8080"
set "TOTAL_TESTS=0"
set "PASSED_TESTS=0"
set "FAILED_TESTS=0"

echo ==================================================
echo   3D Model Generation API - Automated Tests
echo ==================================================
echo.

REM Check if curl is available
curl --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] curl is required but not installed.
    echo Please install curl from https://curl.se/windows/
    exit /b 1
)

REM Check if server is running
echo [INFO] Checking if server is running at %BASE_URL%...
curl -s -o nul -w "%%{http_code}" "%BASE_URL%/app/config" | findstr "200" >nul
if errorlevel 1 (
    echo [ERROR] Server is not running at %BASE_URL%
    echo Please start the backend server: iac-test.exe
    exit /b 1
)
echo [SUCCESS] Server is running
echo.

echo Starting API Tests...
echo.

REM ================================================
REM TEST 1: Health Check
REM ================================================
set /a TOTAL_TESTS+=1
echo [TEST %TOTAL_TESTS%] Health Check - GET /app/config
curl -s -o nul -w "%%{http_code}" "%BASE_URL%/app/config" | findstr "200" >nul
if errorlevel 1 (
    echo [FAILED] Health check failed
    set /a FAILED_TESTS+=1
) else (
    echo [PASSED] Health check passed
    set /a PASSED_TESTS+=1
)
echo.

REM ================================================
REM TEST 2: Create Text-to-3D Generation Job
REM ================================================
set /a TOTAL_TESTS+=1
echo [TEST %TOTAL_TESTS%] Create Text-to-3D Generation Job
curl -s -X POST "%BASE_URL%/3dmodels/generate/text" ^
    -H "Content-Type: application/json" ^
    -d "{\"prompt\": \"A test model from automated script\"}" > text_response.json

REM Extract job ID (simple parsing for Windows)
findstr "\"id\"" text_response.json >nul
if errorlevel 1 (
    echo [FAILED] Failed to create text-to-3D job
    type text_response.json
    set /a FAILED_TESTS+=1
    set "TEXT_JOB_ID="
) else (
    echo [PASSED] Text-to-3D job created
    set /a PASSED_TESTS+=1
    REM Note: You may need jq or similar to extract the actual ID
    echo Response saved to text_response.json
)
echo.

REM ================================================
REM TEST 3: Create Image-to-3D Generation Job
REM ================================================
set /a TOTAL_TESTS+=1
echo [TEST %TOTAL_TESTS%] Create Image-to-3D Generation Job

REM Minimal 1x1 red PNG in base64
set "MINIMAL_IMAGE=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="

curl -s -X POST "%BASE_URL%/3dmodels/generate/image" ^
    -H "Content-Type: application/json" ^
    -d "{\"imageData\": \"%MINIMAL_IMAGE%\", \"prompt\": \"Test image model\"}" > image_response.json

findstr "\"id\"" image_response.json >nul
if errorlevel 1 (
    echo [FAILED] Failed to create image-to-3D job
    type image_response.json
    set /a FAILED_TESTS+=1
) else (
    echo [PASSED] Image-to-3D job created
    set /a PASSED_TESTS+=1
    echo Response saved to image_response.json
)
echo.

REM ================================================
REM TEST 4: List All Models
REM ================================================
set /a TOTAL_TESTS+=1
echo [TEST %TOTAL_TESTS%] List all models - POST /3dmodels/list
curl -s -X POST "%BASE_URL%/3dmodels/list" ^
    -H "Content-Type: application/json" ^
    -d "{}" > list_response.json

findstr "\"data\"" list_response.json >nul
if errorlevel 1 (
    echo [FAILED] Failed to list models
    set /a FAILED_TESTS+=1
) else (
    echo [PASSED] Models listed successfully
    set /a PASSED_TESTS+=1
    echo Response saved to list_response.json
)
echo.

REM ================================================
REM Wait and check status
REM ================================================
echo [INFO] Waiting 15 seconds for generation to complete...
timeout /t 15 /nobreak >nul
echo.

REM ================================================
REM TEST SUMMARY
REM ================================================
echo ==================================================
echo   Test Summary
echo ==================================================
echo.
echo Total Tests:  %TOTAL_TESTS%
echo Passed:       %PASSED_TESTS%
echo Failed:       %FAILED_TESTS%
echo.

set /a SUCCESS_RATE=(%PASSED_TESTS% * 100) / %TOTAL_TESTS%
echo Success Rate: %SUCCESS_RATE%%%
echo.

if %FAILED_TESTS% EQU 0 (
    echo [SUCCESS] All tests passed!
    exit /b 0
) else (
    echo [ERROR] Some tests failed. Please review the output above.
    exit /b 1
)

REM Cleanup
del text_response.json image_response.json list_response.json >nul 2>&1
