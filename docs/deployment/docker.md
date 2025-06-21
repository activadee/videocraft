# =3 Docker Deployment

VideoCraft provides comprehensive Docker support for easy deployment and scaling. This guide covers Docker setup, Docker Compose configurations, and best practices for containerized deployments.

## =€ Quick Start

### Using Docker Compose (Recommended)
```bash
# Clone the repository
git clone https://github.com/your-org/videocraft.git
cd videocraft

# Start with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f videocraft

# Test API
curl http://localhost:3002/health
```

### Using Docker Run
```bash
# Pull the image
docker pull videocraft/videocraft:latest

# Run container
docker run -d \
  --name videocraft \
  -p 3002:3002 \
  -e VIDEOCRAFT_SECURITY_API_KEY="your-api-key" \
  -v $(pwd)/output:/app/output \
  videocraft/videocraft:latest
```

## =Ü Docker Compose Configuration

### Complete docker-compose.yml
```yaml
version: '3.8'

services:
  videocraft:
    image: videocraft/videocraft:latest
    container_name: videocraft
    restart: unless-stopped
    ports:
      - "3002:3002"      # API port
      - "9090:9090"      # Metrics port (optional)
    environment:
      # Server configuration
      VIDEOCRAFT_SERVER_HOST: "0.0.0.0"
      VIDEOCRAFT_SERVER_PORT: "3002"
      
      # Security configuration
      VIDEOCRAFT_SECURITY_ENABLE_AUTH: "true"
      VIDEOCRAFT_SECURITY_API_KEY: "${VIDEOCRAFT_API_KEY}"
      VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS: "${ALLOWED_DOMAINS}"
      VIDEOCRAFT_SECURITY_ENABLE_CSRF: "true"
      VIDEOCRAFT_SECURITY_CSRF_SECRET: "${CSRF_SECRET}"
      
      # Python/Whisper configuration
      VIDEOCRAFT_PYTHON_PATH: "/usr/bin/python3"
      VIDEOCRAFT_PYTHON_WHISPER_MODEL: "base"
      VIDEOCRAFT_PYTHON_WHISPER_DEVICE: "cpu"
      
      # FFmpeg configuration
      VIDEOCRAFT_FFMPEG_PATH: "/usr/bin/ffmpeg"
      VIDEOCRAFT_FFMPEG_TIMEOUT: "3600"
      
      # Storage configuration
      VIDEOCRAFT_STORAGE_OUTPUT_DIR: "/app/output"
      VIDEOCRAFT_STORAGE_TEMP_DIR: "/app/temp"
      
      # Logging
      VIDEOCRAFT_LOGGING_LEVEL: "info"
      VIDEOCRAFT_LOGGING_FORMAT: "json"
    volumes:
      # Persistent storage for generated videos
      - ./output:/app/output
      - ./temp:/app/temp
      # Optional: Custom configuration
      - ./config:/app/config:ro
      # Optional: Custom scripts
      - ./scripts:/app/scripts:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3002/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - videocraft-network
    depends_on:
      - redis  # Optional: for job queue

  # Optional: Redis for job queue
  redis:
    image: redis:7-alpine
    container_name: videocraft-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - videocraft-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 5s
      retries: 3

  # Optional: Nginx reverse proxy
  nginx:
    image: nginx:alpine
    container_name: videocraft-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    networks:
      - videocraft-network
    depends_on:
      - videocraft

volumes:
  redis-data:
    driver: local

networks:
  videocraft-network:
    driver: bridge
```

### Environment Configuration
```bash
# .env file for Docker Compose
VIDEOCRAFT_API_KEY=your-secure-api-key-here
ALLOWED_DOMAINS=yourdomain.com,api.yourdomain.com
CSRF_SECRET=your-csrf-secret-here

# Optional: Custom image tag
VIDEOCRAFT_IMAGE_TAG=v1.2.0

# Optional: Resource limits
VIDEOCRAFT_MEMORY_LIMIT=2g
VIDEOCRAFT_CPU_LIMIT=1
```

## =¾ Storage and Volumes

### Volume Mapping Strategy
```yaml
services:
  videocraft:
    volumes:
      # Essential volumes
      - ./output:/app/output                    # Generated videos
      - ./temp:/app/temp                        # Temporary files
      
      # Configuration volumes
      - ./config/config.yaml:/app/config.yaml:ro  # Custom config
      - ./scripts:/app/scripts:ro                 # Custom scripts
      
      # Optional: Model cache
      - ./models:/app/models                       # Whisper models
      
      # Optional: Logs
      - ./logs:/app/logs                           # Log files
```

### Persistent Storage Setup
```bash
# Create directories
mkdir -p output temp logs models config

# Set permissions
chmod 755 output temp logs models
chmod 644 config/*

# Create sample config
cat > config/config.yaml << 'EOF'
subtitles:
  enabled: true
  style: "progressive"
  font_family: "Arial"
  font_size: 24
EOF
```

