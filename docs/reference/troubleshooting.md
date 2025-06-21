# Troubleshooting Guide

This comprehensive troubleshooting guide helps diagnose and resolve common issues with VideoCraft deployment, configuration, and operation.

## =¨ Quick Diagnosis

### Health Check Commands
```bash
# Basic service health
curl http://localhost:3002/health

# Detailed system status
curl http://localhost:3002/health/detailed

# Check specific service status
curl http://localhost:3002/ready
```

### Log Analysis
```bash
# View recent logs
docker logs videocraft -f

# Search for errors
docker logs videocraft 2>&1 | grep -i error

# Check specific error patterns
grep -E "(FAILED|ERROR|PANIC)" /var/log/videocraft/app.log
```

### Process Status
```bash
# Check if VideoCraft is running
ps aux | grep videocraft

# Check port binding
netstat -tulpn | grep :3002

# Monitor resource usage
top -p $(pgrep videocraft)
```

## =' Installation and Setup Issues

### Go Installation Problems

#### Issue: Go not found or wrong version
```bash
# Check Go installation
go version
which go

# Error: command not found
```

**Solutions:**
```bash
# Install Go (Ubuntu/Debian)
sudo apt update
sudo apt install golang-go

# Install Go (macOS)
brew install go

# Install Go (CentOS/RHEL)
sudo yum install golang

# Verify installation
go version
# Should show: go version go1.21.0 or higher
```

#### Issue: GOPATH/GOROOT configuration
```bash
# Check Go environment
go env GOROOT
go env GOPATH

# Set environment variables
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

### Python/Whisper Issues

#### Issue: Python dependencies missing
```bash
# Error: ModuleNotFoundError: No module named 'whisper'
```

**Solutions:**
```bash
# Install Python requirements
pip3 install -r scripts/requirements.txt

# If pip3 not found
sudo apt install python3-pip  # Ubuntu/Debian
brew install python3          # macOS

# Verify Whisper installation
python3 -c "import whisper; print('Whisper installed successfully')"

# Alternative: Use conda
conda install -c conda-forge openai-whisper
```

#### Issue: Whisper model download fails
```bash
# Error: Failed to download model
```

**Solutions:**
```bash
# Download model manually
python3 -c "import whisper; whisper.load_model('base')"

# Check internet connectivity
curl -I https://openaipublic.azureedge.net/main/whisper/models/

# Use different model if large models fail
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="tiny"
```

### FFmpeg Issues

#### Issue: FFmpeg not found
```bash
# Error: exec: "ffmpeg": executable file not found
```

**Solutions:**
```bash
# Install FFmpeg (Ubuntu/Debian)
sudo apt update
sudo apt install ffmpeg

# Install FFmpeg (macOS)
brew install ffmpeg

# Install FFmpeg (CentOS/RHEL)
sudo yum install epel-release
sudo yum install ffmpeg

# Verify installation
ffmpeg -version
which ffmpeg

# If in custom location, set path
export VIDEOCRAFT_FFMPEG_BINARY_PATH="/usr/local/bin/ffmpeg"
```

#### Issue: FFmpeg version incompatibility
```bash
# Check FFmpeg version
ffmpeg -version

# Minimum required: FFmpeg 4.0+
```

**Solutions:**
```bash
# Update FFmpeg (Ubuntu)
sudo apt update
sudo apt install ffmpeg

# Build from source if needed
wget https://ffmpeg.org/releases/ffmpeg-4.4.tar.bz2
tar xjf ffmpeg-4.4.tar.bz2
cd ffmpeg-4.4
./configure --enable-libx264
make && sudo make install
```

## =3 Docker Issues

### Container Startup Problems

#### Issue: Container exits immediately
```bash
# Check container logs
docker logs videocraft

# Common errors:
# - Configuration file not found
# - Permission denied
# - Port already in use
```

**Solutions:**
```bash
# Check configuration
docker run -it videocraft:latest cat config/config.yaml

# Fix permissions
sudo chown -R 1000:1000 ./generated_videos ./temp

# Check port availability
netstat -tulpn | grep :3002
lsof -i :3002

# Use different port
docker run -p 3003:3002 videocraft:latest
```

#### Issue: Volume mounting problems
```bash
# Error: Permission denied when writing to mounted volumes
```

**Solutions:**
```bash
# Fix directory permissions
sudo mkdir -p generated_videos temp whisper_cache
sudo chown -R $(id -u):$(id -g) generated_videos temp whisper_cache

# Use correct volume syntax
docker run -v $(pwd)/generated_videos:/app/generated_videos videocraft:latest

