package models

type MigrationStatus string

const (
	BEFORE    MigrationStatus = "Before"
	STARTED   MigrationStatus = "Started"
	COMPLETED MigrationStatus = "Completed"
	FAILED    MigrationStatus = "Failed"
)
