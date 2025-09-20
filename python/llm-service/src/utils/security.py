"""
Security utilities for Multi-Agent LLM Service
"""

import re
import logging
from typing import List, Dict, Any
import hashlib
import secrets

from src.core.models import SecurityValidation

logger = logging.getLogger(__name__)

class SecurityValidator:
    """Security validation for user inputs and content"""
    
    def __init__(self):
        # Dangerous patterns to detect
        self.dangerous_patterns = [
            # Code injection patterns
            r'(?i)(exec|eval|import\s+os|subprocess|system|shell)',
            r'(?i)(__import__|getattr|setattr|delattr)',
            r'(?i)(rm\s+-rf|del\s+/|format\s+c:)',
            
            # SQL injection patterns
            r'(?i)(union\s+select|drop\s+table|delete\s+from)',
            r'(?i)(insert\s+into|update\s+set|alter\s+table)',
            
            # XSS patterns
            r'(?i)(<script|javascript:|on\w+\s*=)',
            r'(?i)(document\.cookie|window\.location)',
            
            # Path traversal
            r'\.\.\/|\.\.\\',
            
            # Command injection
            r'[;&|`$(){}[\]]',
            
            # Sensitive information patterns
            r'(?i)(password|secret|token|key)\s*[:=]\s*["\']?\w+',
            r'(?i)(api[_-]?key|access[_-]?token)\s*[:=]',
            
            # Prompt injection attempts
            r'(?i)(ignore\s+previous|forget\s+instructions)',
            r'(?i)(system\s+prompt|you\s+are\s+now)',
            r'(?i)(jailbreak|bypass\s+safety)',
        ]
        
        # Compile patterns for efficiency
        self.compiled_patterns = [re.compile(pattern) for pattern in self.dangerous_patterns]
        
        # Allowed file extensions
        self.allowed_extensions = {'.txt', '.md', '.json', '.yaml', '.yml', '.csv'}
        
        # Maximum content length
        self.max_content_length = 100000  # 100KB
    
    async def validate_content(self, content: str) -> SecurityValidation:
        """Validate content for security issues"""
        
        violations = []
        recommendations = []
        risk_score = 0.0
        
        # Check content length
        if len(content) > self.max_content_length:
            violations.append("Content exceeds maximum length limit")
            risk_score += 0.3
        
        # Check for dangerous patterns
        for i, pattern in enumerate(self.compiled_patterns):
            matches = pattern.findall(content)
            if matches:
                pattern_desc = self._get_pattern_description(i)
                violations.append(f"Detected {pattern_desc}: {matches[:3]}")  # Show first 3 matches
                risk_score += 0.2
        
        # Check for excessive special characters
        special_char_ratio = sum(1 for c in content if not c.isalnum() and not c.isspace()) / len(content) if content else 0
        if special_char_ratio > 0.3:
            violations.append("High ratio of special characters detected")
            risk_score += 0.1
        
        # Check for base64 encoded content (potential data exfiltration)
        base64_pattern = re.compile(r'[A-Za-z0-9+/]{20,}={0,2}')
        if base64_pattern.search(content):
            violations.append("Potential base64 encoded content detected")
            risk_score += 0.15
        
        # Check for repeated patterns (potential spam/DoS)
        if self._has_repeated_patterns(content):
            violations.append("Repeated patterns detected (potential spam)")
            risk_score += 0.1
        
        # Generate recommendations
        if violations:
            recommendations.extend([
                "Review and sanitize input content",
                "Consider using content filtering",
                "Implement rate limiting for this user"
            ])
        
        # Normalize risk score
        risk_score = min(1.0, risk_score)
        
        is_safe = len(violations) == 0 and risk_score < 0.5
        
        return SecurityValidation(
            is_safe=is_safe,
            risk_score=risk_score,
            violations=violations,
            recommendations=recommendations
        )
    
    def _get_pattern_description(self, pattern_index: int) -> str:
        """Get description for pattern index"""
        descriptions = [
            "code injection attempt",
            "attribute manipulation",
            "destructive command",
            "SQL injection",
            "SQL manipulation",
            "XSS attempt",
            "DOM manipulation",
            "path traversal",
            "command injection",
            "credential exposure",
            "API key exposure",
            "prompt injection"
        ]
        
        if pattern_index < len(descriptions):
            return descriptions[pattern_index]
        return "suspicious pattern"
    
    def _has_repeated_patterns(self, content: str) -> bool:
        """Check for repeated patterns that might indicate spam"""
        
        # Check for repeated words
        words = content.split()
        if len(words) > 10:
            word_counts = {}
            for word in words:
                word_counts[word] = word_counts.get(word, 0) + 1
            
            # If any word appears more than 20% of the time, it's suspicious
            max_count = max(word_counts.values()) if word_counts else 0
            if max_count > len(words) * 0.2:
                return True
        
        # Check for repeated character sequences
        for length in [3, 4, 5]:
            sequences = {}
            for i in range(len(content) - length + 1):
                seq = content[i:i+length]
                sequences[seq] = sequences.get(seq, 0) + 1
            
            if sequences:
                max_seq_count = max(sequences.values())
                if max_seq_count > 10:  # Same sequence repeated more than 10 times
                    return True
        
        return False
    
    def validate_file_path(self, file_path: str) -> SecurityValidation:
        """Validate file path for security"""
        
        violations = []
        risk_score = 0.0
        
        # Check for path traversal
        if '..' in file_path:
            violations.append("Path traversal attempt detected")
            risk_score += 0.8
        
        # Check for absolute paths
        if file_path.startswith('/') or (len(file_path) > 1 and file_path[1] == ':'):
            violations.append("Absolute path not allowed")
            risk_score += 0.6
        
        # Check file extension
        import os
        _, ext = os.path.splitext(file_path.lower())
        if ext and ext not in self.allowed_extensions:
            violations.append(f"File extension '{ext}' not allowed")
            risk_score += 0.4
        
        # Check for suspicious filenames
        suspicious_names = ['passwd', 'shadow', 'hosts', '.env', 'config']
        filename = os.path.basename(file_path).lower()
        if any(name in filename for name in suspicious_names):
            violations.append("Suspicious filename detected")
            risk_score += 0.5
        
        is_safe = len(violations) == 0
        
        return SecurityValidation(
            is_safe=is_safe,
            risk_score=min(1.0, risk_score),
            violations=violations,
            recommendations=["Use relative paths only", "Stick to allowed file extensions"] if violations else []
        )
    
    def generate_secure_token(self, length: int = 32) -> str:
        """Generate a secure random token"""
        return secrets.token_urlsafe(length)
    
    def hash_content(self, content: str) -> str:
        """Generate hash of content for caching/deduplication"""
        return hashlib.sha256(content.encode('utf-8')).hexdigest()
    
    def sanitize_filename(self, filename: str) -> str:
        """Sanitize filename for safe storage"""
        # Remove dangerous characters
        sanitized = re.sub(r'[<>:"/\\|?*]', '_', filename)
        
        # Remove leading/trailing dots and spaces
        sanitized = sanitized.strip('. ')
        
        # Limit length
        if len(sanitized) > 255:
            name, ext = os.path.splitext(sanitized)
            sanitized = name[:255-len(ext)] + ext
        
        return sanitized or 'unnamed_file'
    
    def validate_url(self, url: str) -> SecurityValidation:
        """Validate URL for security"""
        
        violations = []
        risk_score = 0.0
        
        # Check for dangerous protocols
        dangerous_protocols = ['file://', 'ftp://', 'javascript:', 'data:']
        for protocol in dangerous_protocols:
            if url.lower().startswith(protocol):
                violations.append(f"Dangerous protocol detected: {protocol}")
                risk_score += 0.8
        
        # Check for localhost/internal IPs
        internal_patterns = [
            r'localhost',
            r'127\.0\.0\.1',
            r'192\.168\.',
            r'10\.',
            r'172\.(1[6-9]|2[0-9]|3[01])\.'
        ]
        
        for pattern in internal_patterns:
            if re.search(pattern, url.lower()):
                violations.append("Internal/localhost URL not allowed")
                risk_score += 0.6
                break
        
        # Check URL length
        if len(url) > 2048:
            violations.append("URL too long")
            risk_score += 0.2
        
        is_safe = len(violations) == 0
        
        return SecurityValidation(
            is_safe=is_safe,
            risk_score=min(1.0, risk_score),
            violations=violations,
            recommendations=["Use HTTPS URLs only", "Avoid internal/localhost URLs"] if violations else []
        )

