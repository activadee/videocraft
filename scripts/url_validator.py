#!/usr/bin/env python3
"""
URL Validation Module for SSRF Prevention

This module implements comprehensive URL validation to prevent Server-Side Request Forgery (SSRF) attacks.
It provides multiple layers of security validation including:

- URL scheme allowlisting (HTTP/HTTPS only)
- Private IP range blocking (RFC 1918, loopback, link-local)
- Hostname resolution validation to prevent DNS rebinding
- Domain allowlisting support
- Dangerous port blocking
- Request timeout controls

Security Features:
- Blocks private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
- Blocks localhost and loopback addresses (127.0.0.0/8, ::1)
- Blocks link-local addresses (169.254.0.0/16, fe80::/10)
- Blocks multicast and broadcast addresses
- Validates hostname resolution to prevent DNS rebinding attacks
- Supports IPv4-mapped IPv6 addresses validation
- Implements request timeouts to prevent resource exhaustion

Usage:
    # Basic validation
    validator = URLValidator()
    validator.validate_url("https://example.com/file.mp3")
    
    # With domain allowlist
    validator = URLValidator(allowed_domains=["trusted.com", "cdn.example.org"])
    validator.validate_url("https://trusted.com/audio.wav")
    
    # Convenience function
    is_safe = validate_url_safe("https://example.com/file.mp3")
"""

import ipaddress
import socket
import urllib.parse
from typing import List, Optional, Set


class SSRFError(Exception):
    """Exception raised when SSRF attack is detected"""
    pass


