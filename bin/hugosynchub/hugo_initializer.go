package main

import (
	"fmt"
	"os"
	"os/exec"
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

// InitializeSite restructures a Hugo site for GitHub Pages:
// - Ensures a hugo/ folder in the project root
// - If no existing site at root, runs `hugo new site hugo/ --force`
// - Otherwise moves existing Hugo files into hugo/
// - Applies mounts from hugo_sync_hub.toml to the Hugo config file (config.toml or hugo.toml)
func InitializeSite(moduleDir string) error {
	// Determine project root (parent of module)
	projectRoot := filepath.Dir(moduleDir)
	// Hugo site destination
	hugoDir := filepath.Join(projectRoot, "hugo")

	// 1) Ensure hugo/ directory exists
	if _, err := os.Stat(hugoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hugoDir, 0755); err != nil {
			return fmt.Errorf("creating hugo directory: %w", err)
		}
	}

	// 2) Detect existing Hugo config at project root
	var rootConfigName string
	for _, name := range []string{"config.toml", "hugo.toml"} {
		if _, err := os.Stat(filepath.Join(projectRoot, name)); err == nil {
			rootConfigName = name
			break
		}
	}
	if rootConfigName == "" {
		// No existing site: initialize new Hugo site into hugo/
		cmd := exec.Command("hugo", "new", "site", hugoDir, "--force")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("initializing new Hugo site: %w", err)
		}
	} else {
		// 3) Move standard Hugo files into hugo/
		items := []string{rootConfigName, "archetypes", "content", "layouts", "static", "data", "i18n", "themes"}
		for _, name := range items {
			src := filepath.Join(projectRoot, name)
			tgt := filepath.Join(hugoDir, name)
			if _, err := os.Stat(src); err == nil {
				if err := os.Rename(src, tgt); err != nil {
					return fmt.Errorf("moving %s into hugo/: %w", name, err)
				}
			}
		}
		rootConfigName = rootConfigName // moved into hugo/
	}

	// 4) Load sync mounts configuration
	cfgPath := filepath.Join(moduleDir, "hugo_sync_hub.toml")
	treeCfg, err := toml.LoadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("loading sync config: %w", err)
	}
	var syncCfg SyncConfig
	if err := treeCfg.Unmarshal(&syncCfg); err != nil {
		return fmt.Errorf("parsing sync config: %w", err)
	}

	// 5) Find Hugo config file under hugo/
	var hugoConfig string
	for _, name := range []string{"config.toml", "hugo.toml"} {
		path := filepath.Join(hugoDir, name)
		if _, err := os.Stat(path); err == nil {
			hugoConfig = path
			break
		}
	}
	if hugoConfig == "" {
		return fmt.Errorf("no Hugo config file found in %s", hugoDir)
	}

	// 6) Load Hugo config
	tree, err := toml.LoadFile(hugoConfig)
	if err != nil {
		return fmt.Errorf("loading Hugo config: %w", err)
	}

	// 7) Ensure [module] exists
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

	// 8) Ensure module.imports includes DocuAPI theme
	// Reset imports and add DocuAPI import
	importsArr := []interface{}{}
	docuapiImport, err := toml.TreeFromMap(map[string]interface{}{"path": "github.com/bep/docuapi/v2"})
	if err != nil {
		return fmt.Errorf("creating import entry: %w", err)
	}
	importsArr = append(importsArr, docuapiImport)
	modTree.Set("imports", importsArr)

	// 9) Rebuild mounts list
	modTree.Delete("mounts")
	var arr []interface{}
	for _, m := range syncCfg.Mounts {
		mntTree, err := toml.TreeFromMap(map[string]interface{}{"source": m.Source, "target": m.Target})
		if err != nil {
			return fmt.Errorf("creating mount for %s: %w", m.Source, err)
		}
		arr = append(arr, mntTree)
	}
	modTree.Set("mounts", arr)

	// 9) Write updated Hugo config back
	out, err := os.Create(hugoConfig)
	if err != nil {
		return fmt.Errorf("opening Hugo config for write: %w", err)
	}
	defer out.Close()
	tree.WriteTo(out)

	return nil
}
