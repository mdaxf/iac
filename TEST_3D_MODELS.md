# 3D Model Generation API - Testing Guide

This document provides comprehensive testing instructions for the 3D Model Generation system.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Backend API Testing](#backend-api-testing)
3. [Frontend Testing](#frontend-testing)
4. [Integration Testing](#integration-testing)
5. [Performance Testing](#performance-testing)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Services
- ✅ MongoDB running (default: mongodb://localhost:27017)
- ✅ Backend server (iac-test.exe or iac.exe)
- ✅ Frontend development server (npm run dev)

### Database Setup
```bash
# Create MongoDB collection
use iac_database
db.createCollection("3D_Models")

# Verify collection
db.getCollectionNames()
```

### Backend Configuration
Ensure `config.json` or environment variables are configured:
```json
{
  "Port": 8080,
  "MongoDB": {
    "Host": "localhost",
    "Port": 27017,
    "Database": "iac_database"
  }
}
```

---

## Backend API Testing

### 1. Start Backend Server

```bash
cd c:\working\projects\iac\iac-main
./iac-test.exe
```

Expected output:
```
[INFO] Starting IAC server on port 8080
[INFO] MongoDB connected successfully
[INFO] Controllers loaded: Models3DController
```

### 2. Test Health Endpoint

```bash
curl http://localhost:8080/app/config
```

Expected: Server configuration JSON

### 3. Test Text-to-3D Generation

**Create Generation Job:**
```bash
curl -X POST http://localhost:8080/3dmodels/generate/text \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A red cube with rounded edges"
  }'
```

Expected response:
```json
{
  "data": {
    "id": "673b8f2a1234567890abcdef",
    "status": "pending",
    "message": "Generation job created successfully"
  }
}
```

**Poll for Status:**
```bash
# Replace {id} with actual ID from previous response
curl http://localhost:8080/3dmodels/673b8f2a1234567890abcdef
```

Expected progression:
- Initial: `{"status": "pending", "progress": 0}`
- After 2s: `{"status": "processing", "progress": 20}`
- After 4s: `{"status": "processing", "progress": 40}`
- After 6s: `{"status": "processing", "progress": 60}`
- After 8s: `{"status": "processing", "progress": 80}`
- After 10s: `{"status": "completed", "progress": 100, "modelUrl": "/storage/..."}`

### 4. Test Image-to-3D Generation

**Prepare Base64 Image:**
```bash
# Convert image to base64
base64 -w 0 test_image.jpg > image_base64.txt
```

**Create Generation Job:**
```bash
curl -X POST http://localhost:8080/3dmodels/generate/image \
  -H "Content-Type: application/json" \
  -d '{
    "imageData": "data:image/jpeg;base64,/9j/4AAQSkZJRg...",
    "prompt": "Make it low-poly style"
  }'
```

Expected: Similar response to text-to-3D

### 5. Test List Models

```bash
curl -X POST http://localhost:8080/3dmodels/list \
  -H "Content-Type: application/json" \
  -d '{}'
```

Expected response:
```json
{
  "data": [
    {
      "_id": "673b8f2a1234567890abcdef",
      "name": "Text-to-3D: A red cube with rounded edges",
      "type": "text-to-3d",
      "status": "completed",
      "progress": 100,
      "modelUrl": "/storage/3d_models/673b8f2a1234567890abcdef.glb",
      "fileSize": 1024,
      "createdOn": "2025-11-06T12:00:00Z"
    }
  ]
}
```

### 6. Test Model Download

```bash
# Download generated model
curl http://localhost:8080/storage/3d_models/673b8f2a1234567890abcdef.glb \
  --output test_model.glb
```

Verify file:
```bash
# Check file size (should be > 0)
ls -lh test_model.glb

# Verify it's a valid GLB file (should start with "glTF")
xxd test_model.glb | head -1
# Expected: 00000000: 676c 5446 0200 0000 ....  "glTF"
```

### 7. Test Delete Model

```bash
curl -X DELETE http://localhost:8080/3dmodels/673b8f2a1234567890abcdef
```

Expected response:
```json
{
  "data": {
    "id": "673b8f2a1234567890abcdef",
    "message": "Model deleted successfully"
  }
}
```

---

## Frontend Testing

### 1. Start Frontend Development Server

```bash
cd c:\working\projects\iac\iac-portal
npm run dev
```

Expected output:
```
VITE v5.x.x ready in xxx ms
➜ Local:   http://localhost:5173/
```

### 2. Test Text-to-3D Interface

**Manual Testing Steps:**

1. Navigate to `http://localhost:5173/3d-designer`
2. Click "✨ Text to 3D" button
3. Verify UI elements:
   - ✅ Prompt textarea visible
   - ✅ Generate button enabled
   - ✅ 3D preview canvas visible
   - ✅ History section visible

4. Enter test prompt: "A simple wooden chair"
5. Click "Generate 3D Model"
6. Verify progress:
   - ✅ Status changes to "processing"
   - ✅ Progress bar appears and updates
   - ✅ Percentage updates (0% → 100%)

7. Wait for completion (~10-12 seconds)
8. Verify completion:
   - ✅ Status changes to "completed"
   - ✅ 3D model loads in preview
   - ✅ Can rotate model with mouse
   - ✅ Model added to history

### 3. Test Image-to-3D Interface

**Manual Testing Steps:**

1. From 3D Designer home, click "🖼️ Image to 3D"
2. Verify UI elements:
   - ✅ Upload dropzone visible
   - ✅ Optional prompt textarea
   - ✅ Generate button disabled (no image yet)
   - ✅ 3D preview canvas visible

3. Click upload area or drag image
4. Select test image (JPG/PNG)
5. Verify:
   - ✅ Image preview appears
   - ✅ Clear button visible
   - ✅ Generate button enabled

6. Optional: Add prompt "Low-poly style"
7. Click "Convert to 3D Model"
8. Verify progress tracking (same as text-to-3D)
9. Verify completion:
   - ✅ Model loads in preview
   - ✅ History shows thumbnail of input image

### 4. Test Error Handling

**Test Invalid Prompt:**
1. Leave prompt empty
2. Click generate
3. Expected: Alert "Please enter a description"

**Test Large Image:**
1. Try uploading 15MB image
2. Expected: Alert "Image file size must be less than 10MB"

**Test Invalid File Type:**
1. Try uploading .txt or .pdf file
2. Expected: Alert "Please select a valid image file"

---

## Integration Testing

### Test Complete Flow

**Scenario 1: Text-to-3D Complete Flow**

```bash
# 1. Start backend
cd c:\working\projects\iac\iac-main
./iac-test.exe &

# 2. Create generation job
RESPONSE=$(curl -s -X POST http://localhost:8080/3dmodels/generate/text \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Integration test model"}')

# 3. Extract job ID
JOB_ID=$(echo $RESPONSE | jq -r '.data.id')
echo "Job ID: $JOB_ID"

# 4. Poll for completion (max 30 seconds)
for i in {1..15}; do
  STATUS=$(curl -s http://localhost:8080/3dmodels/$JOB_ID | jq -r '.data.status')
  PROGRESS=$(curl -s http://localhost:8080/3dmodels/$JOB_ID | jq -r '.data.progress')
  echo "Attempt $i: Status=$STATUS, Progress=$PROGRESS%"

  if [ "$STATUS" = "completed" ]; then
    echo "✅ Generation completed!"
    break
  fi

  sleep 2
done

# 5. Download model
curl http://localhost:8080/storage/3d_models/$JOB_ID.glb --output integration_test.glb

# 6. Verify file exists
if [ -f integration_test.glb ]; then
  echo "✅ Model file downloaded successfully"
  ls -lh integration_test.glb
else
  echo "❌ Model file not found"
fi

# 7. Cleanup
curl -X DELETE http://localhost:8080/3dmodels/$JOB_ID
rm integration_test.glb
```

**Scenario 2: Multiple Concurrent Jobs**

```bash
# Create 3 jobs simultaneously
for i in {1..3}; do
  curl -X POST http://localhost:8080/3dmodels/generate/text \
    -H "Content-Type: application/json" \
    -d "{\"prompt\": \"Test model $i\"}" &
done
wait

# List all models
curl -X POST http://localhost:8080/3dmodels/list \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## Performance Testing

### Load Testing

**Test 1: Single Job Performance**
- Start timer
- Create generation job
- Poll until completion
- Measure total time

Expected: ~10-12 seconds

**Test 2: Concurrent Jobs**
- Create 5 jobs simultaneously
- Measure completion time for all

Expected: ~10-15 seconds (parallel processing)

**Test 3: API Response Time**
- Measure `/3dmodels/:id` GET response time
- Should be < 50ms

**Test 4: File Storage**
- Generate 10 models
- Check storage directory size
- Verify all files accessible

### Memory Testing

Monitor backend memory usage:
```bash
# Windows Task Manager or:
ps aux | grep iac-test

# Memory should remain stable (~40-50MB per goroutine)
```

---

## Troubleshooting

### Common Issues

**1. Backend won't start**
```
Error: Failed to connect to MongoDB
```
Solution:
- Verify MongoDB is running: `mongod --version`
- Check connection string in config

**2. Generation stuck at "pending"**
```
Status remains "pending" after 30 seconds
```
Solution:
- Check backend logs for errors
- Verify goroutine started: Look for "[Text-to-3D] Starting generation"
- Restart backend server

**3. Model file not found (404)**
```
GET /storage/3d_models/xxx.glb → 404
```
Solution:
- Check if `./storage/3d_models/` directory exists
- Verify file was created: `ls -la storage/3d_models/`
- Check static file serving route in main.go

**4. Frontend not connecting to backend**
```
Network error or CORS error
```
Solution:
- Verify backend is running on port 8080
- Check CORS middleware in main.go
- Update frontend API client base URL if needed

**5. Progress not updating**
```
Progress stays at 0%
```
Solution:
- Check MongoDB connection
- Verify `updateModelProgress()` is being called
- Check backend logs for update errors

### Debug Commands

**Check MongoDB data:**
```javascript
use iac_database
db["3D_Models"].find().pretty()
```

**Check storage directory:**
```bash
ls -lah c:\working\projects\iac\iac-main\storage\3d_models\
```

**Backend logs:**
```bash
# Run backend with verbose logging
./iac-test.exe --log-level debug
```

---

## Success Criteria

### Backend API Tests
- ✅ All 6 endpoints respond correctly
- ✅ Text-to-3D generation completes in ~10 seconds
- ✅ Image-to-3D generation completes in ~10 seconds
- ✅ Progress updates occur every 2 seconds
- ✅ GLB files are valid and downloadable
- ✅ MongoDB records are created and updated correctly

### Frontend Tests
- ✅ Text-to-3D UI works end-to-end
- ✅ Image-to-3D UI works end-to-end
- ✅ 3D models render in preview
- ✅ Progress tracking displays correctly
- ✅ History tracking works
- ✅ Error handling works for invalid inputs

### Integration Tests
- ✅ Complete flow from request to 3D preview works
- ✅ Multiple concurrent jobs work correctly
- ✅ File storage and serving works
- ✅ Cleanup (delete) works

---

## Test Report Template

```markdown
# 3D Model Generation - Test Report

**Date**: YYYY-MM-DD
**Tester**: [Name]
**Version**: Backend v1.0, Frontend v1.0

## Test Results Summary

| Category | Tests Passed | Tests Failed | Success Rate |
|----------|--------------|--------------|--------------|
| Backend API | X/7 | X/7 | X% |
| Frontend UI | X/8 | X/8 | X% |
| Integration | X/3 | X/3 | X% |
| Performance | X/4 | X/4 | X% |

## Detailed Results

### Backend API Tests
- [ ] POST /3dmodels/generate/text
- [ ] POST /3dmodels/generate/image
- [ ] GET /3dmodels/:id
- [ ] POST /3dmodels/list
- [ ] DELETE /3dmodels/:id
- [ ] GET /storage/3d_models/:file
- [ ] Progress tracking

### Frontend Tests
- [ ] Text-to-3D UI
- [ ] Image-to-3D UI
- [ ] 3D preview rendering
- [ ] Progress display
- [ ] History tracking
- [ ] Error handling
- [ ] File upload validation
- [ ] Navigation

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]
```

---

**Last Updated**: 2025-11-06
**Document Version**: 1.0
