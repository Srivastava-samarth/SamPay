package testutils

import (
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(envPath string) (*gorm.DB, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, err
	}

	host := os.Getenv("SAMPAY_DB_HOST")
	port := os.Getenv("SAMPAY_DB_PORT")
	user := os.Getenv("SAMPAY_DB_USER")
	password := os.Getenv("SAMPAY_DB_PASSWORD")
	sslMode := os.Getenv("SAMPAY_DB_SSLMODE")
	timeZone := os.Getenv("SAMPAY_DB_TIMEZONE")

	dsn := "host=" + host +
		" port=" + port +
		" user=" + user +
		" password=" + password +
		" dbname=sampay_test" +
		" sslmode=" + sslMode +
		" TimeZone=" + timeZone

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CleanupTestDB(db *gorm.DB) error {
	return db.Exec(`
		DO $$
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN (
				SELECT tablename
				FROM pg_tables
				WHERE schemaname = 'public'
			)
			LOOP
				EXECUTE 'TRUNCATE TABLE public.' ||
					quote_ident(r.tablename) ||
					' RESTART IDENTITY CASCADE';
			END LOOP;
		END $$;
	`).Error
}
