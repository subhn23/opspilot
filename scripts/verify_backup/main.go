package main

import (
	"fmt"
	"log"
	"opspilot/internal/config"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Setup sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open gorm db: %v", err)
	}

	// 2. Mock expectations for ConfigureWALArchiving
	mock.ExpectExec("ALTER SYSTEM SET archive_mode = 'on'").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER SYSTEM SET archive_command = \\$1").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER SYSTEM SET wal_level = 'replica'").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_reload_conf\\(\\)").WillReturnResult(sqlmock.NewResult(0, 0))
	
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO \"audit_logs\"").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// 3. Mock expectations for VerifyBackupStatus
	mock.ExpectQuery("SHOW archive_mode").WillReturnRows(sqlmock.NewRows([]string{"archive_mode"}).AddRow("on"))
	mock.ExpectQuery("SHOW archive_command").WillReturnRows(sqlmock.NewRows([]string{"archive_command"}).AddRow("test_cmd"))
	mock.ExpectQuery("SELECT last_archived_wal, last_failed_wal FROM pg_stat_archiver").
		WillReturnRows(sqlmock.NewRows([]string{"last_archived_wal", "last_failed_wal"}).AddRow("000000010000000000000001", ""))

	// 4. Execute
	fmt.Println("Step 1: Configuring WAL Archiving...")
	err = config.ConfigureWALArchiving(gormDB, "/mnt/backup")
	if err != nil {
		log.Fatalf("Configure failed: %v", err)
	}
	fmt.Println("Configuration successful.")

	fmt.Println("\nStep 2: Verifying Backup Status...")
	status, err := config.VerifyBackupStatus(gormDB)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}
	fmt.Printf("Current Status:\n - Archive Mode: %s\n - Archive Command: %s\n - Last Archived: %s\n", 
		status.ArchiveMode, status.ArchiveCommand, status.LastArchived)

	fmt.Println("\nManual verification of Database Resilience logic PASSED.")
}
