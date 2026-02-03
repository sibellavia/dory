# incident-validator plugin

Example plugin that provides a custom type:

- type: `incident`
- validation method: `dory.type.validate`

Validation rules:

1. `topic` must be present
2. `body` must be present
3. `body` must include `## Impact`
