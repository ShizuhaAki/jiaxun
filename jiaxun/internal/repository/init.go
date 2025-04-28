package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"jiaxun/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// HashPassword generates a bcrypt hash from a password string
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// InitDB initializes the database connection based on the provided config
// and creates the database schema using GORM's Auto Migration
func InitDB(driver, host string, port int, user, password, dbName string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Configure GORM with options
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table names
		},
		DisableForeignKeyConstraintWhenMigrating: false, // Enable foreign key constraints
	}

	// Initialize the database connection based on the driver type
	switch driver {
	case "sqlite3":
		// For SQLite, check if the file exists
		dbFile := fmt.Sprintf("%s.db", dbName)

		// SQLite automatically creates a new database file if it doesn't exist
		db, err = gorm.Open(sqlite.Open(dbFile), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		// Set SQLite pragmas for better performance
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database connection: %w", err)
		}
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
		sqlDB.Exec("PRAGMA foreign_keys=ON;")

	case "postgres":
		// First, try to connect to the target database
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)

		db, err = gorm.Open(postgres.Open(dsn), gormConfig)

		if err != nil {
			// If connection fails, connect to the postgres default database
			log.Printf("Couldn't connect to database %s, attempting to create it...", dbName)

			defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
				host, port, user, password)

			// Open a connection to the default postgres database
			sqlDB, err := sql.Open("postgres", defaultDSN)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to default postgres database: %w", err)
			}
			defer sqlDB.Close()

			// Check if the database exists
			var exists int
			err = sqlDB.QueryRow("SELECT 1 FROM pg_database WHERE datname = $1", dbName).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to check if database exists: %w", err)
			}

			// If the database doesn't exist, create it
			if exists == 0 {
				_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
				if err != nil {
					return nil, fmt.Errorf("failed to create database: %w", err)
				}
				log.Printf("Database %s created successfully", dbName)
			}

			// Now connect to the newly created database
			db, err = gorm.Open(postgres.Open(dsn), gormConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to PostgreSQL database after creation: %w", err)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Create a migrator instance
	migrator := db.Migrator()

	// Auto-migrate models using GORM schema
	log.Println("Running database migrations...")

	// Register all models to be migrated
	models := []interface{}{
		&model.User{},
		// Add other models here as needed
	}

	// Perform the migrations
	err = db.AutoMigrate(models...)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	// Create database indexes (if they don't already exist)
	if !migrator.HasIndex(&model.User{}, "idx_user_email") {
		err = migrator.CreateIndex(&model.User{}, "idx_user_email")
		if err != nil {
			log.Printf("Warning: failed to create index idx_user_email: %v", err)
		}
	}

	// Check if root user exists, if not create it
	var rootUser model.User
	result := db.Where("username = ?", "root").First(&rootUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create root user
			log.Println("Creating root user...")

			// Hash the password "jiaxun"
			hashedPassword, err := hashPassword("jiaxun")
			if err != nil {
				log.Printf("Warning: failed to hash password for root user: %v", err)
				return db, nil // Continue despite the error
			}

			// Create user
			rootUser = model.User{
				Username:  "root",
				Email:     "root@example.com",
				Password:  hashedPassword,
				FullName:  "System Administrator",
				Role:      "admin",
				CreatedAt: time.Now(),
			}

			if err := db.Create(&rootUser).Error; err != nil {
				log.Printf("Warning: failed to create root user: %v", err)
			} else {
				log.Println("Root user created successfully")
			}
		} else {
			// Some other error occurred when querying for the root user
			log.Printf("Warning: error checking for root user: %v", result.Error)
		}
	} else {
		log.Println("Root user already exists")
	}

	log.Printf("Successfully connected to the %s database!", driver)
	return db, nil
}
