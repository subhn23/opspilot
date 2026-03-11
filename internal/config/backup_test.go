package config

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConfigureWALArchiving(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	// Expectations
	mock.ExpectExec("ALTER SYSTEM SET archive_mode = 'on'").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER SYSTEM SET archive_command = \\$1").WithArgs("test ! -f /tmp/%f && cp %p /tmp/%f").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER SYSTEM SET wal_level = 'replica'").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_reload_conf\\(\\)").WillReturnResult(sqlmock.NewResult(0, 0))
	
	// Audit log insertion
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO \"audit_logs\"").WithArgs(sqlmock.AnyArg(), "BACKUP_CONFIGURED", "PostgreSQL", "/tmp", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err = ConfigureWALArchiving(gormDB, "/tmp")
	if err != nil {
		t.Errorf("ConfigureWALArchiving failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}

func TestVerifyBackupStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	mock.ExpectQuery("SHOW archive_mode").WillReturnRows(sqlmock.NewRows([]string{"archive_mode"}).AddRow("on"))
	mock.ExpectQuery("SHOW archive_command").WillReturnRows(sqlmock.NewRows([]string{"archive_command"}).AddRow("test_cmd"))
	mock.ExpectQuery("SELECT last_archived_wal, last_failed_wal FROM pg_stat_archiver").
		WillReturnRows(sqlmock.NewRows([]string{"last_archived_wal", "last_failed_wal"}).AddRow("000000010000000000000001", ""))

	status, err := VerifyBackupStatus(gormDB)
	if err != nil {
		t.Errorf("VerifyBackupStatus failed: %v", err)
	}

	if status.ArchiveMode != "on" {
		t.Errorf("expected archive_mode on, got %s", status.ArchiveMode)
	}
	if status.ArchiveCommand != "test_cmd" {
		t.Errorf("expected archive_command test_cmd, got %s", status.ArchiveCommand)
	}
}
