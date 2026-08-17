package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/db/models"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type patchSet struct {
	os      types.PatchOS
	entries map[string]*types.PatchInfo
	dirty   map[string]bool
	removed map[string]bool
}

func loadPatchSet(database *db.DB, os types.PatchOS) (*patchSet, error) {
	rows, err := database.PatchedStore.ListPatchedObjects()
	if err != nil {
		return nil, err
	}
	set := &patchSet{
		os:      os,
		entries: make(map[string]*types.PatchInfo),
		dirty:   make(map[string]bool),
		removed: make(map[string]bool),
	}
	for _, row := range rows {
		if string(row.Os) != string(os) {
			continue
		}
		set.entries[row.ObjectName] = &types.PatchInfo{
			OldConstructor:     db.FormatID(row.OldConstructor),
			PatchedConstructor: db.FormatID(row.NewConstructor),
		}
	}
	return set, nil
}

func (p *patchSet) get(name string) (*types.PatchInfo, bool) {
	info, ok := p.entries[name]
	return info, ok
}

func (p *patchSet) set(name string, info *types.PatchInfo) {
	p.entries[name] = info
	p.dirty[name] = true
	delete(p.removed, name)
}

func (p *patchSet) delete(name string) {
	delete(p.entries, name)
	delete(p.dirty, name)
	p.removed[name] = true
}

func (p *patchSet) flush(database *db.DB) error {
	for name := range p.removed {
		if err := db.IgnoreNoRows(database.PatchedStore.DeletePatchedObject(models.PatchOsEnum(p.os), name)); err != nil {
			return err
		}
	}
	for name := range p.dirty {
		info := p.entries[name]
		if err := database.PatchedStore.UpsertPatchedObject(
			models.PatchOsEnum(p.os), name,
			db.ParseID(info.OldConstructor),
			db.ParseID(info.PatchedConstructor),
		); err != nil {
			return err
		}
	}
	return nil
}
