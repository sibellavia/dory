#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"command-echo ready"}}'
    ;;
  *'"method":"dory.command.run"'*)
    case "$line" in
      *'"command":"echo"'*)
        printf '%s\n' '{"id":"req-1","result":{"output":"echo from plugin\n"}}'
        ;;
      *'"command":"time"'*)
        now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        printf '{"id":"req-1","result":{"output":"utc time: %s"}}\n' "$now"
        ;;
      *)
        printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"unknown command"}}'
        ;;
    esac
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"method not supported"}}'
    ;;
esac
