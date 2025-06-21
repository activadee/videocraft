# Production Setup Guide

This guide provides comprehensive instructions for deploying VideoCraft in production environments with proper security, monitoring, and performance optimization.

## =€ Production Deployment Overview

VideoCraft production deployment requires careful consideration of:
- **Security Configuration**: Authentication, CORS, CSRF protection
- **Performance Optimization**: Resource allocation and tuning
- **Infrastructure Setup**: Load balancing, monitoring, backup
- **Operational Procedures**: Deployment, updates, monitoring

## = Security Setup

### Required Security Configuration

```bash
# Authentication (REQUIRED in production)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="true"
export VIDEOCRAFT_SECURITY_API_KEY="$(openssl rand -hex 32)"

# CORS Domain Allowlisting (CRITICAL)
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,api.yourdomain.com"

# CSRF Protection (RECOMMENDED)
export VIDEOCRAFT_SECURITY_ENABLE_CSRF="true"
export VIDEOCRAFT_SECURITY_CSRF_SECRET="$(openssl rand -hex 32)"

# Rate Limiting
export VIDEOCRAFT_SECURITY_RATE_LIMIT="500"
```

### SSL/TLS Configuration

```nginx
# nginx.conf - HTTPS configuration
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    
    ssl_certificate /etc/ssl/certs/videocraft.crt;
    ssl_certificate_key /etc/ssl/private/videocraft.key;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    location / {
        proxy_pass http://127.0.0.1:3002;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for video processing
        proxy_connect_timeout 60s;
        proxy_send_timeout 600s;
        proxy_read_timeout 600s;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

## ™ Performance Configuration

### Production Environment Variables

```bash
# Server Configuration
export VIDEOCRAFT_SERVER_HOST="127.0.0.1"  # Bind to localhost (behind proxy)
export VIDEOCRAFT_SERVER_PORT="3002"

# Performance Optimization
export VIDEOCRAFT_JOB_WORKERS="16"
export VIDEOCRAFT_JOB_MAX_CONCURRENT="32"
export VIDEOCRAFT_JOB_QUEUE_SIZE="500"

# FFmpeg Optimization
export VIDEOCRAFT_FFMPEG_TIMEOUT="3h"
export VIDEOCRAFT_FFMPEG_QUALITY="23"
export VIDEOCRAFT_FFMPEG_PRESET="medium"

# Transcription Optimization
export VIDEOCRAFT_TRANSCRIPTION_DAEMON_ENABLED="true"
export VIDEOCRAFT_TRANSCRIPTION_DAEMON_IDLE_TIMEOUT="900s"
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="base"
export VIDEOCRAFT_TRANSCRIPTION_PROCESSING_WORKERS="8"

# Storage Configuration
export VIDEOCRAFT_STORAGE_OUTPUT_DIR="/var/videocraft/output"
export VIDEOCRAFT_STORAGE_TEMP_DIR="/var/videocraft/temp"
export VIDEOCRAFT_STORAGE_CLEANUP_INTERVAL="30m"
export VIDEOCRAFT_STORAGE_RETENTION_DAYS="7"

# Logging
export VIDEOCRAFT_LOG_LEVEL="info"
export VIDEOCRAFT_LOG_FORMAT="json"
```

### Resource Requirements

**Minimum Production Requirements:**
- **CPU**: 4 cores
- **Memory**: 8GB RAM
- **Storage**: 50GB SSD (for OS and application)
- **Temp Storage**: 100GB+ fast storage for video processing
- **Network**: 100 Mbps bandwidth

**Recommended Production Setup:**
- **CPU**: 16+ cores
- **Memory**: 32GB+ RAM
- **Storage**: 100GB SSD (system) + 500GB+ NVMe (processing)
- **Network**: 1 Gbps bandwidth

## =3 Docker Production Deployment

### Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  videocraft:
    image: videocraft:latest
    restart: unless-stopped
    
    environment:
      # Security
      VIDEOCRAFT_SECURITY_ENABLE_AUTH: "true"
      VIDEOCRAFT_SECURITY_API_KEY: "${VIDEOCRAFT_API_KEY}"
      VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS: "${ALLOWED_DOMAINS}"
      VIDEOCRAFT_SECURITY_ENABLE_CSRF: "true"
      VIDEOCRAFT_SECURITY_CSRF_SECRET: "${CSRF_SECRET}"
      VIDEOCRAFT_SECURITY_RATE_LIMIT: "500"
      
      # Performance
      VIDEOCRAFT_JOB_WORKERS: "16"
      VIDEOCRAFT_JOB_MAX_CONCURRENT: "32"
      VIDEOCRAFT_FFMPEG_PRESET: "medium"
      VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL: "base"
      
      # Storage
      VIDEOCRAFT_STORAGE_OUTPUT_DIR: "/app/output"
      VIDEOCRAFT_STORAGE_TEMP_DIR: "/app/temp"
      VIDEOCRAFT_STORAGE_RETENTION_DAYS: "7"
      
      # Logging
      VIDEOCRAFT_LOG_LEVEL: "info"
      VIDEOCRAFT_LOG_FORMAT: "json"
    
    volumes:
      - /var/videocraft/output:/app/output
      - /var/videocraft/temp:/app/temp
      - /var/videocraft/cache:/app/whisper_cache
      - /var/log/videocraft:/app/logs
    
    ports:
      - "127.0.0.1:3002:3002"
    
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3002/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    
    deploy:
      resources:
        limits:
          cpus: '16.0'
          memory: 32G
        reservations:
          cpus: '8.0'
          memory: 16G
    
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "10"
```

