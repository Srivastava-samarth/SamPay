package services

import (
	"os"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/testutils"
)

var testServices *Services

func TestMain(m *testing.M) {
	var err error

	db, err := testutils.SetupTestDB()
	if err != nil {
		panic(err)
	}

	// Initialize repositories
	repo := repositories.NewRepository(db)
	jwtService := middlewares.NewJwt(&config.JWTConfig{})

	notificationService, err := notifications.NewEmailService(config.SMTPConfig{})
	if err != nil {
		panic(err)
	}

	testServices = NewServices(
		db,
		repo,
		jwtService,
		notificationService,
	)

	code := m.Run()

	if err := testutils.CleanupTestDB(db); err != nil {
		panic(err)
	}

	os.Exit(code)
}