## = Health Checks and Monitoring

### Health Check Configuration
```yaml
healthcheck:
  test: |
    curl -f http://localhost:3002/health || exit 1
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Advanced Health Check
```yaml
healthcheck:
  test: |
    #!/bin/bash
    # Check API health
    curl -f http://localhost:3002/health > /dev/null || exit 1
    
    # Check Python daemon
    pgrep -f whisper_daemon.py > /dev/null || exit 1
    
    # Check disk space
    df /app/output | awk 'NR==2 {if($5+0 > 90) exit 1}'
  interval: 30s
  timeout: 15s
  retries: 3
  start_period: 60s
```

### Container Monitoring
```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: videocraft-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    networks:
      - videocraft-network

  grafana:
    image: grafana/grafana:latest
    container_name: videocraft-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana:/etc/grafana/provisioning:ro
    networks:
      - videocraft-network

volumes:
  grafana-data:
```

## < Reverse Proxy Configuration

### Nginx Configuration
```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream videocraft {
        server videocraft:3002;
    }
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    
    server {
        listen 80;
        server_name yourdomain.com;
        
        # Redirect HTTP to HTTPS
        return 301 https://$server_name$request_uri;
    }
    
    server {
        listen 443 ssl http2;
        server_name yourdomain.com;
        
        # SSL configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256;
        
        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
        
        # API endpoints
        location /api/ {
            # Rate limiting
            limit_req zone=api burst=20 nodelay;
            
            # CORS headers
            add_header Access-Control-Allow-Origin "https://yourdomain.com";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-CSRF-Token";
            add_header Access-Control-Allow-Credentials true;
            
            # Handle preflight requests
            if ($request_method = 'OPTIONS') {
                add_header Access-Control-Allow-Origin "https://yourdomain.com";
                add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
                add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-CSRF-Token";
                add_header Access-Control-Max-Age 86400;
                return 204;
            }
            
            # Proxy settings
            proxy_pass http://videocraft;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts for long-running requests
            proxy_connect_timeout 60s;
            proxy_send_timeout 300s;
            proxy_read_timeout 300s;
        }
        
        # Health check endpoint
        location /health {
            proxy_pass http://videocraft;
            access_log off;
        }
        
        # File downloads
        location /downloads/ {
            alias /app/output/;
            add_header Content-Disposition "attachment";
        }
    }
}
```

### Traefik Configuration
```yaml
# docker-compose.traefik.yml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    container_name: videocraft-traefik
    command:
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.myresolver.acme.email=admin@yourdomain.com"
      - "--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Traefik dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./letsencrypt:/letsencrypt
    networks:
      - videocraft-network

  videocraft:
    # ... videocraft service config ...
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.videocraft.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.videocraft.entrypoints=websecure"
      - "traefik.http.routers.videocraft.tls.certresolver=myresolver"
      - "traefik.http.services.videocraft.loadbalancer.server.port=3002"
      
      # Rate limiting
      - "traefik.http.middlewares.api-ratelimit.ratelimit.burst=100"
      - "traefik.http.middlewares.api-ratelimit.ratelimit.average=10"
      - "traefik.http.routers.videocraft.middlewares=api-ratelimit"
```

## =á Security Configuration

### Docker Security Best Practices
```yaml
services:
  videocraft:
    # Use non-root user
    user: "1000:1000"
    
    # Read-only root filesystem
    read_only: true
    tmpfs:
      - /tmp
      - /app/temp
    
    # Security options
    security_opt:
      - no-new-privileges:true
    
    # Capability dropping
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE  # Only if binding to port < 1024
    
    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '0.5'
          memory: 1G
```

### Secrets Management
```yaml
secrets:
  api_key:
    file: ./secrets/api_key.txt
  csrf_secret:
    file: ./secrets/csrf_secret.txt

services:
  videocraft:
    secrets:
      - api_key
      - csrf_secret
    environment:
      - VIDEOCRAFT_SECURITY_API_KEY_FILE=/run/secrets/api_key
      - VIDEOCRAFT_SECURITY_CSRF_SECRET_FILE=/run/secrets/csrf_secret
```

## =€ Production Deployment

### Multi-Stage Production Setup
```bash
# Production deployment script
#!/bin/bash
set -e

echo "Deploying VideoCraft to production..."

# Backup current version
docker-compose down
cp -r output output.backup.$(date +%Y%m%d_%H%M%S)

# Pull latest images
docker-compose pull

# Update configuration
source .env.production
export VIDEOCRAFT_IMAGE_TAG=v1.2.0

# Start services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Wait for health check
echo "Waiting for service to be healthy..."
for i in {1..30}; do
  if curl -f http://localhost:3002/health > /dev/null 2>&1; then
    echo "Service is healthy!"
    break
  fi
  echo "Waiting... ($i/30)"
  sleep 10