### Environment File (.env)

```bash
# .env - Production environment variables
VIDEOCRAFT_API_KEY=your-generated-api-key-here
ALLOWED_DOMAINS=yourdomain.com,api.yourdomain.com
CSRF_SECRET=your-generated-csrf-secret-here
```

## = Deployment Process

### Initial Deployment

```bash
# 1. Prepare production environment
sudo mkdir -p /var/videocraft/{output,temp,cache}
sudo mkdir -p /var/log/videocraft
sudo chown -R videocraft:videocraft /var/videocraft /var/log/videocraft

# 2. Generate secrets
echo "VIDEOCRAFT_API_KEY=$(openssl rand -hex 32)" > .env
echo "CSRF_SECRET=$(openssl rand -hex 32)" >> .env
echo "ALLOWED_DOMAINS=yourdomain.com,api.yourdomain.com" >> .env

# 3. Deploy with Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# 4. Verify deployment
curl -f https://api.yourdomain.com/health
```

### Update Deployment

```bash
# 1. Pull new image
docker pull videocraft:latest

# 2. Graceful update
docker-compose -f docker-compose.prod.yml up -d --no-deps videocraft

# 3. Verify update
curl -f https://api.yourdomain.com/health
docker logs videocraft --tail 100
```

### Rollback Procedure

```bash
# 1. Rollback to previous image
docker tag videocraft:latest videocraft:backup
docker pull videocraft:previous-version
docker tag videocraft:previous-version videocraft:latest

# 2. Restart service
docker-compose -f docker-compose.prod.yml up -d --no-deps videocraft

# 3. Verify rollback
curl -f https://api.yourdomain.com/health
```

## =Ê Monitoring and Observability

### Health Check Monitoring

```bash
# Basic monitoring script
#!/bin/bash
# health-monitor.sh

while true; do
    if ! curl -sf https://api.yourdomain.com/health > /dev/null; then
        echo "$(date): Health check failed" >> /var/log/videocraft/monitor.log
        # Send alert (email, Slack, etc.)
    fi
    sleep 30
done
```

### Log Management

```bash
# Log rotation configuration (/etc/logrotate.d/videocraft)
/var/log/videocraft/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 videocraft videocraft
    postrotate
        docker kill -s USR1 videocraft 2>/dev/null || true
    endscript
}
```

### Metrics Collection

```bash
# Prometheus metrics endpoint
curl https://api.yourdomain.com/metrics

# Key metrics to monitor:
# - videocraft_request_duration_seconds
# - videocraft_job_processing_seconds
# - videocraft_active_jobs
# - go_memstats_alloc_bytes
# - go_goroutines
```

## =¨ Alerting Configuration

### Critical Alerts

```yaml
# prometheus-alerts.yml
groups:
  - name: videocraft
    rules:
      - alert: VideoCraftDown
        expr: up{job="videocraft"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "VideoCraft service is down"
          
      - alert: HighMemoryUsage
        expr: (go_memstats_alloc_bytes / 1024 / 1024 / 1024) > 30
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage: {{ $value }}GB"
          
      - alert: HighJobQueue
        expr: videocraft_active_jobs{status="pending"} > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High job queue: {{ $value }} pending jobs"
```

