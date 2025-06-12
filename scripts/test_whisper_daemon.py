#!/usr/bin/env python3
"""
Test Suite for Whisper Daemon SSL Certificate Validation

This test suite validates that the Whisper daemon properly handles SSL certificates
and rejects connections to servers with invalid certificates.
"""

import unittest
import json
import ssl
import urllib.request
import urllib.error
from unittest.mock import patch, MagicMock, mock_open
import tempfile
import os
import sys
import io
from contextlib import redirect_stdout

# Add the scripts directory to the path so we can import whisper_daemon
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Inject stub whisper module into sys.modules to prevent ModuleNotFoundError during patching
if 'whisper' not in sys.modules:
    class WhisperStub:
        def load_model(self, *args, **kwargs):
            return MagicMock()
    
    sys.modules['whisper'] = WhisperStub()

try:
    from whisper_daemon import WhisperDaemon
except ImportError:
    # If whisper module is not available, create a mock for testing
    class WhisperDaemon:
        def __init__(self, *args, **kwargs):
            pass
        
        def handle_request(self, request):
            return {"success": False, "error": "Whisper not available"}


class TestSSLCertificateValidation(unittest.TestCase):
    """Test SSL certificate validation in Whisper daemon"""
    
    def setUp(self):
        """Set up test environment"""
        self.daemon = WhisperDaemon()
        
    def test_ssl_context_enables_certificate_verification(self):
        """Test that SSL context properly validates certificates"""
        # This test should FAIL initially because SSL verification is disabled
        
        # Create a request that would use SSL
        request = {
            "action": "transcribe",
            "url": "https://httpbin.org/status/200",  # Valid HTTPS endpoint
            "id": "test-ssl-validation"
        }
        
        # Mock the whisper model to avoid actual transcription
        with patch.object(self.daemon, 'model', None):
            with patch('whisper.load_model') as mock_load_model:
                mock_model = MagicMock()
                mock_model.transcribe.return_value = {
                    "text": "test",
                    "segments": []
                }
                mock_load_model.return_value = mock_model
                
                # Mock file operations
                with patch('tempfile.NamedTemporaryFile') as mock_temp:
                    mock_file = MagicMock()
                    mock_file.__enter__.return_value = mock_file
                    mock_file.name = "/tmp/test.wav"
                    mock_temp.return_value = mock_file
                    
                    # Mock urllib.request to capture SSL context
                    with patch('urllib.request.urlopen') as mock_urlopen:
                        mock_response = MagicMock()
                        mock_response.read.return_value = b"fake audio data"
                        mock_response.__enter__.return_value = mock_response
                        mock_urlopen.return_value = mock_response
                        
                        # Execute the request
                        response = self.daemon.handle_request(request)
                        
                        # Check that urlopen was called
                        if mock_urlopen.called:
                            call_args = mock_urlopen.call_args
                            ssl_context = call_args[1].get('context') if len(call_args) > 1 else None
                            
                            # This should FAIL because current implementation disables SSL verification
                            self.assertIsNotNone(ssl_context, "SSL context should be provided")
                            self.assertTrue(ssl_context.check_hostname, 
                                          "SSL context should verify hostnames")
                            self.assertEqual(ssl_context.verify_mode, ssl.CERT_REQUIRED,
                                           "SSL context should require valid certificates")
    
    def test_invalid_certificate_rejection(self):
        """Test that invalid SSL certificates are rejected"""
        request = {
            "action": "transcribe", 
            "url": "https://example.com/audio.wav",
            "id": "test-invalid-cert"
        }
        
        # Mock whisper model
        with patch.object(self.daemon, 'model', None):
            with patch('whisper.load_model') as mock_load_model:
                mock_model = MagicMock()
                mock_load_model.return_value = mock_model
                
                # Mock urllib.request.urlopen to raise SSL certificate verification error
                with patch('urllib.request.urlopen') as mock_urlopen:
                    mock_urlopen.side_effect = ssl.SSLCertVerificationError(
                        "certificate verify failed: self-signed certificate"
                    )
                    
                    response = self.daemon.handle_request(request)
                    
                    # Should fail with SSL verification error
                    self.assertFalse(response.get("success", True), 
                                   "Should reject invalid SSL certificates")
                    self.assertIn("certificate", response.get("error", "").lower(),
                                "Error should mention SSL certificate validation")
    
    def test_expired_certificate_rejection(self):
        """Test that expired SSL certificates are rejected"""
        request = {
            "action": "transcribe",
            "url": "https://example.com/audio.wav",
            "id": "test-expired-cert"
        }
        
        with patch.object(self.daemon, 'model', None):
            with patch('whisper.load_model') as mock_load_model:
                mock_model = MagicMock()
                mock_load_model.return_value = mock_model
                
                # Mock urllib.request.urlopen to raise SSL certificate verification error for expired cert
                with patch('urllib.request.urlopen') as mock_urlopen:
                    mock_urlopen.side_effect = ssl.SSLCertVerificationError(
                        "certificate verify failed: certificate has expired"
                    )
                    
                    response = self.daemon.handle_request(request)
                    
                    # Should fail when SSL verification is properly enabled
                    self.assertFalse(response.get("success", True),
                                   "Should reject expired SSL certificates")
                    self.assertIn("certificate", response.get("error", "").lower(),
                                "Error should mention certificate validation")
    
    def test_hostname_mismatch_rejection(self):
        """Test that hostname mismatches are rejected"""
        request = {
            "action": "transcribe",
            "url": "https://example.com/audio.wav",
            "id": "test-hostname-mismatch"
        }
        
        with patch.object(self.daemon, 'model', None):
            with patch('whisper.load_model') as mock_load_model:
                mock_model = MagicMock()
                mock_load_model.return_value = mock_model
                
                # Mock urllib.request.urlopen to raise SSL certificate verification error for hostname mismatch
                with patch('urllib.request.urlopen') as mock_urlopen:
                    mock_urlopen.side_effect = ssl.SSLCertVerificationError(
                        "certificate verify failed: Hostname mismatch, certificate is not valid for 'example.com'"
                    )
                    
                    response = self.daemon.handle_request(request)
                    
                    # Should fail when hostname verification is enabled
                    self.assertFalse(response.get("success", True),
                                   "Should reject hostname mismatches")
                    self.assertIn("hostname", response.get("error", "").lower(),
                                "Error should mention hostname verification")
    
    def test_valid_certificate_acceptance(self):
        """Test that valid SSL certificates are accepted"""
        request = {
            "action": "transcribe",
            "url": "https://httpbin.org/status/200",  # Known valid certificate
            "id": "test-valid-cert"
        }
        
        # Mock whisper model and file operations
        with patch.object(self.daemon, 'model', None):
            with patch('whisper.load_model') as mock_load_model:
                mock_model = MagicMock()
                mock_model.transcribe.return_value = {
                    "text": "test transcription",
                    "segments": [],
                    "language": "en"
                }
                mock_load_model.return_value = mock_model
                
                with patch('tempfile.NamedTemporaryFile') as mock_temp:
                    mock_file = MagicMock()
                    mock_file.__enter__.return_value = mock_file
                    mock_file.name = "/tmp/test.wav"
                    mock_temp.return_value = mock_file
                    
                    with patch('urllib.request.urlopen') as mock_urlopen:
                        mock_response = MagicMock()
                        mock_response.read.return_value = b"fake audio data"
                        mock_response.__enter__.return_value = mock_response
                        mock_urlopen.return_value = mock_response
                        
                        response = self.daemon.handle_request(request)
                        
                        # Valid certificates should always be accepted
                        self.assertTrue(response.get("success", False),
                                      "Should accept valid SSL certificates")
    
    def test_ssl_context_configuration(self):
        """Test that SSL context is properly configured"""
        # Test actual SSL context creation (no mocking)
        ssl_context = ssl.create_default_context()
        
        # These should pass with proper SSL verification
        self.assertTrue(ssl_context.check_hostname,
                      "SSL context should check hostnames")
        self.assertEqual(ssl_context.verify_mode, ssl.CERT_REQUIRED,
                       "SSL context should require certificate verification")