# For SELinux systems
sudo setsebool -P container_manage_cgroup true
```

### Docker Compose Issues

#### Issue: Service dependencies not ready
```bash
# Error: Connection refused when services start
```

**Solutions:**
```yaml
# docker-compose.yml - Add health checks
version: '3.8'
services:
  videocraft:
    image: videocraft:latest
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3002/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    depends_on:
      - redis  # if using external services
```

## ™ Configuration Issues

### Environment Variables

#### Issue: Configuration not loading
```bash
# Check environment variables
env | grep VIDEOCRAFT

# Verify configuration loading
curl http://localhost:3002/health/detailed | jq '.config'
```

**Solutions:**
```bash
# Set required environment variables
export VIDEOCRAFT_SERVER_HOST="0.0.0.0"
export VIDEOCRAFT_SERVER_PORT="3002"
export VIDEOCRAFT_FFMPEG_BINARY_PATH="ffmpeg"

# Create configuration file
cat > config.yaml << EOF
server:
  host: "0.0.0.0"
  port: 3002
ffmpeg:
  binary_path: "ffmpeg"
transcription:
  enabled: true
EOF

# Verify configuration
./videocraft --config config.yaml --validate
```

### Security Configuration

#### Issue: CORS errors in browser
```javascript
// Error: Access to fetch blocked by CORS policy
```

**Solutions:**
```bash
# Configure allowed domains
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,yourdomain.com"

# For development (not production!)
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"

# Restart service after configuration change
```

#### Issue: CSRF token errors
```bash
# Error: CSRF token required
```

**Solutions:**
```bash
# Get CSRF token first
curl http://localhost:3002/api/v1/csrf-token

# Include token in requests
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-token-here" \
  http://localhost:3002/api/v1/generate-video

# For development, disable CSRF
export VIDEOCRAFT_SECURITY_ENABLE_CSRF="false"
```

## <¬ Video Generation Issues

### Input Validation Errors

#### Issue: Invalid video configuration
```json
{
  "error": "Invalid configuration",
  "code": "INVALID_INPUT",
  "details": "Scene 'intro' is missing audio element"
}
```

**Solutions:**
```json
// Ensure each scene has at least one audio element
{
  "scenes": [
    {
      "id": "intro",
      "elements": [
        {
          "type": "audio",
          "src": "https://example.com/audio.mp3",
          "volume": 1.0
        }
      ]
    }
  ]
}

// Validate JSON structure before sending
// Use schema validation tools
```

#### Issue: File not found errors
```bash
# Error: FILE_NOT_FOUND for audio/video sources
```

**Solutions:**
```bash
# Test URL accessibility
curl -I https://example.com/audio.mp3

# Check file permissions for local files
ls -la /path/to/audio/file.mp3

# Use absolute URLs
# Ensure files are publicly accessible
# Check for authentication requirements

# Debug with wget/curl
wget https://example.com/audio.mp3 -O test.mp3
```

### Processing Failures

#### Issue: FFmpeg processing fails
```bash
# Error: FFMPEG_FAILED
```

**Solutions:**
```bash
# Check FFmpeg manually
ffmpeg -i input.mp3 -c:a copy output.wav

# Common fixes:
# 1. Verify input file format
file input.mp3

# 2. Check file size and length
ffprobe -i input.mp3 -show_format

# 3. Test with different codec
ffmpeg -i input.mp3 -c:a pcm_s16le output.wav

# 4. Increase timeout
export VIDEOCRAFT_FFMPEG_TIMEOUT="2h"

# 5. Reduce quality for faster processing
export VIDEOCRAFT_FFMPEG_QUALITY="28"
export VIDEOCRAFT_FFMPEG_PRESET="fast"
```

#### Issue: Transcription failures
```bash
# Error: TRANSCRIPTION_FAILED
```

**Solutions:**
```bash
# Test Whisper manually
python3 -c "
import whisper
model = whisper.load_model('base')
result = model.transcribe('test.mp3')
print(result['text'])
"

# Check audio file format
ffprobe -i audio.mp3

# Convert to supported format
ffmpeg -i input.mp4 -vn -acodec pcm_s16le -ar 16000 output.wav

# Try smaller model
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="tiny"

# Increase timeout
export VIDEOCRAFT_TRANSCRIPTION_PROCESSING_TIMEOUT="300s"
```

## = Network and Connectivity Issues

### External Resource Access

#### Issue: Cannot download external files
```bash
# Error: DOWNLOAD_FAILED
```

**Solutions:**
```bash
# Test network connectivity
ping google.com
curl -I https://example.com

# Check DNS resolution
nslookup example.com

# Test specific URL
curl -v https://example.com/audio.mp3

