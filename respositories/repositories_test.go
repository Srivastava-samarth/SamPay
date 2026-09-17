package repositories

import (
	"os"
	"testing"

	"github.com/Srivastava-samarth/sampay/testutils"
	"gorm.io/gorm"
)

var db *gorm.DB
var testRepo *Repository

func TestMain(m *testing.M) {
	var err error

	db, err = testutils.SetupTestDB("../.env")
	if err != nil {
		panic(err)
	}

	testRepo = &Repository{
		DB: db,
	}

	code := m.Run()

	if err := testutils.CleanupTestDB(db); err != nil {
		panic(err)
	}

	os.Exit(code)
}
