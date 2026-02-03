package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sibellavia/dory/internal/plugin"
	"github.com/sibellavia/dory/internal/store"
	"github.com/spf13/cobra"
)

var typeCreateCmd = &cobra.Command{
	Use:   "create <type> <oneliner>",
	Short: "Create a knowledge item using a plugin-provided custom type",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		RequireStore()

		itemType := args[0]
		oneliner := strings.Join(args[1:], " ")
		topic, _ := cmd.Flags().GetString("topic")
		bodyFlag, _ := cmd.Flags().GetString("body")
		refs, _ := cmd.Flags().GetStringSlice("refs")
		validateTimeout, _ := cmd.Flags().GetDuration("validate-timeout")

		CheckError(validateItemType(itemType))
		if isCoreItemType(itemType) {
			CheckError(fmt.Errorf("type %q is a core type; use dory %s instead", itemType, itemType))
		}

		plugins, _, err := plugin.Discover(doryRoot)
		CheckError(err)

		var providers []plugin.PluginInfo
		for _, p := range plugins {
			if !p.Enabled {
				continue
			}
			for _, t := range p.Capabilities.Types {
				if t == itemType {
					providers = append(providers, p)
					break
				}
			}
		}

		if len(providers) == 0 {
			CheckError(fmt.Errorf("no enabled plugin provides custom type %q", itemType))
		}
		if len(providers) > 1 {
			var names []string
			for _, provider := range providers {
				names = append(names, provider.Name)
			}
			CheckError(fmt.Errorf("custom type %q is provided by multiple enabled plugins: %s", itemType, strings.Join(names, ", ")))
		}

		var body string
		if bodyFlag == "-" {
			content, err := io.ReadAll(os.Stdin)
			CheckError(err)
			body = string(content)
		} else {
			body = bodyFlag
		}

		provider := providers[0]
		validation, err := plugin.ValidateCustomType(provider, itemType, oneliner, topic, body, refs, validateTimeout)
		CheckError(err)
		if validation.Stderr != "" {
			fmt.Fprintf(os.Stderr, "Warning: plugin %s validation stderr: %s\n", provider.Name, validation.Stderr)
		}

		runPluginHooks(plugin.HookBeforeCreate, map[string]interface{}{
			"type":     itemType,
			"oneliner": oneliner,
			"topic":    topic,
			"refs":     refs,
		})

		s := store.New(doryRoot)
		defer s.Close()

		id, err := s.CreateCustom(itemType, oneliner, topic, body, refs)
		CheckError(err)

		runPluginHooks(plugin.HookAfterCreate, map[string]interface{}{
			"id":       id,
			"type":     itemType,
			"oneliner": oneliner,
			"topic":    topic,
			"refs":     refs,
		})

		payload := map[string]interface{}{
			"id":       id,
			"type":     itemType,
			"status":   "created",
			"oneliner": oneliner,
			"topic":    topic,
			"plugin":   provider.Name,
			"validation": map[string]interface{}{
				"valid":       validation.Valid,
				"message":     validation.Message,
				"errors":      validation.Errors,
				"duration_ms": validation.DurationMS,
			},
		}

		OutputResult(cmd, payload, func() {
			fmt.Printf("Created %s (%s via plugin %s)\n", id, itemType, provider.Name)
		})
	},
}

func init() {
	typeCreateCmd.Flags().StringP("topic", "t", "", "Optional topic for this item")
	typeCreateCmd.Flags().StringP("body", "b", "", "Full markdown body content (use - to read from stdin)")
	typeCreateCmd.Flags().StringSliceP("refs", "R", []string{}, "References to other knowledge items (e.g., L-01JX...,D-01JY...)")
	typeCreateCmd.Flags().Duration("validate-timeout", 2*time.Second, "Custom type validation timeout")
	typeCmd.AddCommand(typeCreateCmd)
}
