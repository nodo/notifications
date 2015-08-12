package services

import (
	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/v1/models"
)

type TemplateUpdaterInterface interface {
	Update(db.DatabaseInterface, string, models.Template) error
}

type TemplateUpdater struct {
	templatesRepo TemplatesRepo
}

func NewTemplateUpdater(templatesRepo TemplatesRepo) TemplateUpdater {
	return TemplateUpdater{
		templatesRepo: templatesRepo,
	}
}

func (updater TemplateUpdater) Update(database db.DatabaseInterface, templateID string, template models.Template) error {
	_, err := updater.templatesRepo.Update(database.Connection(), templateID, template)
	if err != nil {
		return err
	}
	return nil
}
