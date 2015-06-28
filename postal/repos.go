package postal

import "github.com/cloudfoundry-incubator/notifications/models"

type messagesRepo interface {
	Upsert(models.ConnectionInterface, models.Message) (models.Message, error)
}

type kindsRepo interface {
	Find(models.ConnectionInterface, string, string) (models.Kind, error)
}

type receiptsRepo interface {
	CreateReceipts(models.ConnectionInterface, []string, string, string) error
}

type templatesRepo interface {
	FindByID(models.ConnectionInterface, string) (models.Template, error)
}

type unsubscribesRepo interface {
	Get(models.ConnectionInterface, string, string, string) (bool, error)
}
