package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/sibellavia/dory/internal/plugin"
)

const defaultHookTimeout = 2 * time.Second

func runPluginHooks(event plugin.HookEvent, context map[string]interface{}) {
	results, err := plugin.RunHooks(doryRoot, event, context, defaultHookTimeout)

	for _, result := range results {
		switch result.Status {
		case "warning":
			if result.Error != "" {
				fmt.Fprintf(os.Stderr, "Warning: hook %s (%s): %s\n", result.Plugin, result.Event, result.Error)
			}
		case "error":
			if result.Error != "" {
				fmt.Fprintf(os.Stderr, "Warning: hook %s (%s) failed: %s\n", result.Plugin, result.Event, result.Error)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: hook %s (%s) failed\n", result.Plugin, result.Event)
			}
		}
	}

	CheckError(err)
}
