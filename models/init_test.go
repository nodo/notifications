package models_test

import (
	"database/sql"
	"testing"

	"github.com/cloudfoundry-incubator/notifications/application"
	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestModelsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}

var sqlDB *sql.DB

var _ = BeforeEach(func() {
	env := application.NewEnvironment()

	var err error
	sqlDB, err = sql.Open("mysql", env.DatabaseURL)
	Expect(err).NotTo(HaveOccurred())
})

func TruncateTables() {
	database := db.NewDatabase(sqlDB, db.Config{})
	env := application.NewEnvironment()
	dbMigrator := models.DatabaseMigrator{}
	dbMigrator.Migrate(database.RawConnection(), env.ModelMigrationsPath)
	models.Setup(database)

	connection := database.Connection().(*db.Connection)
	err := connection.TruncateTables()
	if err != nil {
		panic(err)
	}
}

func findReceipt(conn db.ConnectionInterface, userGUID, clientID, kindID string) (models.Receipt, error) {
	receipt := models.Receipt{}
	err := conn.SelectOne(&receipt, "SELECT * FROM  `receipts` WHERE `user_guid` = ? AND `client_id` = ? AND `kind_id` = ?", userGUID, clientID, kindID)
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func createReceipt(conn db.ConnectionInterface, receipt models.Receipt) (models.Receipt, error) {
	err := conn.Insert(&receipt)
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}
