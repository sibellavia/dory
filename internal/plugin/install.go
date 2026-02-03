package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InstallOptions controls plugin installation behavior.
type InstallOptions struct {
	Force bool
}

// Install copies a plugin directory into .dory/plugins/<manifest.name>.
// source may be either a plugin directory or a plugin.yaml path.
func Install(doryRoot, source string, opts InstallOptions) (*PluginInfo, error) {
	srcDir, err := normalizePluginSource(source)
	if err != nil {
		return nil, err
	}

	srcManifestPath := ManifestPath(srcDir)
	manifest, err := LoadManifest(srcManifestPath)
	if err != nil {
		return nil, err
	}

	pluginsDir := PluginsDirPath(doryRoot)
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, err
	}

	dstDir := filepath.Join(pluginsDir, manifest.Name)
	if st, err := os.Stat(dstDir); err == nil && st.IsDir() {
		if !opts.Force {
			return nil, fmt.Errorf("plugin %q already exists at %s (use --force to overwrite)", manifest.Name, dstDir)
		}
		if err := os.RemoveAll(dstDir); err != nil {
			return nil, err
		}
	} else if err == nil && !st.IsDir() {
		return nil, fmt.Errorf("cannot install plugin: destination %s exists and is not a directory", dstDir)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		_ = os.RemoveAll(dstDir)
		return nil, err
	}

	dstManifestPath := ManifestPath(dstDir)
	dstManifest, err := LoadManifest(dstManifestPath)
	if err != nil {
		_ = os.RemoveAll(dstDir)
		return nil, err
	}

	cfg, err := LoadProjectConfig(doryRoot)
	if err != nil {
		return nil, err
	}

	return &PluginInfo{
		Name:         dstManifest.Name,
		Version:      dstManifest.Version,
		APIVersion:   dstManifest.APIVersion,
		Description:  dstManifest.Description,
		Command:      append([]string(nil), dstManifest.Command...),
		Capabilities: dstManifest.Capabilities,
		Enabled:      cfg.Enabled[dstManifest.Name],
		Dir:          dstDir,
		ManifestPath: dstManifestPath,
	}, nil
}

func normalizePluginSource(source string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("source path is required")
	}

	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return source, nil
	}
	if filepath.Base(source) == ManifestFileName {
		return filepath.Dir(source), nil
	}
	return "", fmt.Errorf("source must be a plugin directory or %s path", ManifestFileName)
}

func copyDir(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(dstPath, info.Mode().Perm())
		}

		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not supported in plugin install: %s", path)
		}

		return copyFile(path, dstPath, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
