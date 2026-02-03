#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"tui-insights ready"}}'
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"method not supported"}}'
    ;;
esac
