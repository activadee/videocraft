package security

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// DockerComposeService represents a service in docker-compose.yml
type DockerComposeService struct {
	SecurityOpt []string `yaml:"security_opt,omitempty"`
	ReadOnly    bool     `yaml:"read_only,omitempty"`
	CapDrop     []string `yaml:"cap_drop,omitempty"`
	CapAdd      []string `yaml:"cap_add,omitempty"`
	Privileged  bool     `yaml:"privileged,omitempty"`
	Deploy      struct {
		Resources struct {
			Limits struct {
				Memory string `yaml:"memory,omitempty"`
				CPUs   string `yaml:"cpus,omitempty"`
			} `yaml:"limits,omitempty"`
			Reservations struct {
				Memory string `yaml:"memory,omitempty"`
				CPUs   string `yaml:"cpus,omitempty"`
			} `yaml:"reservations,omitempty"`
		} `yaml:"resources,omitempty"`
	} `yaml:"deploy,omitempty"`
	User  string   `yaml:"user,omitempty"`
	Tmpfs []string `yaml:"tmpfs,omitempty"`
}

// DockerCompose represents the docker-compose.yml structure
type DockerCompose struct {
	Version  string                          `yaml:"version"`
	Services map[string]DockerComposeService `yaml:"services"`
}

// loadDockerCompose loads and parses docker-compose.yml
func loadDockerCompose(t *testing.T) *DockerCompose {
	data, err := os.ReadFile("../../docker-compose.yml")
	require.NoError(t, err, "Failed to read docker-compose.yml")

	var compose DockerCompose
	err = yaml.Unmarshal(data, &compose)
	require.NoError(t, err, "Failed to parse docker-compose.yml")

	return &compose
}

// TestContainerSecurityContext tests all security context requirements
func TestContainerSecurityContext(t *testing.T) {
	compose := loadDockerCompose(t)

	// Ensure videocraft service exists
	service, exists := compose.Services["videocraft"]
	require.True(t, exists, "videocraft service not found in docker-compose.yml")

	t.Run("NonRootUserExecution", func(t *testing.T) {
		// Test that container runs as non-root user
		// This should be enforced at Dockerfile level (USER videocraft)
		// But can also be enforced in docker-compose with user directive
		if service.User != "" {
			assert.NotEqual(t, "root", service.User, "Container should not run as root user")
			assert.NotEqual(t, "0", service.User, "Container should not run with UID 0")
		}
		// Note: If User is empty, it relies on Dockerfile USER directive
	})

	t.Run("SecurityContextsEnforced", func(t *testing.T) {
		// Test that security options are configured
		assert.NotEmpty(t, service.SecurityOpt, "Security options should be configured")

		// Check for specific security options
		securityOptString := strings.Join(service.SecurityOpt, " ")
		assert.Contains(t, securityOptString, "no-new-privileges",
			"no-new-privileges should be enabled")
	})

	t.Run("ReadOnlyRootFilesystem", func(t *testing.T) {
		// Test that read-only root filesystem is implemented
		assert.True(t, service.ReadOnly, "Read-only root filesystem should be enabled")
	})

	t.Run("CapabilitiesDropped", func(t *testing.T) {
		// Test that unnecessary capabilities are dropped
		assert.NotEmpty(t, service.CapDrop, "Capabilities should be dropped")

		// Check for specific dangerous capabilities that should be dropped
		capDropString := strings.Join(service.CapDrop, " ")
		dangerousCaps := []string{"ALL", "NET_RAW", "SYS_ADMIN", "SYS_PTRACE"}

		hasDroppedDangerousCaps := false
		for _, cap := range dangerousCaps {
			if strings.Contains(capDropString, cap) {
				hasDroppedDangerousCaps = true
				break
			}
		}
		assert.True(t, hasDroppedDangerousCaps,
			"Dangerous capabilities should be dropped")
	})

	t.Run("ResourceLimitsConfigured", func(t *testing.T) {
		// Test that resource limits are configured
		assert.NotEmpty(t, service.Deploy.Resources.Limits.Memory,
			"Memory limits should be configured")
		assert.NotEmpty(t, service.Deploy.Resources.Limits.CPUs,
			"CPU limits should be configured")
	})

	t.Run("NoPrivilegedMode", func(t *testing.T) {
		// Test that container does not run in privileged mode
		assert.False(t, service.Privileged, "Container should not run in privileged mode")
	})

	t.Run("TemporaryFilesystemForWritableDirectories", func(t *testing.T) {
		// Test that writable directories use tmpfs
		assert.NotEmpty(t, service.Tmpfs, "Tmpfs should be configured for writable directories")
	})
}

// TestDockerfileSecurityContext tests Dockerfile security configurations
func TestDockerfileSecurityContext(t *testing.T) {
	data, err := os.ReadFile("../../Dockerfile")
	require.NoError(t, err, "Failed to read Dockerfile")

	dockerfile := string(data)

	t.Run("NonRootUserInDockerfile", func(t *testing.T) {
		// Test that Dockerfile creates and uses non-root user
		assert.Contains(t, dockerfile, "adduser", "Dockerfile should create a non-root user")
		assert.Contains(t, dockerfile, "USER videocraft", "Dockerfile should switch to non-root user")
	})

	t.Run("MinimalBaseImage", func(t *testing.T) {
		// Test that production image uses minimal base (alpine)
		assert.Contains(t, dockerfile, "FROM alpine:latest",
			"Production image should use minimal alpine base")
	})

	t.Run("NoUnnecessaryPackages", func(t *testing.T) {
		// Test that minimal packages are installed
		lines := strings.Split(dockerfile, "\n")
		runtimePackages := []string{}

		inRuntimeSection := false
		for _, line := range lines {
			if strings.Contains(line, "# Final stage") {
				inRuntimeSection = true
				continue
			}
			if inRuntimeSection && strings.Contains(line, "apk add") {
				runtimePackages = append(runtimePackages, line)
			}
		}

		// Should not contain development tools in runtime
		for _, pkg := range runtimePackages {
			assert.NotContains(t, pkg, "gcc", "Runtime should not contain gcc")
			assert.NotContains(t, pkg, "git", "Runtime should not contain git")
			assert.NotContains(t, pkg, "musl-dev", "Runtime should not contain musl-dev")
		}
	})
}

// TestCurrentSecurityViolations tests current security violations (should fail initially)
func TestCurrentSecurityViolations(t *testing.T) {
	t.Run("ExpectedSecurityFailures", func(t *testing.T) {
		// These tests document current security issues and should fail initially
		// They will pass once security is properly implemented

		compose := loadDockerCompose(t)
		service := compose.Services["videocraft"]

		// This should fail initially - no security options configured
		assert.NotEmpty(t, service.SecurityOpt,
			"EXPECTED FAILURE: Security options not yet configured")

		// This should fail initially - read-only filesystem not configured
		assert.True(t, service.ReadOnly,
			"EXPECTED FAILURE: Read-only filesystem not yet configured")

		// This should fail initially - capabilities not dropped
		assert.NotEmpty(t, service.CapDrop,
			"EXPECTED FAILURE: Capabilities not yet dropped")

		// This should fail initially - resource limits not configured
		assert.NotEmpty(t, service.Deploy.Resources.Limits.Memory,
			"EXPECTED FAILURE: Resource limits not yet configured")
	})
}
