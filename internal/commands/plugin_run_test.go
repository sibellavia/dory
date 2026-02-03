package commands

import "testing"

func TestResolvePluginCommand(t *testing.T) {
	t.Run("single command default", func(t *testing.T) {
		name, args, err := resolvePluginCommand(nil, []string{"sync"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if name != "sync" || len(args) != 0 {
			t.Fatalf("unexpected result: %q %v", name, args)
		}
	})

	t.Run("single command treat requested as args", func(t *testing.T) {
		name, args, err := resolvePluginCommand([]string{"--all"}, []string{"sync"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if name != "sync" || len(args) != 1 || args[0] != "--all" {
			t.Fatalf("unexpected result: %q %v", name, args)
		}
	})

	t.Run("multi command requires command", func(t *testing.T) {
		if _, _, err := resolvePluginCommand(nil, []string{"sync", "prune"}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("multi command exact match", func(t *testing.T) {
		name, args, err := resolvePluginCommand([]string{"prune", "--dry-run"}, []string{"sync", "prune"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if name != "prune" || len(args) != 1 || args[0] != "--dry-run" {
			t.Fatalf("unexpected result: %q %v", name, args)
		}
	})

	t.Run("no capabilities", func(t *testing.T) {
		if _, _, err := resolvePluginCommand([]string{"a"}, nil); err == nil {
			t.Fatal("expected error")
		}
	})
}
