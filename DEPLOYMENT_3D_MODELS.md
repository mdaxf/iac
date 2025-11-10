# 3D Model Generation System - Deployment Guide

This guide provides step-by-step instructions for deploying the complete 3D Model Generation system.

## Table of Contents
1. [System Requirements](#system-requirements)
2. [Backend Deployment](#backend-deployment)
3. [Frontend Deployment](#frontend-deployment)
4. [Database Configuration](#database-configuration)
5. [Production Configuration](#production-configuration)
6. [AI Service Integration](#ai-service-integration)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)

---

## System Requirements

### Minimum Requirements
- **OS**: Windows Server 2019+, Linux (Ubuntu 20.04+), macOS 10.15+
- **CPU**: 2 cores
- **RAM**: 4GB
- **Disk**: 10GB free space
- **Database**: MongoDB 4.4+
- **Go**: 1.19+ (for building from source)
- **Node.js**: 16+ (for frontend)

### Recommended Requirements
- **CPU**: 4+ cores
- **RAM**: 8GB+
- **Disk**: 50GB+ (for model storage)
- **Database**: MongoDB 5.0+ with replica set
- **Network**: 100 Mbps+

---

## Backend Deployment

### 1. Prerequisites

**Install MongoDB:**
```bash
# Ubuntu/Debian
sudo apt-get install -y mongodb-org

# Windows
# Download from https://www.mongodb.com/try/download/community

# macOS
brew install mongodb-community
```

**Start MongoDB:**
```bash
# Linux
sudo systemctl start mongod
sudo systemctl enable mongod

# Windows
net start MongoDB

# macOS
brew services start mongodb-community
```

### 2. Build Backend

**From Source:**
```bash
cd c:\working\projects\iac\iac-main

# Build for production
go build -o iac.exe -ldflags="-s -w"

# Or use existing build
# iac-test.exe or iac.exe
```

**Binary Size Optimization:**
```bash
# Build with compression
go build -o iac.exe -ldflags="-s -w" -trimpath

# Using UPX (optional)
upx --best --lzma iac.exe
```

### 3. Configuration

**Create config.json:**
```json
{
  "Port": 8080,
  "Timeout": 30000,
  "MongoDB": {
    "Host": "localhost",
    "Port": 27017,
    "Database": "iac_production",
    "Username": "",
    "Password": "",
    "ReplicaSet": ""
  },
  "LogLevel": "info",
  "StoragePath": "./storage/3d_models",
  "MaxUploadSize": 10485760,
  "CORS": {
    "AllowOrigins": ["https://yourdomain.com"],
    "AllowMethods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    "AllowHeaders": ["Content-Type", "Authorization"]
  }
}
```

**Environment Variables (Alternative):**
```bash
export IAC_PORT=8080
export IAC_MONGODB_HOST=localhost
export IAC_MONGODB_PORT=27017
export IAC_MONGODB_DATABASE=iac_production
export IAC_LOG_LEVEL=info
```

### 4. Database Setup

**Create Database and Collection:**
```javascript
// Connect to MongoDB
mongosh

// Create database
use iac_production

// Create 3D_Models collection with validation
db.createCollection("3D_Models", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["name", "type", "status", "progress", "createdOn"],
      properties: {
        name: { bsonType: "string" },
        type: {
          enum: ["text-to-3d", "image-to-3d", "manual"]
        },
        status: {
          enum: ["pending", "processing", "completed", "failed"]
        },
        progress: {
          bsonType: "int",
          minimum: 0,
          maximum: 100
        }
      }
    }
  }
})

// Create indexes
db["3D_Models"].createIndex({ "createdOn": -1 })
db["3D_Models"].createIndex({ "status": 1 })
db["3D_Models"].createIndex({ "type": 1 })
db["3D_Models"].createIndex({ "generatedBy": 1 })

// Verify
db.getCollectionInfos({ name: "3D_Models" })
```

### 5. Create Storage Directory

```bash
# Create storage directory with proper permissions
mkdir -p ./storage/3d_models
chmod 755 ./storage/3d_models

# Windows
mkdir storage\3d_models
```

### 6. Run Backend

**Development:**
```bash
./iac-test.exe
```

**Production (Linux/macOS with systemd):**

Create `/etc/systemd/system/iac-backend.service`:
```ini
[Unit]
Description=IAC 3D Model Generation Backend
After=network.target mongod.service

[Service]
Type=simple
User=iac
WorkingDirectory=/opt/iac
ExecStart=/opt/iac/iac.exe
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=iac-backend

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable iac-backend
sudo systemctl start iac-backend
sudo systemctl status iac-backend
```

**Production (Windows Service):**

Using NSSM (Non-Sucking Service Manager):
```cmd
nssm install IAC-Backend "C:\iac\iac.exe"
nssm set IAC-Backend AppDirectory "C:\iac"
nssm set IAC-Backend Start SERVICE_AUTO_START
nssm start IAC-Backend
```

**Using Docker:**

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o iac.exe -ldflags="-s -w"

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/iac.exe .
COPY config.json .
EXPOSE 8080
CMD ["./iac.exe"]
```

Build and run:
```bash
docker build -t iac-backend:latest .
docker run -d -p 8080:8080 \
  -v $(pwd)/storage:/root/storage \
  --name iac-backend \
  iac-backend:latest
```

### 7. Verify Backend

```bash
# Check if server is running
curl http://localhost:8080/app/config

# Check health
curl http://localhost:8080/app/debug

# Test API
curl -X POST http://localhost:8080/3dmodels/list \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## Frontend Deployment

### 1. Build Frontend

```bash
cd c:\working\projects\iac\iac-portal

# Install dependencies
npm install

# Build for production
npm run build
```

Output will be in `dist/` directory.

### 2. Configure API Endpoint

**Update `.env.production`:**
```env
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_API_TIMEOUT=30000
```

**Or update `src/services/api/client.ts`:**
```typescript
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'https://api.yourdomain.com',
  timeout: 30000,
})
```

### 3. Deploy Frontend

**Option 1: Nginx**

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    root /var/www/iac-portal;
    index index.html;

    # Enable gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # SPA routing
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Model storage
    location /storage/ {
        proxy_pass http://localhost:8080/storage/;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

**Option 2: Apache**

```apache
<VirtualHost *:80>
    ServerName yourdomain.com
    DocumentRoot /var/www/iac-portal

    <Directory /var/www/iac-portal>
        Options -Indexes +FollowSymLinks
        AllowOverride All
        Require all granted

        # SPA routing
        RewriteEngine On
        RewriteBase /
        RewriteRule ^index\.html$ - [L]
        RewriteCond %{REQUEST_FILENAME} !-f
        RewriteCond %{REQUEST_FILENAME} !-d
        RewriteRule . /index.html [L]
    </Directory>

    # API proxy
    ProxyPass /api/ http://localhost:8080/
    ProxyPassReverse /api/ http://localhost:8080/

    # Model storage
    ProxyPass /storage/ http://localhost:8080/storage/
    ProxyPassReverse /storage/ http://localhost:8080/storage/
</VirtualHost>
```

**Option 3: Vercel/Netlify**

Create `vercel.json` or `netlify.toml`:

```json
// vercel.json
{
  "rewrites": [
    { "source": "/api/:path*", "destination": "https://api.yourdomain.com/:path*" },
    { "source": "/(.*)", "destination": "/index.html" }
  ],
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-XSS-Protection", "value": "1; mode=block" }
      ]
    }
  ]
}
```

### 4. SSL/TLS Configuration

**Using Let's Encrypt (Certbot):**
```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d yourdomain.com

# Auto-renewal
sudo certbot renew --dry-run
```

---

## Production Configuration

### 1. Security Hardening

**Backend Security:**

- Enable HTTPS only
- Use strong MongoDB authentication
- Implement rate limiting
- Add request validation
- Enable CORS with specific origins
- Use secure headers

**Add to main.go:**
```go
// Security middleware
router.Use(func(c *gin.Context) {
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    c.Next()
})

// Rate limiting
// Install: go get github.com/ulule/limiter/v3
// Add rate limiting middleware
```

### 2. File Storage Configuration

**Local Storage (Production):**
```go
// config.json
{
  "StoragePath": "/var/lib/iac/3d_models",
  "MaxStorageSize": 107374182400  // 100GB
}
```

**Cloud Storage (AWS S3):**
```go
// Modify saveGeneratedModel() in models3d.go
// Use AWS SDK to upload to S3
// Return S3 URL instead of local path

import "github.com/aws/aws-sdk-go/service/s3"

func (c *Models3DController) saveToS3(modelID string, data []byte) (string, error) {
    // Upload to S3
    // Return CloudFront URL
}
```

### 3. Logging Configuration

**Structured Logging:**
```go
// Use structured logging
logger.Info("Model generated",
    "modelId", modelID,
    "type", "text-to-3d",
    "duration", duration,
    "fileSize", fileSize)
```

**Log Rotation:**
```bash
# Install logrotate
sudo apt-get install logrotate

# Create /etc/logrotate.d/iac
/var/log/iac/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 iac iac
    sharedscripts
    postrotate
        systemctl reload iac-backend
    endscript
}
```

### 4. Monitoring Setup

**Prometheus Metrics:**
```go
// Add metrics endpoint
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

**Health Check Endpoint:**
```go
router.GET("/health", func(c *gin.Context) {
    // Check MongoDB connection
    // Check storage availability
    // Return status
    c.JSON(200, gin.H{
        "status": "healthy",
        "mongodb": "connected",
        "storage": "available"
    })
})
```

---

## AI Service Integration

### Option 1: Meshy AI

```go
// Add to generateWithAIService()
import "net/http"

func (c *Models3DController) callMeshyAI(prompt string) ([]byte, error) {
    apiKey := os.Getenv("MESHY_API_KEY")

    // 1. Create generation task
    payload := map[string]interface{}{
        "mode": "preview",
        "prompt": prompt,
        "art_style": "realistic",
        "negative_prompt": ""
    }

    resp, err := http.Post(
        "https://api.meshy.ai/v2/text-to-3d",
        "application/json",
        bytes.NewBuffer(jsonPayload))

    // Extract task ID
    taskID := resp.Data.Result

    // 2. Poll for completion
    for {
        statusResp, _ := http.Get(
            fmt.Sprintf("https://api.meshy.ai/v2/text-to-3d/%s", taskID))

        if statusResp.Status == "SUCCEEDED" {
            // Download model
            modelURL := statusResp.ModelURLs.GLB
            return downloadFile(modelURL)
        }

        time.Sleep(5 * time.Second)
    }
}
```

### Option 2: Zoo ML API

```go
func (c *Models3DController) callZooML(prompt string) ([]byte, error) {
    apiKey := os.Getenv("ZOO_API_KEY")

    // Call Zoo ML text-to-CAD API
    // https://zoo.dev/docs/api/text-to-cad

    resp, err := http.Post(
        "https://zoo.dev/api/text-to-cad",
        "application/json",
        bytes.NewBuffer(jsonPayload))

    // Poll and download model
}
```

---

## Monitoring and Maintenance

### 1. Log Monitoring

```bash
# View backend logs
sudo journalctl -u iac-backend -f

# Search for errors
sudo journalctl -u iac-backend | grep ERROR

# Check specific time period
sudo journalctl -u iac-backend --since "1 hour ago"
```

### 2. Database Maintenance

```javascript
// Connect to MongoDB
mongosh

use iac_production

// Check collection stats
db["3D_Models"].stats()

// Count by status
db["3D_Models"].aggregate([
  { $group: { _id: "$status", count: { $sum: 1 } } }
])

// Clean up old failed jobs
db["3D_Models"].deleteMany({
  status: "failed",
  createdOn: { $lt: new Date(Date.now() - 30*24*60*60*1000) }
})

// Compact collection
db["3D_Models"].compact()
```

### 3. Storage Cleanup

```bash
# Find models older than 30 days
find ./storage/3d_models -name "*.glb" -mtime +30

# Delete old models
find ./storage/3d_models -name "*.glb" -mtime +30 -delete

# Check storage usage
du -sh ./storage/3d_models
```

### 4. Backup Strategy

**Database Backup:**
```bash
# Daily backup
mongodump --db iac_production --out /backup/mongodb/$(date +%Y%m%d)

# Restore
mongorestore --db iac_production /backup/mongodb/20251106/iac_production
```

**File Storage Backup:**
```bash
# Sync to backup location
rsync -av ./storage/3d_models/ /backup/models/

# Or to S3
aws s3 sync ./storage/3d_models/ s3://backup-bucket/3d_models/
```

### 5. Performance Tuning

**MongoDB Optimization:**
```javascript
// Add compound indexes
db["3D_Models"].createIndex({ "status": 1, "createdOn": -1 })
db["3D_Models"].createIndex({ "generatedBy": 1, "status": 1 })

// Enable profiling
db.setProfilingLevel(1, { slowms: 100 })

// View slow queries
db.system.profile.find().sort({ ts: -1 }).limit(5)
```

**Go Backend Optimization:**
```go
// Adjust goroutine pool size
runtime.GOMAXPROCS(runtime.NumCPU())

// Connection pooling
// Set MongoDB connection pool size in config
```

---

## Troubleshooting

### Common Issues

**1. High Memory Usage**
- Check for goroutine leaks
- Monitor MongoDB connections
- Review file storage size

**2. Slow Generation**
- Verify AI service response time
- Check network latency
- Review database query performance

**3. Storage Full**
- Implement automatic cleanup
- Move to cloud storage
- Add storage monitoring alerts

---

## Deployment Checklist

- [ ] MongoDB installed and configured
- [ ] Backend compiled and tested
- [ ] Configuration files created
- [ ] Storage directories created
- [ ] Database collections and indexes created
- [ ] Frontend built for production
- [ ] SSL certificates installed
- [ ] Reverse proxy configured
- [ ] Firewall rules configured
- [ ] Monitoring setup
- [ ] Backup strategy implemented
- [ ] Log rotation configured
- [ ] Health checks working
- [ ] Load testing completed
- [ ] Documentation updated

---

**Last Updated**: 2025-11-06
**Version**: 1.0
