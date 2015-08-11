package templates

import (
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/v1/services"
	"github.com/ryanmoran/stack"
)

type UpdateHandler struct {
	updater     services.TemplateUpdaterInterface
	errorWriter errorWriter
}

func NewUpdateHandler(updater services.TemplateUpdaterInterface, errWriter errorWriter) UpdateHandler {
	return UpdateHandler{
		updater:     updater,
		errorWriter: errWriter,
	}
}

func (h UpdateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, context stack.Context) {
	templateID := strings.Split(req.URL.String(), "/templates/")[1]

	templateParams, err := NewTemplateParams(req.Body)
	if err != nil {
		h.errorWriter.Write(w, err)
		return
	}

	err = h.updater.Update(context.Get("database").(db.DatabaseInterface), templateID, templateParams.ToModel())
	if err != nil {
		h.errorWriter.Write(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
