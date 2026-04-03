#!/usr/bin/env bash
# validators/check_stack.sh
# Reads tool input from stdin and blocks forbidden substitutions.
# Exit 0 = allow. Exit 2 = block (message sent to agent).

set -euo pipefail

input=$(cat)

check_forbidden() {
  local pattern="$1"
  local message="$2"
  if echo "$input" | grep -qiE "$pattern"; then
    echo "BLOCKED by stack_config.json: $message"
    exit 2
  fi
}

# Package manager
check_forbidden "pip install" "Use 'uv add' instead of 'pip install'"

# Password hashing
check_forbidden "import bcrypt|from bcrypt" "Use 'argon2-cffi' (argon2id). bcrypt is FORBIDDEN."
check_forbidden "hashlib\.md5|hashlib\.sha1" "Use argon2id for passwords. md5/sha1 are FORBIDDEN."

# HTTP client
check_forbidden "import requests|from requests" "Use 'httpx' instead of 'requests'"

# Hard deletes
check_forbidden "\.delete\(\)|DELETE FROM|session\.delete\(" "Use soft delete: set deleted_at. Hard deletes are FORBIDDEN."

exit 0
