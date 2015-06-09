package strategies

import (
	"time"

	"github.com/cloudfoundry-incubator/notifications/cf"
	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/cloudfoundry-incubator/notifications/postal"
)

const EmailEndorsement = "This message was sent directly to your email address."

type EmailStrategy struct {
	mailer MailerInterface
}

func NewEmailStrategy(mailer MailerInterface) EmailStrategy {
	return EmailStrategy{
		mailer: mailer,
	}
}

func (strategy EmailStrategy) Dispatch(clientID, guid, vcapRequestID string, requestReceived time.Time, options postal.Options, conn models.ConnectionInterface) ([]Response, error) {
	options.Endorsement = EmailEndorsement
	responses := strategy.mailer.Deliver(conn, []User{{Email: options.To}}, options, cf.CloudControllerSpace{}, cf.CloudControllerOrganization{}, clientID, "", vcapRequestID, requestReceived)

	return responses, nil
}
