#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"hook-guard ready"}}'
    ;;
  *'"method":"dory.hook.run"'*)
    case "$line" in
      *'"event":"before_remove"'*)
        case "$line" in
          *'"id":"D'*)
            printf '%s\n' '{"id":"req-1","result":{"allow":false,"message":"Decision entries are protected by hook-guard"}}'
            ;;
          *)
            printf '%s\n' '{"id":"req-1","result":{"allow":true,"message":"remove allowed"}}'
            ;;
        esac
        ;;
      *'"event":"after_create"'*)
        printf '%s\n' '{"id":"req-1","result":{"allow":true,"message":"create observed by hook-guard"}}'
        ;;
      *)
        printf '%s\n' '{"id":"req-1","result":{"allow":true}}'
        ;;
    esac
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"method not supported"}}'
    ;;
esac
