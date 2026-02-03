#!/bin/sh
set -eu

if ! IFS= read -r line; then
  exit 1
fi

case "$line" in
  *'"method":"dory.health"'*)
    printf '%s\n' '{"id":"req-1","result":{"status":"ok","message":"postmortem-validator ready"}}'
    ;;
  *'"method":"dory.type.validate"'*)
    case "$line" in
      *'"type":"postmortem"'*)
        ;;
      *)
        printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"unsupported type","errors":["type must be postmortem"]}}'
        exit 0
        ;;
    esac

    if printf '%s' "$line" | grep -q '"topic":""'; then
      printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["topic is required"]}}'
      exit 0
    fi

    if printf '%s' "$line" | grep -q '"body":""'; then
      printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["body is required"]}}'
      exit 0
    fi

    has_timeline=0
    has_actions=0
    case "$line" in
      *'## Timeline'*) has_timeline=1 ;;
    esac
    case "$line" in
      *'## Action Items'*) has_actions=1 ;;
    esac

    if [ "$has_timeline" -eq 1 ] && [ "$has_actions" -eq 1 ]; then
      printf '%s\n' '{"id":"req-1","result":{"valid":true,"message":"schema valid"}}'
    else
      printf '%s\n' '{"id":"req-1","result":{"valid":false,"message":"invalid payload","errors":["body must include headings: ## Timeline and ## Action Items"]}}'
    fi
    ;;
  *)
    printf '%s\n' '{"id":"req-1","error":{"code":404,"message":"method not supported"}}'
    ;;
esac
