package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/TGScheme/TLExtractorBot/internal/db/models"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"github.com/jackc/pgx/v5"
)

func IgnoreNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func ParseID(id string) int64 {
	n, _ := strconv.ParseInt(id, 10, 64)
	return int64(uint32(n))
}

func FormatID(crc32 int64) string {
	return strconv.FormatInt(int64(int32(uint32(crc32))), 10)
}

func (db *DB) LoadScheme(role models.SchemeRoleEnum) (*types.TLFullScheme, error) {
	head, err := db.SchemeStore.GetSchemeByRole(role)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}
	objects, err := db.SchemeStore.GetSchemeObjects(head.ID)
	if err != nil {
		return nil, err
	}
	params, err := db.SchemeStore.GetSchemeParams(head.ID)
	if err != nil {
		return nil, err
	}
	byObject := make(map[int64][]types.Parameter, len(objects))
	for _, p := range params {
		byObject[p.ObjectID] = append(byObject[p.ObjectID], types.Parameter{Name: p.Name, Type: p.ParamType})
	}

	full := &types.TLFullScheme{Layer: int(head.Layer), IsSync: head.IsSync}
	for _, o := range objects {
		base := types.TLBase{
			ID:          FormatID(o.Crc32),
			Params:      byObject[o.ID],
			Type:        o.Result,
			Layer:       int(o.Layer),
			ForceSecret: o.ForceSecret,
		}
		target := &full.MainApi
		if o.API == models.ApiKindEnumE2e {
			target = &full.E2EApi
		}
		if o.Kind == models.TlKindEnumConstructor {
			target.Constructors = append(target.Constructors, &types.TLConstructor{TLBase: base, Predicate: o.ObjectName})
		} else {
			target.Methods = append(target.Methods, &types.TLMethod{TLBase: base, Method: o.ObjectName})
		}
	}
	return full, nil
}

func (db *DB) SaveScheme(full *types.TLFullScheme, role models.SchemeRoleEnum) error {
	ctx := context.Background()
	id, err := db.SchemeStore.CreateScheme(int64(full.Layer), full.IsSync)
	if err != nil {
		return err
	}

	type pending struct {
		api    models.ApiKindEnum
		kind   models.TlKindEnum
		object types.TLInterface
	}
	var flat []pending
	for _, api := range []struct {
		kind   models.ApiKindEnum
		scheme types.TLScheme
	}{{models.ApiKindEnumMain, full.MainApi}, {models.ApiKindEnumE2e, full.E2EApi}} {
		for _, o := range api.scheme.GetConstructors() {
			flat = append(flat, pending{api.kind, models.TlKindEnumConstructor, o})
		}
		for _, o := range api.scheme.GetMethods() {
			flat = append(flat, pending{api.kind, models.TlKindEnumMethod, o})
		}
	}

	rows := make([][]any, 0, len(flat))
	for i, p := range flat {
		rows = append(rows, []any{
			id, string(p.api), string(p.kind), ParseID(p.object.Constructor()),
			p.object.Package(), p.object.Result(), int64(p.object.GetLayer()),
			p.object.IsSecret(), int64(i),
		})
	}
	if _, err = db.Pool.CopyFrom(ctx,
		pgx.Identifier{"tl_objects"},
		[]string{"scheme_id", "api", "kind", "crc32", "object_name", "result", "layer", "force_secret", "position"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return fmt.Errorf("copy tl_objects: %w", err)
	}

	ids, err := db.objectIDs(ctx, id)
	if err != nil {
		return err
	}
	var paramRows [][]any
	for i, p := range flat {
		for j, param := range p.object.Parameters() {
			paramRows = append(paramRows, []any{ids[i], int64(j), param.Name, param.Type})
		}
	}
	if len(paramRows) > 0 {
		if _, err = db.Pool.CopyFrom(ctx,
			pgx.Identifier{"tl_params"},
			[]string{"object_id", "position", "name", "param_type"},
			pgx.CopyFromRows(paramRows),
		); err != nil {
			return fmt.Errorf("copy tl_params: %w", err)
		}
	}

	if err = IgnoreNoRows(db.SchemeStore.ClearSchemeRole(role)); err != nil {
		return err
	}
	if err = db.SchemeStore.SetSchemeRole(role, id); err != nil {
		return err
	}
	return IgnoreNoRows(db.SchemeStore.DeleteUnreferencedSchemes())
}

func (db *DB) objectIDs(ctx context.Context, schemeID int64) ([]int64, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id FROM tl_objects WHERE scheme_id = $1 ORDER BY position`, schemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *DB) ReplaceReleasedLayer(layer int, released types.ReleasedLayer) error {
	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err = tx.Exec(ctx, `DELETE FROM released_objects WHERE layer = $1`, layer); err != nil {
		return err
	}
	seen := make(map[[2]any]bool)
	var rows [][]any
	add := func(kind models.TlKindEnum, objects []types.ReleasedConstructor) {
		for _, o := range objects {
			key := [2]any{string(kind), ParseID(o.ID)}
			if seen[key] {
				continue
			}
			seen[key] = true
			rows = append(rows, []any{int64(layer), string(kind), ParseID(o.ID)})
		}
	}
	add(models.TlKindEnumConstructor, released.Constructors)
	add(models.TlKindEnumMethod, released.Methods)

	if len(rows) > 0 {
		if _, err = tx.CopyFrom(ctx,
			pgx.Identifier{"released_objects"},
			[]string{"layer", "kind", "crc32"},
			pgx.CopyFromRows(rows),
		); err != nil {
			return fmt.Errorf("copy released_objects: %w", err)
		}
	}
	return tx.Commit(ctx)
}
