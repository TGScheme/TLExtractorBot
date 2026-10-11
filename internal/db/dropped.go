package db

func (db *DB) LoadDroppedObjects() (map[string]int64, error) {
	rows, err := db.DroppedStore.ListDroppedObjects()
	if err != nil {
		return nil, err
	}
	dropped := make(map[string]int64, len(rows))
	for _, row := range rows {
		dropped[formatHexID(row.Crc32)] = row.Build
	}
	return dropped, nil
}

func (db *DB) SetDroppedObject(id string, build int64) error {
	crc32, ok := parseHexID(id)
	if !ok {
		return nil
	}
	return db.DroppedStore.SetDroppedObject(crc32, build)
}

func (db *DB) RestoreDroppedObject(id string) error {
	crc32, ok := parseHexID(id)
	if !ok {
		return nil
	}
	return IgnoreNoRows(db.DroppedStore.DeleteDroppedObject(crc32))
}