## =á Security Hardening

### System Security

```bash
# Firewall configuration (UFW)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH (restrict to admin IPs)
sudo ufw allow 80/tcp    # HTTP redirect
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Fail2ban for SSH protection
sudo apt install fail2ban
sudo systemctl enable fail2ban
```

### Application Security

```bash
# Run as non-root user
sudo useradd -r -s /bin/false videocraft
sudo chown -R videocraft:videocraft /var/videocraft

# File permissions
sudo chmod 755 /var/videocraft
sudo chmod 750 /var/videocraft/output
sudo chmod 750 /var/videocraft/temp
sudo chmod 600 .env  # Protect environment file
```

### Regular Security Updates

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Update Docker images
docker pull videocraft:latest

# Security scanning
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image videocraft:latest
```

## =Ë Backup and Recovery

### Backup Strategy

```bash
#!/bin/bash
# backup.sh - Automated backup script

BACKUP_DIR="/backup/videocraft/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup configuration
cp .env "$BACKUP_DIR/"
cp docker-compose.prod.yml "$BACKUP_DIR/"

# Backup important data (exclude temp files)
tar -czf "$BACKUP_DIR/output.tar.gz" /var/videocraft/output/
tar -czf "$BACKUP_DIR/cache.tar.gz" /var/videocraft/cache/

# Backup logs (last 7 days)
find /var/log/videocraft -name "*.log" -mtime -7 -exec cp {} "$BACKUP_DIR/" \;

# Clean old backups (keep 30 days)
find /backup/videocraft -type d -mtime +30 -exec rm -rf {} +
```

### Recovery Procedures

```bash
# 1. Stop service
docker-compose -f docker-compose.prod.yml down

# 2. Restore data
cd /backup/videocraft/YYYYMMDD
tar -xzf output.tar.gz -C /
tar -xzf cache.tar.gz -C /

# 3. Restore configuration
cp .env /opt/videocraft/
cp docker-compose.prod.yml /opt/videocraft/

# 4. Restart service
cd /opt/videocraft
docker-compose -f docker-compose.prod.yml up -d

# 5. Verify recovery
curl -f https://api.yourdomain.com/health
```

## =' Maintenance Procedures

### Regular Maintenance Tasks

**Daily:**
- [ ] Check service health
- [ ] Monitor resource usage
- [ ] Review error logs

**Weekly:**
- [ ] Clean up old videos
- [ ] Review security logs
- [ ] Update security patches

**Monthly:**
- [ ] Update application
- [ ] Review and rotate API keys
- [ ] Performance optimization review
- [ ] Backup verification

### Maintenance Windows

```bash
# Graceful maintenance mode
# 1. Stop accepting new jobs
curl -X POST https://api.yourdomain.com/admin/maintenance-mode

# 2. Wait for current jobs to complete
while [ $(curl -s https://api.yourdomain.com/api/v1/jobs | jq '.jobs[] | select(.status=="processing") | length') -gt 0 ]; do
    echo "Waiting for jobs to complete..."
    sleep 30
done

# 3. Perform maintenance
docker-compose -f docker-compose.prod.yml down
# ... perform updates ...
docker-compose -f docker-compose.prod.yml up -d

# 4. Exit maintenance mode
curl -X DELETE https://api.yourdomain.com/admin/maintenance-mode
```

## =Þ Production Support

### Escalation Procedures

**Level 1: Service Issues**
- Check health endpoints
- Review recent logs
- Verify resource availability
- Restart service if needed

**Level 2: Performance Issues**
- Analyze metrics and logs
- Check resource constraints
- Review job queue status
- Scale resources if needed

**Level 3: Security Issues**
- Isolate affected systems
- Review security logs
- Contact security team
- Implement emergency procedures

### Emergency Contacts

- **Primary Support**: ops@yourdomain.com
- **Security Issues**: security@yourdomain.com
- **Infrastructure**: infra@yourdomain.com

## =Ú Additional Resources

- [Security Configuration Guide](../configuration/security-configuration.md)
- [Performance Optimization](../reference/performance.md)
- [Troubleshooting Guide](../reference/troubleshooting.md)
- [Docker Deployment Guide](docker.md)

This production setup guide ensures VideoCraft runs securely and efficiently in production environments.