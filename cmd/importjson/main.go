package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/db/models"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type legacyStorage struct {
	LastVersionCode   uint32                                        `json:"last_version_code"`
	LastTDeskID       int                                           `json:"last_tdesk_id"`
	LastTDLibID       int                                           `json:"last_tdlib_id"`
	LastCoreForkLayer int                                           `json:"last_corefork_layer"`
	MessageId         int64                                         `json:"message_id"`
	StableLayer       *types.TLFullScheme                           `json:"stable_layer,omitempty"`
	PreviewLayer      *types.TLFullScheme                           `json:"preview_layer,omitempty"`
	PatchedObjects    map[types.PatchOS]map[string]*types.PatchInfo `json:"patched_objects"`
	RecentLayers      []int                                         `json:"recent_layers"`
	ReleasedLayers    map[int]types.ReleasedLayer                   `json:"released_layers"`
}

func main() {
	path := "storage.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		gologging.Fatal(err)
	}
	var storage legacyStorage
	if err = json.Unmarshal(raw, &storage); err != nil {
		gologging.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		gologging.Fatal(err)
	}
	database, err := db.NewDB(cfg)
	if err != nil {
		gologging.Fatal(err)
	}

	if err = importSettings(database, storage); err != nil {
		gologging.Fatal(err)
	}
	if err = importSchemes(database, storage); err != nil {
		gologging.Fatal(err)
	}
	if err = importReleased(database, storage); err != nil {
		gologging.Fatal(err)
	}
	if err = importPatched(database, storage); err != nil {
		gologging.Fatal(err)
	}
	if err = importRecent(database, storage); err != nil {
		gologging.Fatal(err)
	}
	fmt.Println("import complete")
}

func importSettings(database *db.DB, storage legacyStorage) error {
	for _, step := range []struct {
		name string
		run  func() error
	}{
		{"last_version_code", func() error {
			return database.SettingsStore.SetLastVersionCode(int64(storage.LastVersionCode))
		}},
		{"last_tdesk_id", func() error { return database.SettingsStore.SetLastTDeskID(int64(storage.LastTDeskID)) }},
		{"last_tdlib_id", func() error { return database.SettingsStore.SetLastTDLibID(int64(storage.LastTDLibID)) }},
		{"last_corefork_layer", func() error {
			return database.SettingsStore.SetLastCoreForkLayer(int64(storage.LastCoreForkLayer))
		}},
		{"status_message_id", func() error { return database.SettingsStore.SetStatusMessageID(storage.MessageId) }},
	} {
		if err := step.run(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	fmt.Println("settings imported")
	return nil
}

func importSchemes(database *db.DB, storage legacyStorage) error {
	for _, entry := range []struct {
		role   models.SchemeRoleEnum
		scheme *types.TLFullScheme
	}{
		{models.SchemeRoleEnumStable, storage.StableLayer},
		{models.SchemeRoleEnumPreview, storage.PreviewLayer},
	} {
		if entry.scheme == nil {
			continue
		}
		if err := database.SaveScheme(entry.scheme, entry.role); err != nil {
			return fmt.Errorf("scheme %s: %w", entry.role, err)
		}
		fmt.Printf("scheme %s imported (layer %d, %d+%d main, %d+%d e2e)\n",
			entry.role, entry.scheme.Layer,
			len(entry.scheme.MainApi.Constructors), len(entry.scheme.MainApi.Methods),
			len(entry.scheme.E2EApi.Constructors), len(entry.scheme.E2EApi.Methods))
	}
	return nil
}

func importReleased(database *db.DB, storage legacyStorage) error {
	for layer, released := range storage.ReleasedLayers {
		if err := database.ReplaceReleasedLayer(layer, released); err != nil {
			return fmt.Errorf("released layer %d: %w", layer, err)
		}
	}
	fmt.Printf("released layers imported (%d layers)\n", len(storage.ReleasedLayers))
	return nil
}

func importPatched(database *db.DB, storage legacyStorage) error {
	count := 0
	for os, objects := range storage.PatchedObjects {
		for name, info := range objects {
			if err := database.PatchedStore.UpsertPatchedObject(
				models.PatchOsEnum(os), name,
				db.ParseID(info.OldConstructor),
				db.ParseID(info.PatchedConstructor),
			); err != nil {
				return fmt.Errorf("patched %s/%s: %w", os, name, err)
			}
			count++
		}
	}
	fmt.Printf("patched objects imported (%d)\n", count)
	return nil
}

func importRecent(database *db.DB, storage legacyStorage) error {
	for _, layer := range storage.RecentLayers {
		if err := db.IgnoreNoRows(database.RecentStore.AddRecentLayer(int64(layer))); err != nil {
			return fmt.Errorf("recent layer %d: %w", layer, err)
		}
	}
	fmt.Printf("recent layers imported (%d)\n", len(storage.RecentLayers))
	return nil
}
