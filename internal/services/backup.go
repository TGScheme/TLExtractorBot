package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/db"
)

func (s *Service) Backup() (data []byte, filename string, err error) {
	data, err = db.ExecuteBackup(s.cfg)
	if err != nil {
		return nil, "", err
	}
	return data, fmt.Sprintf("backup_%s.sql", time.Now().Format("2006-01-02_15-04-05")), nil
}

func (s *Service) backupDatabase() {
	data, filename, err := s.Backup()
	if err != nil {
		gologging.Error("backup: ", err)
		return
	}
	if err = s.bot.SendDocument(
		s.cfg.BackupChatID, filename, data,
		fmt.Sprintf("<b>Automatic backup</b> — %s", time.Now().UTC().Format(time.RFC1123)),
	); err != nil {
		gologging.Error("backup upload: ", err)
	}
}

func (s *Service) RestoreBackup(data []byte) error {
	if !s.building.CompareAndSwap(false, true) {
		return errors.New("an extraction is running, retry once it is done")
	}
	defer s.building.Store(false)

	if err := db.RestoreBackup(s.cfg, data); err != nil {
		return err
	}
	s.db.Pool.Reset()
	s.scheme.InvalidateCaches()
	return nil
}
