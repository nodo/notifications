package services_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/notifications/testing/fakes"
	"github.com/cloudfoundry-incubator/notifications/v1/models"
	"github.com/cloudfoundry-incubator/notifications/v1/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Updater", func() {
	Describe("Update", func() {
		var (
			templatesRepo *fakes.TemplatesRepo
			template      models.Template
			updater       services.TemplateUpdater
			database      *fakes.Database
		)

		BeforeEach(func() {
			templatesRepo = fakes.NewTemplatesRepo()
			template = models.Template{
				Name: "gobble template",
				Text: "gobble",
				HTML: "<p>gobble</p>",
			}
			database = fakes.NewDatabase()

			updater = services.NewTemplateUpdater(templatesRepo)
		})

		It("Inserts templates into the templates repo", func() {
			Expect(templatesRepo.Templates).ToNot(ContainElement(template))

			err := updater.Update(database, "my-awesome-id", template)
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesRepo.Templates).To(ContainElement(template))
			Expect(database.ConnectionWasCalled).To(BeTrue())
		})

		It("propagates errors from repo", func() {
			expectedErr := errors.New("Boom!")

			templatesRepo.UpdateError = expectedErr
			err := updater.Update(database, "unimportant", template)

			Expect(err).To(Equal(expectedErr))
		})
	})
})
