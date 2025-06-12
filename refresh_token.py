#!/usr/bin/env python3

import subprocess
import json
import sys


def fetch_from_keychain(service, account):
    """Fetch password from macOS keychain"""
    cmd = ["security", "find-generic-password", "-s", service, "-a", account, "-w"]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Failed to fetch from keychain: {result.stderr}")
    return result.stdout.strip()


def set_github_secret(repo, key, value, environment=None):
    """Set a GitHub secret using gh CLI"""
    cmd = ["gh", "secret", "set", key, "--repo", repo, "--body", str(value)]
    if environment:
        cmd.extend(["--env", environment])

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Failed to set secret {key}: {result.stderr}")


def main():
    # Configuration
    REPO = "activadee/videocraft"
    SERVICE_NAME = "Claude Code-credentials"
    ACCOUNT_NAME = "patryk"
    ENVIRONMENT = None  # Set to environment name if needed, e.g. "staging"

    # Fetch JSON from keychain
    json_string = fetch_from_keychain(SERVICE_NAME, ACCOUNT_NAME)
    secrets = json.loads(json_string)

    CLAUDE_ACCESS_TOKEN = secrets["claudeAiOauth"]["accessToken"]
    CLAUDE_REFRESH_TOKEN = secrets["claudeAiOauth"]["refreshToken"]
    CLAUDE_EXPIRES_AT = secrets["claudeAiOauth"]["expiresAt"]

    set_github_secret(REPO, "CLAUDE_ACCESS_TOKEN", CLAUDE_ACCESS_TOKEN, ENVIRONMENT)
    set_github_secret(REPO, "CLAUDE_REFRESH_TOKEN", CLAUDE_REFRESH_TOKEN, ENVIRONMENT)
    set_github_secret(REPO, "CLAUDE_EXPIRES_AT", CLAUDE_EXPIRES_AT, ENVIRONMENT)


    print(f"\n✅ All secrets updated successfully!")


if __name__ == "__main__":
    main()