package services

import "github.com/cloudfoundry-incubator/notifications/models"

type clientsRepo interface {
	FindAll(models.ConnectionInterface) ([]models.Client, error)
	Find(models.ConnectionInterface, string) (models.Client, error)
	Upsert(models.ConnectionInterface, models.Client) (models.Client, error)
	Update(models.ConnectionInterface, models.Client) (models.Client, error)
	FindAllByTemplateID(models.ConnectionInterface, string) ([]models.Client, error)
}

type kindsRepo interface {
	FindAll(models.ConnectionInterface) ([]models.Kind, error)
	Find(models.ConnectionInterface, string, string) (models.Kind, error)
	Update(models.ConnectionInterface, models.Kind) (models.Kind, error)
	Upsert(models.ConnectionInterface, models.Kind) (models.Kind, error)
	FindAllByTemplateID(models.ConnectionInterface, string) ([]models.Kind, error)
	Trim(models.ConnectionInterface, string, []string) (int, error)
}

type preferencesRepo interface {
	FindNonCriticalPreferences(models.ConnectionInterface, string) ([]models.Preference, error)
}

type templatesRepo interface {
	FindByID(models.ConnectionInterface, string) (models.Template, error)
	Create(models.ConnectionInterface, models.Template) (models.Template, error)
	Destroy(models.ConnectionInterface, string) error
	ListIDsAndNames(models.ConnectionInterface) ([]models.Template, error)
	Update(models.ConnectionInterface, string, models.Template) (models.Template, error)
}