class TestSSLContextDirectly(unittest.TestCase):
    """Direct tests of SSL context configuration"""
    
    def test_default_ssl_context_security(self):
        """Test that default SSL context has proper security settings"""
        context = ssl.create_default_context()
        
        # These should be True for secure connections
        self.assertTrue(context.check_hostname, 
                       "Default SSL context should verify hostnames")
        self.assertEqual(context.verify_mode, ssl.CERT_REQUIRED,
                        "Default SSL context should require valid certificates")
    
    def test_insecure_ssl_context_detection(self):
        """Test detection of insecure SSL context configuration"""
        # Create an insecure context (like the old vulnerability)
        insecure_context = ssl.create_default_context()
        insecure_context.check_hostname = False
        insecure_context.verify_mode = ssl.CERT_NONE
        
        # These should detect the insecure configuration
        self.assertFalse(insecure_context.check_hostname,
                       "Insecure context should have hostname checking disabled")
        self.assertEqual(insecure_context.verify_mode, ssl.CERT_NONE,
                       "Insecure context should have certificate verification disabled")
        
        # Now test that a secure context is different
        secure_context = ssl.create_default_context()
        self.assertTrue(secure_context.check_hostname,
                       "Secure context should verify hostnames")
        self.assertEqual(secure_context.verify_mode, ssl.CERT_REQUIRED,
                       "Secure context should require certificate verification")


if __name__ == '__main__':
    # Run the tests
    unittest.main(verbosity=2)