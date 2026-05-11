#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: bash db/seed.sh <password>" >&2
  exit 1
fi

HASH=$(python3 -c "import bcrypt, sys; print(bcrypt.hashpw(sys.argv[1].encode(), bcrypt.gensalt(12)).decode())" "$1")
SEED_FILE="$(dirname "$0")/migrations/000003_seed.up.sql"

sed -i "s|'PLACEHOLDER_HASH'|'${HASH}'|" "$SEED_FILE"
echo "Seed hash written to $SEED_FILE"