done

# Run smoke tests
./scripts/smoke-test.sh

echo "Deployment completed successfully!"
```

### Docker Compose Override for Production
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  videocraft:
    # Production image
    image: videocraft/videocraft:${VIDEOCRAFT_IMAGE_TAG:-latest}
    
    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 8G
        reservations:
          cpus: '1.0'
          memory: 2G
      restart_policy:
        condition: on-failure
        max_attempts: 3
    
    # Production logging
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
    
    # Production environment
    environment:
      VIDEOCRAFT_LOGGING_LEVEL: "warn"
      VIDEOCRAFT_MONITORING_ENABLE_METRICS: "true"
      VIDEOCRAFT_SECURITY_RATE_LIMIT: "50"
```

## =Ê Scaling and Load Balancing

### Horizontal Scaling
```yaml
# docker-compose.scale.yml
version: '3.8'

services:
  videocraft:
    # Remove container_name for scaling
    # container_name: videocraft  # Remove this
    
    # Load balancer will handle port mapping
    expose:
      - "3002"
    # ports:  # Remove direct port mapping
    #   - "3002:3002"
    
    # Scale configuration
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 30s
        failure_action: rollback
      restart_policy:
        condition: on-failure
        max_attempts: 3

  # Load balancer
  haproxy:
    image: haproxy:alpine
    container_name: videocraft-lb
    ports:
      - "3002:3002"
    volumes:
      - ./haproxy/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    depends_on:
      - videocraft
    networks:
      - videocraft-network
```

### HAProxy Configuration
```
# haproxy/haproxy.cfg
global
    daemon
    log stdout len 65536 local0 info

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

frontend videocraft_frontend
    bind *:3002
    default_backend videocraft_backend

backend videocraft_backend
    balance roundrobin
    option httpchk GET /health
    server videocraft1 videocraft_videocraft_1:3002 check
    server videocraft2 videocraft_videocraft_2:3002 check
    server videocraft3 videocraft_videocraft_3:3002 check
```

## =Ë Maintenance and Updates

### Update Strategy
```bash
#!/bin/bash
# update-videocraft.sh

set -e

OLD_VERSION=$(docker-compose images videocraft --format "table {{.Tag}}" | tail -n +2)
NEW_VERSION=$1

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <new-version>"
    exit 1
fi

echo "Updating VideoCraft from $OLD_VERSION to $NEW_VERSION"

# Backup
echo "Creating backup..."
docker-compose exec videocraft tar -czf /app/output/backup-$(date +%Y%m%d_%H%M%S).tar.gz /app/output

# Update image tag
echo "Updating image tag..."
sed -i "s/VIDEOCRAFT_IMAGE_TAG=.*/VIDEOCRAFT_IMAGE_TAG=$NEW_VERSION/" .env

# Pull new image
echo "Pulling new image..."
docker-compose pull videocraft

# Rolling update
echo "Performing rolling update..."
docker-compose up -d videocraft

# Health check
echo "Waiting for health check..."
for i in {1..30}; do
    if curl -f http://localhost:3002/health; then
        echo "Update successful!"
        exit 0
    fi
    sleep 10
done

echo "Update failed, rolling back..."
sed -i "s/VIDEOCRAFT_IMAGE_TAG=.*/VIDEOCRAFT_IMAGE_TAG=$OLD_VERSION/" .env
docker-compose up -d videocraft
echo "Rollback completed"
exit 1
```

### Cleanup Script
```bash
#!/bin/bash
# cleanup-docker.sh

echo "Cleaning up Docker resources..."

# Remove unused containers
docker container prune -f

# Remove unused images
docker image prune -f

# Remove unused volumes (be careful!)
# docker volume prune -f

# Remove unused networks
docker network prune -f

# Show disk usage
docker system df

echo "Cleanup completed"
```

## =Ú Related Topics

### Deployment
- **[Kubernetes Deployment](kubernetes.md)** - Container orchestration
- **[Production Setup](production-setup.md)** - Production configuration
- **[Security Checklist](security-checklist.md)** - Security validation

### Configuration
- **[Configuration Overview](../configuration/overview.md)** - Configuration management
- **[Environment Variables](../configuration/environment-variables.md)** - Environment setup
- **[Performance Tuning](../configuration/performance-tuning.md)** - Optimization

### Monitoring
- **[Monitoring & Metrics](monitoring.md)** - Observability setup
- **[Logging](../reference/logging.md)** - Log management
- **[Troubleshooting](../reference/troubleshooting.md)** - Common issues

---

**= Next Steps**: [Kubernetes Deployment](kubernetes.md) | [Production Setup](production-setup.md) | [Monitoring Setup](monitoring.md)