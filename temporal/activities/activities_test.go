package activities

import (
	"os"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/testutils"
	"gorm.io/gorm"
)

var (
	db           *gorm.DB
	testRepo     *repositories.Repository
	testServices *services.Services
)

func TestMain(m *testing.M) {
	var err error

	db, err = testutils.SetupTestDB("../../.env")
	if err != nil {
		panic(err)
	}

	testRepo = repositories.NewRepository(db)

	jwtService := middlewares.NewJwt(&config.JWTConfig{
		AccessExpiry:  "3600",
		RefreshExpiry: "86400",
	})

	notificationService, err := notifications.NewEmailService(config.SMTPConfig{})
	if err != nil {
		panic(err)
	}

	testServices = services.NewServices(
		db,
		testRepo,
		jwtService,
		notificationService,
	)

	code := m.Run()

	if err := testutils.CleanupTestDB(db); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func cleanTestDB(t *testing.T) {
	t.Helper()

	if err := testutils.CleanupTestDB(db); err != nil {
		t.Fatalf("failed to clean test DB: %v", err)
	}
}
