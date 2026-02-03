# Dory Plugin Interface

This page explains how Dory talks to plugins.

## How it works

- A plugin is an executable declared in `plugin.yaml`.
- Dory starts the plugin process when needed.
- Dory sends one JSON request line on stdin.
- The plugin returns one JSON response line on stdout.
- Stderr is captured and shown as warnings/diagnostics.

## Request and response format

Request:

```json
{
  "id": "req-1",
  "method": "dory.health",
  "params": {
    "api_version": "v1"
  }
}
```

Response (success):

```json
{
  "id": "req-1",
  "result": {
    "status": "ok",
    "message": "healthy"
  }
}
```

Response (error):

```json
{
  "id": "req-1",
  "error": {
    "code": 400,
    "message": "bad request"
  }
}
```

## Supported methods

### `dory.health`
Used by `dory plugin doctor`.

Input:
- `api_version`

Typical result fields:
- `status`
- `message`

### `dory.command.run`
Used by `dory plugin run`.

Input:
- `api_version`
- `plugin`
- `command`
- `args`
- `cwd`

Typical result fields:
- `output`
- `message`

### `dory.hook.run`
Used for lifecycle hooks.

Input:
- `api_version`
- `event` (`before_create`, `after_create`, `before_remove`, `after_remove`, `after_compact`)
- `context` (event payload)

Typical result fields:
- `allow` (`false` can block `before_*` operations)
- `message`

### `dory.type.validate`
Used before `dory type create` persists a custom type item.

Input:
- `api_version`
- `type`
- `oneliner`
- `topic`
- `body`
- `refs`

Required result field:
- `valid` (boolean)

Optional result fields:
- `message`
- `errors` (array of strings)

## Compatibility notes

- `api_version` must be `v1`.
- Unknown fields should be ignored when possible.
- Plugin storage backends are not supported; `.dory` stays the only backend.
