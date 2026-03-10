package models

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedSystemData creates the default roles and permissions
func SeedSystemData(db *gorm.DB) {
	var count int64
	db.Model(&Role{}).Count(&count)
	if count > 0 {
		return // Already seeded
	}

	log.Println("Seeding initial system data...")

	// 1. Create Base Permissions
	permissions := []Permission{
		{Slug: "proxy:read", Module: "OpsProxy"},
		{Slug: "proxy:write", Module: "OpsProxy"},
		{Slug: "deploy:read", Module: "OpsDeploy"},
		{Slug: "deploy:write", Module: "OpsDeploy"},
		{Slug: "infra:admin", Module: "Terraform"},
		{Slug: "system:admin", Module: "Auth"},
	}

	for i := range permissions {
		db.FirstOrCreate(&permissions[i], Permission{Slug: permissions[i].Slug})
	}

	// 2. Define Roles
	roles := []Role{
		{
			ID:          uuid.New(),
			Name:        "Master Admin",
			Permissions: permissions, // All permissions
		},
		{
			ID:   uuid.New(),
			Name: "Developer",
			Permissions: []Permission{
				permissions[0], // proxy:read
				permissions[2], // deploy:read
				permissions[3], // deploy:write
			},
		},
		{
			ID:   uuid.New(),
			Name: "Viewer",
			Permissions: []Permission{
				permissions[0], // proxy:read
				permissions[2], // deploy:read
			},
		},
	}

	for _, r := range roles {
		db.Create(&r)
	}

	log.Println("Seeding completed: Default roles created.")
}
