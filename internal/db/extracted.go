package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func parseHexID(id string) (int64, bool) {
	value, err := strconv.ParseUint(id, 16, 32)
	if err != nil {
		return 0, false
	}
	return int64(value), true
}

func formatHexID(crc32 int64) string {
	return fmt.Sprintf("%x", uint32(crc32))
}

func (db *DB) SaveExtractedObjects(source string, ids []string) error {
	ctx := context.Background()
	if err := IgnoreNoRows(db.ExtractedStore.DeleteExtractedObjects(source)); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]bool, len(ids))
	rows := make([][]any, 0, len(ids))
	for _, id := range ids {
		crc32, ok := parseHexID(id)
		if !ok || seen[crc32] {
			continue
		}
		seen[crc32] = true
		rows = append(rows, []any{source, crc32})
	}
	if _, err := db.Pool.CopyFrom(ctx,
		pgx.Identifier{"extracted_objects"},
		[]string{"source", "crc32"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return fmt.Errorf("copy extracted_objects: %w", err)
	}
	return nil
}

func (db *DB) LoadExtractedObjects(source string) ([]string, error) {
	ids, err := db.ExtractedStore.ListExtractedObjects(source)
	if err != nil {
		return nil, err
	}
	extracted := make([]string, 0, len(ids))
	for _, id := range ids {
		extracted = append(extracted, formatHexID(id))
	}
	return extracted, nil
}