class RateLimiter:
    """Simple in-memory rate limiter"""
    
    def __init__(self):
        self.requests = {}  # {key: [timestamps]}
        self.cleanup_interval = 3600  # 1 hour
        self.last_cleanup = 0
    
    def is_allowed(self, key: str, limit: int, window: int) -> bool:
        """Check if request is allowed under rate limit"""
        
        current_time = time.time()
        
        # Cleanup old entries periodically
        if current_time - self.last_cleanup > self.cleanup_interval:
            self._cleanup_old_entries(current_time)
            self.last_cleanup = current_time
        
        # Get or create request history for key
        if key not in self.requests:
            self.requests[key] = []
        
        request_times = self.requests[key]
        
        # Remove requests outside the window
        cutoff_time = current_time - window
        request_times[:] = [t for t in request_times if t > cutoff_time]
        
        # Check if under limit
        if len(request_times) < limit:
            request_times.append(current_time)
            return True
        
        return False
    
    def _cleanup_old_entries(self, current_time: float):
        """Clean up old rate limit entries"""
        
        cutoff_time = current_time - 3600  # Keep 1 hour of history
        
        for key in list(self.requests.keys()):
            request_times = self.requests[key]
            request_times[:] = [t for t in request_times if t > cutoff_time]
            
            # Remove empty entries
            if not request_times:
                del self.requests[key]

# Global instances
security_validator = SecurityValidator()
rate_limiter = RateLimiter()