package config

import (
	"fmt"
	"log"
	"opspilot/internal/audit"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConfigureWALArchiving sets up PostgreSQL for point-in-time recovery
func ConfigureWALArchiving(db *gorm.DB, backupPath string) error {
	log.Println("Configuring Postgres WAL archiving...")

	// 1. Set archive_mode to on
	if err := db.Exec("ALTER SYSTEM SET archive_mode = 'on'").Error; err != nil {
		return fmt.Errorf("failed to set archive_mode: %w", err)
	}

	// 2. Set archive_command
	// Example command: test ! -f /mnt/server/archivedir/%f && cp %p /mnt/server/archivedir/%f
	archiveCmd := fmt.Sprintf("test ! -f %s/%%f && cp %%p %s/%%f", backupPath, backupPath)
	if err := db.Exec("ALTER SYSTEM SET archive_command = ?", archiveCmd).Error; err != nil {
		return fmt.Errorf("failed to set archive_command: %w", err)
	}

	// 3. Set wal_level to replica
	if err := db.Exec("ALTER SYSTEM SET wal_level = 'replica'").Error; err != nil {
		return fmt.Errorf("failed to set wal_level: %w", err)
	}

	// 4. Reload configuration
	if err := db.Exec("SELECT pg_reload_conf()").Error; err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	log.Println("WAL archiving configured successfully")
	audit.LogAction(db, uuid.Nil, "BACKUP_CONFIGURED", "PostgreSQL", "", backupPath)

	return nil
}

type BackupStatus struct {
	ArchiveMode    string
	ArchiveCommand string
	LastArchived   string
	LastFailed     string
}

// VerifyBackupStatus checks the current state of WAL archiving
func VerifyBackupStatus(db *gorm.DB) (*BackupStatus, error) {
	var status BackupStatus

	// Check archive_mode
	db.Raw("SHOW archive_mode").Scan(&status.ArchiveMode)
	
	// Check archive_command
	db.Raw("SHOW archive_command").Scan(&status.ArchiveCommand)

	// Check pg_stat_archiver
	type statArchiver struct {
		LastArchivedWal string
		LastFailedWal   string
	}
	var sa statArchiver
	if err := db.Raw("SELECT last_archived_wal, last_failed_wal FROM pg_stat_archiver").Scan(&sa).Error; err == nil {
		status.LastArchived = sa.LastArchivedWal
		status.LastFailed = sa.LastFailedWal
	}

	return &status, nil
}