# Check for proxy/firewall issues
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Verify SSL certificates
curl -k https://example.com/audio.mp3  # Skip SSL verification (testing only)
```

#### Issue: Timeout errors
```bash
# Error: Request timeout
```

**Solutions:**
```bash
# Increase timeouts
export VIDEOCRAFT_FFMPEG_TIMEOUT="3h"
export VIDEOCRAFT_TRANSCRIPTION_PROCESSING_TIMEOUT="600s"

# Check network latency
ping -c 10 example.com

# Monitor bandwidth usage
iftop
nethogs

# Use local files instead of URLs for testing
cp audio.mp3 /app/temp/
# Reference as: file:///app/temp/audio.mp3
```

### API Connectivity

#### Issue: API not responding
```bash
# Error: Connection refused
```

**Solutions:**
```bash
# Check if service is running
ps aux | grep videocraft

# Check port binding
netstat -tulpn | grep :3002

# Test local connectivity
curl http://localhost:3002/health

# Check firewall rules
sudo ufw status
sudo iptables -L

# Start service if not running
./videocraft

# Bind to correct interface
export VIDEOCRAFT_SERVER_HOST="0.0.0.0"  # All interfaces
# export VIDEOCRAFT_SERVER_HOST="localhost"  # Local only
```

## =¾ Storage and File System Issues

### Disk Space Problems

#### Issue: Storage operation failed
```bash
# Error: STORAGE_FAILED - No space left on device
```

**Solutions:**
```bash
# Check disk space
df -h

