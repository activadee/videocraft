#!/usr/bin/env python3
"""
Test suite for URL validation security measures
Tests for SSRF vulnerability prevention in whisper_daemon.py
"""

import unittest
import ipaddress
from unittest.mock import patch, Mock
import urllib.parse

# URL validation module that we will implement
from url_validator import URLValidator, SSRFError


class TestURLValidation(unittest.TestCase):
    """Test cases for URL validation and SSRF prevention"""
    
    def setUp(self):
        """Set up test fixtures"""
        self.validator = URLValidator()
    
    @patch('socket.getaddrinfo')
    def test_valid_http_urls_allowed(self, mock_getaddrinfo):
        """Test that valid HTTP URLs are allowed"""
        # Mock DNS resolution to return a public IP
        mock_getaddrinfo.return_value = [
            (2, 1, 6, '', ('8.8.8.8', 80))  # Public IP
        ]
        
        valid_urls = [
            "http://example.com/audio.wav",
            "http://trusted-cdn.com/file.mp3",
            "http://api.service.com/download/audio",
        ]
        
        for url in valid_urls:
            with self.subTest(url=url):
                # Should not raise exception
                self.validator.validate_url(url)
    
    @patch('socket.getaddrinfo')
    def test_valid_https_urls_allowed(self, mock_getaddrinfo):
        """Test that valid HTTPS URLs are allowed"""
        # Mock DNS resolution to return a public IP
        mock_getaddrinfo.return_value = [
            (2, 1, 6, '', ('8.8.8.8', 443))  # Public IP
        ]
        
        valid_urls = [
            "https://example.com/audio.wav",
            "https://trusted-cdn.com/file.mp3", 
            "https://api.service.com/download/audio",
        ]
        
        for url in valid_urls:
            with self.subTest(url=url):
                # Should not raise exception
                self.validator.validate_url(url)
    
    def test_invalid_schemes_blocked(self):
        """Test that invalid URL schemes are blocked"""
        invalid_urls = [
            "file:///etc/passwd",
            "ftp://internal.server/file.wav",
            "gopher://old.server/audio",
            "ldap://internal.ldap/query",
            "dict://internal.dict/word",
            "sftp://internal.sftp/file.wav",
        ]
        
        for url in invalid_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertIn("not allowed", str(context.exception).lower())
    
    def test_private_ip_ranges_blocked(self):
        """Test that private IP ranges are blocked (RFC 1918)"""
        private_urls = [
            "http://10.0.0.1/audio.wav",
            "http://10.255.255.255/file.mp3",
            "http://172.16.0.1/audio.wav", 
            "http://172.31.255.255/file.mp3",
            "http://192.168.0.1/audio.wav",
            "http://192.168.255.255/file.mp3",
        ]
        
        for url in private_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertIn("private", str(context.exception).lower())
    
    def test_localhost_blocked(self):
        """Test that localhost and loopback addresses are blocked"""
        localhost_urls = [
            "http://localhost/audio.wav",
            "http://127.0.0.1/file.mp3",
            "http://127.255.255.255/audio.wav",
            "http://[::1]/file.mp3",
            "http://[::ffff:127.0.0.1]/audio.wav",
        ]
        
        for url in localhost_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertTrue("localhost" in str(context.exception).lower() or "loopback" in str(context.exception).lower())
    
    def test_link_local_blocked(self):
        """Test that link-local addresses are blocked"""
        link_local_urls = [
            "http://169.254.1.1/audio.wav",
            "http://169.254.255.254/metadata",  # AWS metadata endpoint
            "http://[fe80::1]/file.mp3",
        ]
        
        for url in link_local_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertTrue("link-local" in str(context.exception).lower() or "private" in str(context.exception).lower())
    
    def test_multicast_blocked(self):
        """Test that multicast addresses are blocked"""
        multicast_urls = [
            "http://224.0.0.1/audio.wav",
            "http://239.255.255.255/file.mp3",
            "http://[ff00::1]/file.mp3",
        ]
        
        for url in multicast_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertIn("multicast", str(context.exception).lower())
    
    def test_hostname_resolution_validation(self):
        """Test that hostname resolution is validated against private IPs"""
        # Mock DNS resolution to return private IP
        with patch('socket.getaddrinfo') as mock_getaddrinfo:
            mock_getaddrinfo.return_value = [
                (2, 1, 6, '', ('10.0.0.1', 80))  # Private IP
            ]
            
            with self.assertRaises(SSRFError) as context:
                self.validator.validate_url("http://evil-domain.com/audio.wav")
            self.assertIn("resolves to private", str(context.exception).lower())
    
    def test_url_allowlist_functionality(self):
        """Test URL allowlist functionality"""
        allowlist = ["trusted-cdn.com", "api.service.com"]
        validator_with_allowlist = URLValidator(allowed_domains=allowlist)
        
        # Allowed domain should pass
        validator_with_allowlist.validate_url("https://trusted-cdn.com/audio.wav")
        
        # Non-allowed domain should fail
        with self.assertRaises(SSRFError) as context:
            validator_with_allowlist.validate_url("https://untrusted.com/audio.wav")
        self.assertIn("not in allowlist", str(context.exception).lower())
    
    def test_url_parsing_edge_cases(self):
        """Test edge cases in URL parsing"""
        edge_case_urls = [
            "http://[::ffff:10.0.0.1]/audio.wav",  # IPv4-mapped IPv6
            "http://user:pass@127.0.0.1/audio.wav",  # URL with credentials
            "http://127.0.0.1:8080/audio.wav",  # URL with port
            "http://0x7f000001/audio.wav",  # Hex encoded localhost
            "http://2130706433/audio.wav",  # Decimal encoded localhost
        ]
        
        for url in edge_case_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError):
                    self.validator.validate_url(url)
    
    def test_request_timeout_configuration(self):
        """Test that request timeout can be configured"""
        validator = URLValidator(request_timeout=30)
        self.assertEqual(validator.request_timeout, 30)
        
        # Test default timeout
        default_validator = URLValidator()
        self.assertEqual(default_validator.request_timeout, 10)  # Default should be 10 seconds
    
    def test_malformed_urls_rejected(self):
        """Test that malformed URLs are rejected"""
        malformed_urls = [
            "",
            "not-a-url",
            "http://",
            "://missing-scheme.com",
            "http:///missing-host",
        ]
        
        for url in malformed_urls:
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertTrue(any(keyword in str(context.exception).lower() for keyword in ["malformed", "empty", "not allowed"]))
    
    def test_port_validation(self):
        """Test that dangerous ports are blocked"""
        dangerous_ports = [22, 23, 25, 53, 110, 143, 993, 995]  # SSH, Telnet, SMTP, DNS, etc.
        
        for port in dangerous_ports:
            url = f"http://example.com:{port}/audio.wav"
            with self.subTest(url=url):
                with self.assertRaises(SSRFError) as context:
                    self.validator.validate_url(url)
                self.assertIn("not allowed", str(context.exception).lower())
    
    @patch('socket.getaddrinfo')
    def test_safe_ports_allowed(self, mock_getaddrinfo):
        """Test that safe ports are allowed"""
        # Mock DNS resolution to return a public IP
        mock_getaddrinfo.return_value = [
            (2, 1, 6, '', ('8.8.8.8', 80))  # Public IP
        ]
        
        safe_ports = [80, 443, 8080, 8443]
        
        for port in safe_ports:
            url = f"http://example.com:{port}/audio.wav"
            with self.subTest(url=url):
                # Should not raise exception
                self.validator.validate_url(url)


