package notifications

import (
	"io"

	"github.com/cloudfoundry-incubator/notifications/v1/models"
	"github.com/cloudfoundry-incubator/notifications/valiant"
	"github.com/cloudfoundry-incubator/notifications/web/webutil"
)

type NotificationUpdateParams struct {
	Description string `json:"description" validate-required:"true"`
	Critical    bool   `json:"critical"    validate-required:"true"`
	TemplateID  string `json:"template"    validate-required:"true"`
}

func NewNotificationParams(body io.Reader) (NotificationUpdateParams, error) {
	var params NotificationUpdateParams

	validator := valiant.NewValidator(body)
	err := validator.Validate(&params)
	if err != nil {
		switch err.(type) {
		case valiant.RequiredFieldError:
			return params, webutil.ValidationError([]string{err.Error()})
		default:
			return params, webutil.ParseError{}
		}
	}
	return params, nil
}

func (params NotificationUpdateParams) ToModel(clientID, notificationID string) models.Kind {
	return models.Kind{
		Description: params.Description,
		Critical:    params.Critical,
		TemplateID:  params.TemplateID,
		ClientID:    clientID,
		ID:          notificationID,
	}
}
