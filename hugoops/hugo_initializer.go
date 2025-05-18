package hugoops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// MountConfig represents a single mount entry in hugo_sync_hub.toml
// with a source path on disk and target path in Hugo's virtual FS.
type MountConfig struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// SyncConfig holds all mounts to apply
type SyncConfig struct {
	Mounts []MountConfig `toml:"mounts"`
}

// InitializeSite moves standard Hugo site files into 'hugo/' and applies mounts
// defined in hugo_sync_hub.toml by updating the module.mounts in config.toml.
func InitializeSite(root string) error {
	// 1. Path to Hugo site folder
	hugoDir := filepath.Join(root, "hugo")

	// 2. Ensure hugo/ exists
	if _, err := os.Stat(hugoDir); os.IsNotExist(err) {
		if err := os.Mkdir(hugoDir, 0755); err != nil {
			return fmt.Errorf("creating hugo directory: %w", err)
		}
	}

	// 3. Move standard Hugo folders if at root
	items := []string{"config.toml", "archetypes", "content", "layouts", "static", "data", "i18n", "themes"}
	for _, name := range items {
		src := filepath.Join(root, name)
		tgt := filepath.Join(hugoDir, name)
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, tgt); err != nil {
				return fmt.Errorf("moving %s to %s: %w", src, tgt, err)
			}
		}
	}

	// 4. Load mounts config
	cfgPath := filepath.Join(root, "hugo_sync_hub.toml")
	treeCfg, err := toml.LoadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("loading sync config: %w", err)
	}
	var syncCfg SyncConfig
	if err := treeCfg.Unmarshal(&syncCfg); err != nil {
		return fmt.Errorf("parsing sync config: %w", err)
	}

	// 5. Load Hugo config
	configPath := filepath.Join(hugoDir, "config.toml")
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		return fmt.Errorf("loading hugo config.toml: %w", err)
	}

	// 6. Ensure module section exists
	modVal := tree.Get("module")
	var modTree *toml.Tree
	if modVal == nil {
		newTree, err := toml.TreeFromMap(map[string]interface{}{})
		if err != nil {
			return fmt.Errorf("initializing module tree: %w", err)
		}
		modTree = newTree
		tree.Set("module", modTree)
	} else {
		modTree = modVal.(*toml.Tree)
	}

	// 7. Build mounts array
	modTree.Delete("mounts")
	var mountsArr []interface{}
	for _, m := range syncCfg.Mounts {
		mntTree, err := toml.TreeFromMap(map[string]interface{}{
			"source": m.Source,
			"target": m.Target,
		})
		if err != nil {
			return fmt.Errorf("creating mount entry for %s: %w", m.Source, err)
		}
		mountsArr = append(mountsArr, mntTree)
	}
	modTree.Set("mounts", mountsArr)

	// 8. Write updated Hugo config back
	out, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("opening hugo config for write: %w", err)
	}
	defer out.Close()
	tree.WriteTo(out)

	return nil
}
