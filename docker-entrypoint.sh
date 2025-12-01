#!/bin/sh
set -eu

# Ensure upload and log directories exist (bind mounts may create them with host ownership)
mkdir -p /app/uploads/avatars /app/uploads/videos /app/uploads/files /app/logs

# Allow override to skip chown from compose/env
if [ "${ENTRYPOINT_SKIP_CHOWN:-}" = "true" ]; then
  echo "[entrypoint] ENTRYPOINT_SKIP_CHOWN=true - skipping chown" 1>&2
else
  can_chown=true

  # Detect rootless mapping: /proc/1/uid_map shows mapping of container UIDs to host UIDs
  if [ -r /proc/1/uid_map ]; then
    host_uid=$(awk 'NR==1 {print $2}' /proc/1/uid_map 2>/dev/null || echo "")
    if [ -n "$host_uid" ] && [ "$host_uid" != "0" ]; then
      echo "[entrypoint] rootless uid mapping detected (host uid: $host_uid) - skipping chown" 1>&2
      can_chown=false
    fi
  fi

  if [ "$can_chown" = true ]; then
    # Only perform chown if uploads is owned by root (0) or doesn't exist yet; this avoids reassigning
    # ownership for host bind-mounts that were created by another UID.
    if [ -e /app/uploads ]; then
      owner_uid=$(stat -c '%u' /app/uploads 2>/dev/null || echo "")
      if [ -z "$owner_uid" ] || [ "$owner_uid" = "0" ]; then
        echo "[entrypoint] setting ownership to 1000:1000 for /app/uploads and /app/logs" 1>&2
        chown -R 1000:1000 /app/uploads /app/logs || true
      else
        echo "[entrypoint] /app/uploads owned by uid $owner_uid - skipping chown to avoid changing host ownership" 1>&2
      fi
    else
      chown -R 1000:1000 /app/uploads /app/logs || true
    fi
  fi
fi

# Execute the binary as appuser if possible (use su-exec), otherwise run the binary directly
run_binary() {
  if command -v su-exec >/dev/null 2>&1; then
    exec su-exec 1000:1000 "$@"
  else
    exec "$@"
  fi
}

if [ $# -eq 0 ]; then
  run_binary /app/api
else
  cmd="$1"
  shift
  if [ "${cmd##*/}" = "api" ] || [ "${cmd}" = "/app/api" ]; then
    run_binary /app/api "$@"
  else
    exec "$cmd" "$@"
  fi
fi
