# Plugin Examples

This folder contains runnable reference plugins you can install locally with:

```bash
dory plugin install ./examples/plugins/<plugin-folder> --enable
```

Available examples:

- `incident-validator`: custom type `incident` with schema validation (`dory.type.validate`)
- `postmortem-validator`: custom type `postmortem` with richer section checks
- `command-echo`: command plugin example (`dory.command.run`)
- `hook-guard`: hook plugin example (`dory.hook.run`) that can block removals
- `tui-insights`: declares TUI extension points for discovery (`dory plugin tui`)