# Find large files
du -sh ./generated_videos/*
du -sh ./temp/*

# Clean up old files
find ./generated_videos -mtime +7 -delete
find ./temp -mtime +1 -delete

# Configure automatic cleanup
export VIDEOCRAFT_STORAGE_CLEANUP_INTERVAL="15m"
export VIDEOCRAFT_STORAGE_RETENTION_DAYS="1"

# Move to larger disk
export VIDEOCRAFT_STORAGE_OUTPUT_DIR="/mnt/large-disk/videocraft"
```

### Permission Issues

#### Issue: Permission denied errors
```bash
# Error: Permission denied when writing files
```

**Solutions:**
```bash
# Check directory permissions
ls -la generated_videos/ temp/

# Fix permissions
sudo chown -R $(whoami):$(whoami) generated_videos temp
chmod 755 generated_videos temp

# For Docker containers
sudo chown -R 1000:1000 generated_videos temp

# Create directories if missing
mkdir -p generated_videos temp whisper_cache
```

## = Security and Authentication Issues

### Authentication Problems

#### Issue: Invalid API key
```bash
# Error: INVALID_API_KEY
```

**Solutions:**
```bash
# Check API key configuration
curl -H "Authorization: Bearer wrong-key" http://localhost:3002/health

# Generate new API key
openssl rand -hex 32

# Set correct API key
export VIDEOCRAFT_SECURITY_API_KEY="your-api-key-here"

# For development, disable auth
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="false"

# Verify configuration
curl http://localhost:3002/health/detailed | jq '.config'
```

### Rate Limiting

#### Issue: Too many requests
```bash
# Error: RATE_LIMIT_EXCEEDED
```

**Solutions:**
```bash
# Increase rate limit
export VIDEOCRAFT_SECURITY_RATE_LIMIT="1000"

# Implement client-side rate limiting
sleep 1  # Between requests

# Check current limits
curl -I http://localhost:3002/api/v1/jobs

# Response headers show limits:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 95
# X-RateLimit-Reset: 1705318260
```

## =Ê Performance Issues

### High Resource Usage

#### Issue: High memory usage
```bash
# Check memory usage
free -h
ps aux | grep videocraft | awk '{print $6}'
```

**Solutions:**
```bash
# Reduce concurrent jobs
export VIDEOCRAFT_JOB_MAX_CONCURRENT="4"

# Use smaller Whisper model
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="tiny"

# Increase cleanup frequency
export VIDEOCRAFT_STORAGE_CLEANUP_INTERVAL="5m"

# Monitor memory usage
curl http://localhost:3002/health/detailed | jq '.system.memory'
```

#### Issue: High CPU usage
```bash
# Monitor CPU usage
top -p $(pgrep videocraft)
```

**Solutions:**
```bash
# Reduce worker count
export VIDEOCRAFT_JOB_WORKERS="2"

# Use faster FFmpeg preset
export VIDEOCRAFT_FFMPEG_PRESET="ultrafast"

# Lower processing quality
export VIDEOCRAFT_FFMPEG_QUALITY="30"

# Check for CPU throttling
grep MHz /proc/cpuinfo
```

### Slow Processing

#### Issue: Video generation takes too long
```bash
# Monitor job status
curl http://localhost:3002/api/v1/jobs | jq '.jobs[] | select(.status=="processing")'
```

**Solutions:**
```bash
# Optimize FFmpeg settings
export VIDEOCRAFT_FFMPEG_PRESET="fast"      # vs "medium"
export VIDEOCRAFT_FFMPEG_QUALITY="25"       # vs "23" (lower quality, faster)

# Use smaller Whisper model
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="base"  # vs "small"

# Increase worker count (if resources allow)
export VIDEOCRAFT_JOB_WORKERS="8"

# Check for I/O bottlenecks
iostat -x 1

# Use SSD storage
export VIDEOCRAFT_STORAGE_OUTPUT_DIR="/fast-ssd/videocraft"
export VIDEOCRAFT_STORAGE_TEMP_DIR="/fast-ssd/temp"
```

## =' Debugging Tools

### Log Analysis Tools

```bash
# Real-time log monitoring
tail -f /var/log/videocraft/app.log

# Search for specific errors
grep "FFMPEG_FAILED" /var/log/videocraft/app.log

# Count error types
grep -o '"code":"[^"]*"' app.log | sort | uniq -c

# View structured logs
cat app.log | jq 'select(.level == "error")'

# Monitor specific job
grep "job_id:abc123" app.log | tail -20
```

### System Monitoring

```bash
# Monitor system resources
htop
iotop  # I/O usage
iftop  # Network usage

# Check system limits
ulimit -a

# Monitor file descriptors
lsof -p $(pgrep videocraft) | wc -l

# Check for memory leaks
valgrind --tool=memcheck ./videocraft
```

### Network Debugging

```bash
# Monitor network connections
netstat -tulpn | grep videocraft

# Capture network traffic
tcpdump -i any port 3002

# Test SSL/TLS connections
openssl s_client -connect example.com:443

# DNS debugging
dig example.com
nslookup example.com
```

## <˜ Emergency Procedures

### Service Recovery

```bash
# Quick service restart
systemctl restart videocraft
# or
docker restart videocraft

# Force kill if unresponsive
pkill -9 videocraft

# Clean restart with logs
./videocraft 2>&1 | tee videocraft.log

# Backup and restore configuration
cp config.yaml config.yaml.backup
# ... make changes ...
cp config.yaml.backup config.yaml  # restore if needed
```

### Data Recovery

```bash
# Recover failed jobs
find ./temp -name "*.partial" -exec mv {} ./generated_videos/ \;

# Check for corrupted files
find ./generated_videos -name "*.mp4" -exec ffprobe {} \; 2>&1 | grep "Invalid"

# Cleanup corrupted files
find ./generated_videos -size 0 -delete
```

### Performance Emergency

```bash
# Immediate resource relief
export VIDEOCRAFT_JOB_MAX_CONCURRENT="1"
export VIDEOCRAFT_FFMPEG_PRESET="ultrafast"
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="tiny"

# Stop all processing
curl -X POST http://localhost:3002/admin/stop-all-jobs

# Clear job queue
curl -X DELETE http://localhost:3002/admin/clear-queue
```

## =Þ Getting Help

### Information to Gather

When reporting issues, include:

1. **System Information**
   ```bash
   # System details
   uname -a
   go version
   ffmpeg -version
   python3 --version
   
   # VideoCraft version
   ./videocraft --version
   
   # Configuration
   curl http://localhost:3002/health/detailed
   ```

2. **Error Details**
   ```bash
   # Recent logs
   tail -100 /var/log/videocraft/app.log
   
   # Specific error
   grep -A 5 -B 5 "ERROR" app.log | tail -20
   
   # Resource usage
   free -h && df -h
   ```

3. **Reproduction Steps**
   - Exact API request that fails
   - Configuration used
   - Expected vs actual behavior
   - Frequency of occurrence

### Support Channels

- **Documentation**: Check this troubleshooting guide and other docs
- **GitHub Issues**: [Report bugs and issues](https://github.com/activadee/videocraft/issues)
- **Configuration Help**: Review [configuration documentation](../configuration/overview.md)
- **Performance Help**: See [performance guide](performance.md)

### Self-Help Checklist

Before seeking support:

- [ ] Check this troubleshooting guide
- [ ] Verify system requirements are met
- [ ] Review recent configuration changes
- [ ] Check logs for error messages
- [ ] Test with minimal configuration
- [ ] Verify network connectivity
- [ ] Try with different input files
- [ ] Check resource availability (CPU, memory, disk)

This troubleshooting guide covers the most common issues encountered with VideoCraft. For issues not covered here, please refer to the specific component documentation or seek support through the appropriate channels.