class TestSSRFError(unittest.TestCase):
    """Test cases for SSRFError exception"""
    
    def test_ssrf_error_creation(self):
        """Test SSRFError exception creation"""
        error = SSRFError("Test message")
        self.assertEqual(str(error), "Test message")
        self.assertIsInstance(error, Exception)


class TestWhisperDaemonIntegration(unittest.TestCase):
    """Test integration of URL validation with WhisperDaemon"""
    
    @patch('urllib.request.urlopen')
    def test_transcribe_audio_validates_url(self, mock_urlopen):
        """Test that transcribe_audio validates URLs before processing"""
        from whisper_daemon import WhisperDaemon
        
        daemon = WhisperDaemon()
        
        # Test with private IP - should fail
        request = {
            "url": "http://10.0.0.1/audio.wav",
            "language": "auto",
            "word_timestamps": True
        }
        
        response = daemon.transcribe_audio(request)
        
        # Should return error response without calling urlopen
        self.assertFalse(response["success"])
        self.assertIn("private", response["error"].lower())
        mock_urlopen.assert_not_called()
    
    @patch('socket.getaddrinfo')
    @patch('urllib.request.urlopen')
    def test_transcribe_audio_allows_valid_url(self, mock_urlopen, mock_getaddrinfo):
        """Test that transcribe_audio allows valid URLs"""
        from whisper_daemon import WhisperDaemon
        
        # Mock DNS resolution to return a public IP
        mock_getaddrinfo.return_value = [
            (2, 1, 6, '', ('8.8.8.8', 443))  # Public IP
        ]
        
        # Mock successful HTTP response
        mock_response = Mock()
        mock_response.read.return_value = b"fake audio data"
        mock_urlopen.return_value.__enter__.return_value = mock_response
        
        daemon = WhisperDaemon()
        
        # Test with valid public URL
        request = {
            "url": "https://example.com/audio.wav",
            "language": "auto", 
            "word_timestamps": True
        }
        
        # Should not raise SSRF exception (may fail later due to invalid audio data)
        response = daemon.transcribe_audio(request)
        
        # URL validation should pass, allowing the request to proceed
        mock_urlopen.assert_called_once()


if __name__ == "__main__":
    unittest.main()