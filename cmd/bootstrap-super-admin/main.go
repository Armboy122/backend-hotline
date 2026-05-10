package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/feature/user/bootstrap"
	"backend-hotlines3/internal/feature/user/repository"
	dbpkg "backend-hotlines3/pkg/db"
)

func main() {
	username := flag.String("username", os.Getenv("SUPER_ADMIN_USERNAME"), "super admin username (or SUPER_ADMIN_USERNAME)")
	password := flag.String("password", os.Getenv("SUPER_ADMIN_PASSWORD"), "super admin password (or SUPER_ADMIN_PASSWORD)")
	flag.Parse()
	if *username == "" || *password == "" {
		log.Fatal("username and password are required")
	}

	ctx := context.Background()
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := dbpkg.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	repo := repository.NewRepository(db)
	if err := bootstrap.CreateSuperAdmin(ctx, repo, bootstrap.Input{Username: *username, Password: *password}); err != nil {
		if errors.Is(err, bootstrap.ErrSuperAdminExists) {
			log.Fatal("refusing to create super_admin: an active super_admin already exists")
		}
		log.Fatalf("create super_admin: %v", err)
	}
	fmt.Println("super_admin created")
}
