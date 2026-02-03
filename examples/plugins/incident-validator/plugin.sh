#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"incident-validator ready"}}'
    ;;
  *'"method":"dory.type.validate"'*)
    case "$line" in
      *'"type":"incident"'*)
        ;;
      *)
        printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"unsupported type","errors":["type must be incident"]}}'
        exit 0
        ;;
    esac

    case "$line" in
      *'"topic":""'*)
        printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["topic is required"]}}'
        exit 0
        ;;
    esac

    case "$line" in
      *'"body":""'*)
        printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["body is required"]}}'
        exit 0
        ;;
    esac

    case "$line" in
      *'## Impact'*)
        printf '%s\n' '{"id":"req-1","result":{"valid":true,"message":"schema valid"}}'
        ;;
      *)
        printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["body must include section heading: ## Impact"]}}'
        ;;
    esac
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"method not supported"}}'
    ;;
esac
