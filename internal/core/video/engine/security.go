package engine

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/activadee/videocraft/internal/api/models"
)

// Security validation constants and patterns
var (
	// Prohibited characters that could be used for command injection
	// Includes: semicolon, ampersand, pipe, backtick, dollar, parentheses, braces
	prohibitedCharsRegex = regexp.MustCompile(`[;|` + "`" + `$(){}]`)

	// Path traversal patterns for directory navigation attacks
	pathTraversalRegex = regexp.MustCompile(`\.\.\/|\.\.\\`)

	// Allowed URL protocols for security
	allowedProtocols = map[string]bool{
		"http":  true,
		"https": true,
	}

	// Dangerous commands that should be rejected after sanitization
	dangerousCommands = map[string]bool{
		"rm": true, "cat": true, "ls": true, "chmod": true, "chown": true,
		"sudo": true, "su": true, "bash": true, "sh": true, "cmd": true,
		"powershell": true, "wget": true, "curl": true, "nc": true, "netcat": true,
	}
)

// ValidateURL validates a URL for security threats using a multi-layered approach
func (s *service) ValidateURL(rawURL string) error {
	// Allow empty URLs (will be handled by other validation layers)
	if rawURL == "" {
		return nil
	}

	// Early rejection of data URIs (most common injection vector)
	if err := s.checkForDataURI(rawURL); err != nil {
		return err
	}

	// Character-based injection detection
	if err := s.checkForInjectionChars(rawURL); err != nil {
		return err
	}

	// Path traversal detection
	if err := s.checkForPathTraversal(rawURL); err != nil {
		return err
	}

	// URL structure validation and protocol checking
	return s.validateURLStructureAndProtocol(rawURL)
}

// checkForDataURI checks for data URI scheme which could bypass file restrictions
func (s *service) checkForDataURI(rawURL string) error {
	lowerURL := strings.ToLower(rawURL)

	// Check for dangerous URI schemes
	dangerousSchemes := []string{"data:", "javascript:", "vbscript:", "file:"}

	for _, scheme := range dangerousSchemes {
		if strings.HasPrefix(lowerURL, scheme) {
			s.logSecurityViolation("URL validation failed", map[string]interface{}{
				"url":            rawURL,
				"violation_type": "protocol_violation",
				"reason":         fmt.Sprintf("Protocol %s not allowed", scheme),
			})
			return errors.New("protocol not allowed")
		}
	}

	return nil
}

// checkForInjectionChars checks for characters that could enable command injection
func (s *service) checkForInjectionChars(rawURL string) error {
	if prohibitedCharsRegex.MatchString(rawURL) {
		s.logSecurityViolation("URL validation failed", map[string]interface{}{
			"url":            rawURL,
			"violation_type": "command_injection",
			"reason":         "URL contains prohibited characters",
		})
		return errors.New("URL contains prohibited characters")
	}
	return nil
}

// checkForPathTraversal checks for directory traversal attempts
func (s *service) checkForPathTraversal(rawURL string) error {
	if pathTraversalRegex.MatchString(rawURL) {
		s.logSecurityViolation("URL validation failed", map[string]interface{}{
			"url":            rawURL,
			"violation_type": "path_traversal",
			"reason":         "URL contains path traversal sequences",
		})
		return errors.New("URL contains path traversal sequences")
	}
	return nil
}

// validateURLStructureAndProtocol validates URL parsing and protocol allowlist
func (s *service) validateURLStructureAndProtocol(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if !allowedProtocols[parsedURL.Scheme] {
		s.logSecurityViolation("URL validation failed", map[string]interface{}{
			"url":            rawURL,
			"violation_type": "protocol_violation",
			"reason":         fmt.Sprintf("Protocol %s not allowed", parsedURL.Scheme),
		})
		return errors.New("protocol not allowed")
	}

	return nil
}

// SanitizeInput sanitizes input by removing dangerous characters
func (s *service) SanitizeInput(input string) (string, error) {
	original := input

	// Remove prohibited characters
	sanitized := prohibitedCharsRegex.ReplaceAllString(input, "")

	// Clean path traversal sequences
	sanitized = pathTraversalRegex.ReplaceAllString(sanitized, "")

	// Split by spaces and keep only the first token (before any command)
	tokens := strings.Fields(sanitized)
	if len(tokens) > 0 {
		sanitized = tokens[0]
	}

	// Remove extra whitespace
	sanitized = strings.TrimSpace(sanitized)

	// If the entire input was malicious content or only common command names, reject it
	if sanitized == "" && original != "" {
		return "", errors.New("input contains only malicious content")
	}

	// Additional check: reject if sanitized result is a common dangerous command
	if dangerousCommands[strings.ToLower(sanitized)] {
		return "", errors.New("input contains only malicious content")
	}

	return sanitized, nil
}

// ValidateURLAllowlist validates URL against domain allowlist
func (s *service) ValidateURLAllowlist(rawURL string) error {
	// If no allowlist is configured, allow all valid URLs
	if len(s.cfg.Security.AllowedDomains) == 0 {
		return nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if domain is in allowlist
	for _, allowedDomain := range s.cfg.Security.AllowedDomains {
		if parsedURL.Host == allowedDomain {
			return nil
		}
	}

	s.logSecurityViolation("Domain not in allowlist", map[string]interface{}{
		"url":            rawURL,
		"domain":         parsedURL.Host,
		"violation_type": "domain_not_allowed",
	})

	return errors.New("domain not in allowlist")
}

// validateAllURLsInConfig validates all URLs in a video configuration
// This is the main entry point for security validation during command building
func (s *service) validateAllURLsInConfig(config *models.VideoConfigArray) error {
	urlCount := 0

	for projectIdx, project := range *config {
		for sceneIdx, scene := range project.Scenes {
			for elementIdx, element := range scene.Elements {
				if element.Src != "" {
					urlCount++

					// Create context for better error reporting
					elementContext := fmt.Sprintf("project[%d].scene[%d].element[%d](%s)",
						projectIdx, sceneIdx, elementIdx, element.Type)

					// Basic URL validation
					if err := s.ValidateURL(element.Src); err != nil {
						return fmt.Errorf("security validation failed for %s: %w", elementContext, err)
					}

					// Domain allowlist validation
					if err := s.ValidateURLAllowlist(element.Src); err != nil {
						return fmt.Errorf("security validation failed for %s: %w", elementContext, err)
					}
				}
			}
		}
	}

	// Log successful validation for monitoring
	s.log.WithFields(map[string]interface{}{
		"urls_validated": urlCount,
		"projects":       len(*config),
	}).Info("All URLs passed security validation")

	return nil
}

// logSecurityViolation logs security violations with structured data
func (s *service) logSecurityViolation(message string, fields map[string]interface{}) {
	s.log.WithFields(fields).Errorf("SECURITY_VIOLATION: %s", message)
}
