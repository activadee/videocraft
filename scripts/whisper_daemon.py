#!/usr/bin/env python3
"""
Whisper Transcription Daemon

A persistent Python process that handles audio transcription requests
via stdin/stdout communication with idle timeout for resource optimization.
"""
import contextlib
import io
import json
import logging
import os
import secrets
import ssl
import sys
import tempfile
import threading
import time
import traceback
import urllib.request
import warnings
from typing import Dict, Any, Optional
from url_validator import URLValidator, SSRFError

# Suppress FP16 warnings that interfere with JSON communication
warnings.filterwarnings("ignore", message="FP16 is not supported on CPU.*")

try:
    import whisper
    import torch

    WHISPER_AVAILABLE = True
except ImportError:
    WHISPER_AVAILABLE = False
    whisper = None
    torch = None

# Configure logging - only errors to stderr, clean JSON communication
logging.basicConfig(
    level=logging.ERROR,  # Only log errors to stderr
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stderr)],
)
logger = logging.getLogger(__name__)

# Create a separate logger for internal debugging that doesn't interfere with JSON
debug_logger = logging.getLogger("whisper_debug")
debug_logger.setLevel(logging.INFO)
# Don't add handlers - this logger will be silent during normal operation


class WhisperDaemon:
    """Whisper transcription daemon with idle timeout"""

    def __init__(
        self,
        idle_timeout: int = 300,
        model_name: str = "base",
        allowed_domains: Optional[list] = None,
    ):
        """
        Initialize Whisper daemon

        Args:
            idle_timeout: Seconds to wait before auto-shutdown (default 5 minutes)
            model_name: Whisper model to use (tiny/base/small/medium/large)
            allowed_domains: Optional list of allowed domains for URL allowlisting
        """
        self.idle_timeout = idle_timeout
        self.model_name = model_name
        self.model = None
        self.device = self._get_optimal_device()
        self.last_activity = time.time()
        self.running = True
        self.shutdown_event = threading.Event()

        # Initialize URL validator for SSRF protection
        self.url_validator = URLValidator(
            allowed_domains=allowed_domains, request_timeout=30
        )

        # Start idle checker thread
        self.idle_thread = threading.Thread(target=self._idle_checker, daemon=True)
        self.idle_thread.start()

        # Pre-load model during initialization to be ready for requests
        self._load_model()

    def _get_optimal_device(self) -> str:
        """Determine the best available device for Whisper"""
        if not torch:
            return "cpu"

        # Check available devices in order of preference
        # Note: MPS has sparse tensor issues with Whisper, force CPU for now
        if torch.cuda.is_available():
            return "cuda"
        else:
            return "cpu"

    def _load_model(self) -> None:
        """Load Whisper model if not already loaded"""
        if self.model is None:
            # Silent loading - no logging to stderr
            try:
                self.model = whisper.load_model(self.model_name, device=self.device)
            except Exception as e:
                logger.error(f"Failed to load model {self.model_name}: {e}")
                raise

    def _unload_model(self) -> None:
        """Unload model to free memory"""
        if self.model is not None:
            del self.model
            self.model = None

            # Force garbage collection
            if torch and torch.cuda.is_available():
                torch.cuda.empty_cache()

    def _idle_checker(self) -> None:
        """Background thread that checks for idle timeout"""
        while self.running:
            time.sleep(10)  # Check every 10 seconds

            if time.time() - self.last_activity > self.idle_timeout:
                self._shutdown()
                break

    def _shutdown(self) -> None:
        """Graceful shutdown"""
        self.running = False
        self._unload_model()
        self.shutdown_event.set()

    def _create_secure_temp_file(self) -> str:
        """
        Create a secure temporary file with cryptographically random name.
        Simple implementation appropriate for 10-user scale.

        Returns:
            Path to secure temporary file

        Raises:
            OSError: If file creation fails
        """
        # Generate cryptographically secure random filename
        random_name = secrets.token_hex(16)  # 32 characters, 128 bits entropy
        temp_dir = tempfile.gettempdir()

        # Verify temp directory is writable
        if not os.access(temp_dir, os.W_OK):
            raise OSError(f"Temp directory not writable: {temp_dir}")

        temp_path = os.path.join(temp_dir, f"whisper_{random_name}.wav")

        try:
            # Create file with secure permissions (owner only)
            fd = os.open(temp_path, os.O_CREAT | os.O_WRONLY | os.O_EXCL, 0o600)
            os.close(fd)  # Close immediately, we'll reopen for writing

            logger.debug(f"Created secure temp file: {os.path.basename(temp_path)}")
            return temp_path
        except OSError as e:
            logger.error(f"Failed to create secure temp file: {e}")
            raise

    def transcribe_audio(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """
        Transcribe audio from URL

        Args:
            request: Request dictionary with url, language, etc.

        Returns:
            Response dictionary with transcription results
        """
        try:
            # Update activity timestamp
            self.last_activity = time.time()

            # Load model if needed
            self._load_model()

            # Extract parameters
            audio_url = request.get("url")
            language = request.get("language", "auto")
            word_timestamps = request.get("word_timestamps", True)

            if not audio_url:
                raise ValueError("Missing 'url' parameter")

            # Validate URL for SSRF protection
            try:
                self.url_validator.validate_url(audio_url)
                debug_logger.info(f"URL validation passed for: {audio_url}")
            except SSRFError as e:
                # Log security violation
                logger.error(
                    f"SECURITY: SSRF attempt blocked for URL: {audio_url} - {e}"
                )
                raise ValueError(f"URL validation failed: {e}")

            # Download and transcribe audio

            # Create SSL context with proper certificate verification
            ssl_context = ssl.create_default_context()
            # Keep default security settings:
            # ssl_context.check_hostname = True (default)
            # ssl_context.verify_mode = ssl.CERT_REQUIRED (default)
            #
            # Note: Certificate pinning is not implemented as the daemon
            # handles arbitrary audio URLs from various sources, making
            # pinning impractical. Standard certificate validation provides
            # adequate security for this use case.

            # Download audio to secure temporary file
            temp_path = self._create_secure_temp_file()

            try:
                # Use urllib with SSL context and timeout
                req = urllib.request.Request(
                    audio_url, headers={"User-Agent": "Mozilla/5.0"}
                )
                with urllib.request.urlopen(
                    req, context=ssl_context, timeout=self.url_validator.request_timeout
                ) as response:
                    with open(temp_path, "wb") as temp_file:
                        temp_file.write(response.read())

                # Log security event
                logger.info(
                    f"Processing audio file: {audio_url} -> "
                    f"{os.path.basename(temp_path)}"
                )

                # Perform transcription on local file with output redirection

                # Capture stdout to prevent "Detected language" from
                # interfering with JSON
                captured_output = io.StringIO()
                with contextlib.redirect_stdout(captured_output):
                    with contextlib.redirect_stderr(captured_output):
                        result = self.model.transcribe(
                            temp_path,
                            language=None if language == "auto" else language,
                            word_timestamps=word_timestamps,
                            verbose=False,
                            temperature=0,
                            best_of=1,
                            beam_size=1,
                        )

            finally:
                # Guaranteed cleanup with error logging
                try:
                    if os.path.exists(temp_path):
                        os.unlink(temp_path)
                        logger.debug(
                            f"Cleaned up temp file: {os.path.basename(temp_path)}"
                        )
                except Exception as cleanup_error:
                    logger.error(
                        f"SECURITY: Failed to cleanup temp file {temp_path}: "
                        f"{cleanup_error}"
                    )

            # Extract word timestamps for progressive subtitles
            word_timestamps_list = []
            if "segments" in result:
                for segment in result["segments"]:
                    if "words" in segment:
                        word_timestamps_list.extend(segment["words"])

            response = {
                "success": True,
                "text": result["text"].strip(),
                "language": result.get("language", "unknown"),
                "duration": sum(
                    segment.get("end", 0) for segment in result.get("segments", [])
                ),
                "segments": result.get("segments", []),
                "word_timestamps": word_timestamps_list,
            }

            return response

        except Exception as e:
            logger.error(f"Transcription failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "traceback": traceback.format_exc(),
            }

    def handle_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """
        Handle incoming request

        Args:
            request: Request dictionary

        Returns:
            Response dictionary
        """
        request_id = request.get("id", "unknown")
        action = request.get("action")

        try:
            if action == "transcribe":
                response = self.transcribe_audio(request)
            elif action == "ping":
                response = {"success": True, "message": "pong"}
            elif action == "status":
                response = {
                    "success": True,
                    "model": self.model_name,
                    "device": self.device,
                    "model_loaded": self.model is not None,
                    "last_activity": self.last_activity,
                    "idle_timeout": self.idle_timeout,
                }
            elif action == "shutdown":
                response = {"success": True, "message": "shutting down"}
                self._shutdown()
            else:
                response = {"success": False, "error": f"Unknown action: {action}"}

            # Add request ID to response
            response["id"] = request_id
            return response

        except Exception as e:
            logger.error(f"Request handling failed: {e}")
            return {
                "id": request_id,
                "success": False,
                "error": str(e),
                "traceback": traceback.format_exc(),
            }

    def run(self) -> None:
        """Main daemon loop - read requests from stdin, write responses to stdout"""
        # Redirect stderr to devnull to prevent interference with JSON communication
        stderr_fd = os.open(os.devnull, os.O_WRONLY)
        os.dup2(stderr_fd, sys.stderr.fileno())
        os.close(stderr_fd)

        try:
            while self.running:
                # Read request from stdin
                try:
                    line = sys.stdin.readline()
                    if not line:  # EOF
                        break

                    line = line.strip()
                    if not line:
                        continue

                    # Parse JSON request
                    request = json.loads(line)

                    # Handle request
                    response = self.handle_request(request)

                    # Send JSON response to stdout
                    print(json.dumps(response), flush=True)

                except json.JSONDecodeError as e:
                    error_response = {"success": False, "error": f"Invalid JSON: {e}"}
                    print(json.dumps(error_response), flush=True)

                except KeyboardInterrupt:
                    break

                except Exception as e:
                    error_response = {
                        "success": False,
                        "error": str(e),
                        "traceback": traceback.format_exc(),
                    }
                    print(json.dumps(error_response), flush=True)

        finally:
            self._shutdown()


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Whisper Transcription Daemon")
    parser.add_argument(
        "--idle-timeout",
        type=int,
        default=300,
        help="Idle timeout in seconds (default: 300)",
    )
    parser.add_argument(
        "--model",
        type=str,
        default="base",
        choices=["tiny", "base", "small", "medium", "large-v1", "large-v2", "large-v3"],
        help="Whisper model to use (default: base)",
    )
    parser.add_argument(
        "--log-level",
        type=str,
        default="INFO",
        choices=["DEBUG", "INFO", "WARNING", "ERROR"],
        help="Log level (default: INFO)",
    )

    args = parser.parse_args()

    # Set log level
    logging.getLogger().setLevel(getattr(logging, args.log_level))

    # Check if Whisper is available
    if not WHISPER_AVAILABLE:
        logger.error(
            "Whisper not available! Please install: pip install openai-whisper torch"
        )
        sys.exit(1)

    # Create and run daemon
    daemon = WhisperDaemon(idle_timeout=args.idle_timeout, model_name=args.model)

    try:
        daemon.run()
    except KeyboardInterrupt:
        logger.info("Daemon stopped by user")
    except Exception as e:
        logger.error(f"Daemon error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
