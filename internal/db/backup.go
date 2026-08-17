package db

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/TGScheme/TLExtractorBot/internal/config"
)

func ExecuteBackup(cfg *config.Config) ([]byte, error) {
	var out, stdErr bytes.Buffer
	cmd := exec.Command(
		"pg_dump",
		"-h", cfg.DBHost,
		"-p", strconv.Itoa(cfg.DBPort),
		"-U", cfg.DBUser,
		"-d", cfg.DBName,
		"-Fc",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.DBPassword))
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pg_dump: %w: %s", err, stdErr.String())
	}
	return out.Bytes(), nil
}

func RestoreBackup(cfg *config.Config, dumpData []byte) error {
	cmd := exec.Command(
		"pg_restore",
		"-h", cfg.DBHost,
		"-p", strconv.Itoa(cfg.DBPort),
		"-U", cfg.DBUser,
		"-d", cfg.DBName,
		"--clean",
		"--if-exists",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.DBPassword))
	cmd.Stdin = bytes.NewReader(dumpData)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_restore: %w: %s", err, out.String())
	}
	return nil
}
