#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"fixture healthy"}}'
    ;;
  *'"method":"dory.command.run"'*)
    printf '%s\n' '{"id":"req-1","result":{"output":"fixture command output\n","message":"fixture done"}}'
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"unknown method"}}'
    ;;
esac