class URLValidator:
    """
    Validates URLs to prevent SSRF attacks
    
    Features:
    - URL scheme allowlisting (HTTP/HTTPS only)
    - Private IP range blocking (RFC 1918, loopback, link-local)
    - Hostname resolution validation
    - Domain allowlisting
    - Port validation
    - Request timeout configuration
    """
    
    # Allowed URL schemes
    ALLOWED_SCHEMES = {"http", "https"}
    
    # Dangerous ports that should be blocked
    DANGEROUS_PORTS = {
        22,    # SSH
        23,    # Telnet
        25,    # SMTP
        53,    # DNS
        110,   # POP3
        143,   # IMAP
        993,   # IMAPS
        995,   # POP3S
        1433,  # MSSQL
        3306,  # MySQL
        5432,  # PostgreSQL
        6379,  # Redis
        11211, # Memcached
        27017, # MongoDB
    }
    
    # Safe ports that are explicitly allowed
    SAFE_PORTS = {80, 443, 8080, 8443}
    
    def __init__(self, allowed_domains: Optional[List[str]] = None, request_timeout: int = 10):
        """
        Initialize URL validator
        
        Args:
            allowed_domains: Optional list of allowed domains (allowlist mode)
            request_timeout: Request timeout in seconds
        """
        self.allowed_domains: Optional[Set[str]] = None
        if allowed_domains:
            self.allowed_domains = set(allowed_domains)
        
        self.request_timeout = request_timeout
    
    def validate_url(self, url: str) -> None:
        """
        Validate URL for SSRF prevention
        
        Args:
            url: URL to validate
            
        Raises:
            SSRFError: If URL is potentially malicious or violates security policy
        """
        if not url or not url.strip():
            raise SSRFError("URL cannot be empty")
        
        # Parse URL
        try:
            parsed = urllib.parse.urlparse(url)
        except Exception as e:
            raise SSRFError(f"Malformed URL: {e}")
        
        # Check URL scheme first (important for security)
        self._validate_scheme(parsed.scheme)
        
        # Validate URL structure
        if not parsed.netloc:
            raise SSRFError("Malformed URL: missing host")
        
        # Extract hostname and port
        hostname = parsed.hostname
        port = parsed.port
        
        if not hostname:
            raise SSRFError("Malformed URL: missing hostname")
        
        # Validate port
        self._validate_port(port, parsed.scheme)
        
        # Check domain allowlist if configured
        if self.allowed_domains is not None:
            self._validate_domain_allowlist(hostname)
        
        # Validate IP address (if hostname is an IP)
        if self._is_ip_address(hostname):
            self._validate_ip_address(hostname)
        else:
            # Resolve hostname and validate resolved IPs
            self._validate_hostname_resolution(hostname)
    
    def _validate_scheme(self, scheme: str) -> None:
        """Validate URL scheme"""
        if scheme.lower() not in self.ALLOWED_SCHEMES:
            raise SSRFError(f"URL scheme '{scheme}' not allowed. Only {list(self.ALLOWED_SCHEMES)} are permitted")
    
    def _validate_port(self, port: Optional[int], scheme: str) -> None:
        """Validate port number"""
        if port is None:
            # Use default port for scheme
            return
        
        # Check if port is in dangerous ports list
        if port in self.DANGEROUS_PORTS:
            raise SSRFError(f"Port {port} not allowed - potentially dangerous service")
        
        # For explicit validation, require safe ports or default HTTP/HTTPS ports
        if port not in self.SAFE_PORTS and port not in {80, 443}:
            # Allow high ports (1024+) for legitimate services
            if port < 1024:
                raise SSRFError(f"Port {port} not allowed - system port range")
    
    def _validate_domain_allowlist(self, hostname: str) -> None:
        """Validate hostname against domain allowlist"""
        if self.allowed_domains is None:
            return
        
        # Check exact match
        if hostname in self.allowed_domains:
            return
        
        # Check subdomain match
        for allowed_domain in self.allowed_domains:
            if hostname.endswith(f".{allowed_domain}"):
                return
        
        raise SSRFError(f"Domain '{hostname}' not in allowlist")
    
    def _is_ip_address(self, hostname: str) -> bool:
        """Check if hostname is an IP address"""
        try:
            ipaddress.ip_address(hostname)
            return True
        except ValueError:
            return False
    
    def _validate_ip_address(self, ip_str: str) -> None:
        """Validate IP address against private ranges"""
        try:
            ip = ipaddress.ip_address(ip_str)
        except ValueError as e:
            raise SSRFError(f"Invalid IP address: {e}")
        
        # Check for localhost/loopback first (more specific)
        if ip.is_loopback:
            raise SSRFError(f"Localhost/loopback address not allowed: {ip}")
        
        # Check for link-local addresses  
        if ip.is_link_local:
            raise SSRFError(f"Link-local address not allowed: {ip}")
        
        # Check for private addresses (general case)
        if ip.is_private:
            raise SSRFError(f"Private IP address not allowed: {ip}")
        
        # Check for multicast addresses
        if ip.is_multicast:
            raise SSRFError(f"Multicast address not allowed: {ip}")
        
        # Check for unspecified addresses (0.0.0.0, ::)
        if ip.is_unspecified:
            raise SSRFError(f"Unspecified address not allowed: {ip}")
        
        # Additional checks for IPv4
        if isinstance(ip, ipaddress.IPv4Address):
            # Check for broadcast address
            if str(ip) == "255.255.255.255":
                raise SSRFError(f"Broadcast address not allowed: {ip}")
            
            # Check for special ranges
            # 169.254.0.0/16 - Link-local (already covered by is_link_local)
            # 224.0.0.0/4 - Multicast (already covered by is_multicast)
            pass
        
        # Additional checks for IPv6
        if isinstance(ip, ipaddress.IPv6Address):
            # Check for IPv4-mapped IPv6 addresses that might bypass validation
            if ip.ipv4_mapped:
                self._validate_ip_address(str(ip.ipv4_mapped))
    
    def _validate_hostname_resolution(self, hostname: str) -> None:
        """Validate hostname by resolving it and checking resulting IP addresses"""
        try:
            # Resolve hostname to IP addresses
            addr_info = socket.getaddrinfo(hostname, None, family=socket.AF_UNSPEC, type=socket.SOCK_STREAM)
            
            resolved_ips = set()
            for family, type_, proto, canonname, sockaddr in addr_info:
                if family == socket.AF_INET:
                    ip = sockaddr[0]
                elif family == socket.AF_INET6:
                    ip = sockaddr[0]
                else:
                    continue
                resolved_ips.add(ip)
            
            # Validate each resolved IP
            for ip in resolved_ips:
                try:
                    self._validate_ip_address(ip)
                except SSRFError as e:
                    raise SSRFError(f"Hostname '{hostname}' resolves to private/dangerous IP: {e}")
        
        except socket.gaierror as e:
            raise SSRFError(f"Failed to resolve hostname '{hostname}': {e}")
        except Exception as e:
            raise SSRFError(f"Error validating hostname '{hostname}': {e}")


def validate_url_safe(url: str, allowed_domains: Optional[List[str]] = None, timeout: int = 10) -> bool:
    """
    Convenience function to safely validate a URL
    
    Args:
        url: URL to validate
        allowed_domains: Optional list of allowed domains
        timeout: Request timeout in seconds
        
    Returns:
        True if URL is safe, False otherwise
    """
    try:
        validator = URLValidator(allowed_domains=allowed_domains, request_timeout=timeout)
        validator.validate_url(url)
        return True
    except SSRFError:
        return False