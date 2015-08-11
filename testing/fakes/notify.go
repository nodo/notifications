package fakes

import (
	"net/http"

	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/v1/services"
	"github.com/cloudfoundry-incubator/notifications/v1/web/notify"
	"github.com/ryanmoran/stack"
)

type Notify struct {
	ExecuteCall struct {
		Receives struct {
			Connection    db.ConnectionInterface
			Request       *http.Request
			Context       stack.Context
			GUID          string
			Strategy      services.StrategyInterface
			Validator     notify.ValidatorInterface
			VCAPRequestID string
		}
		Returns struct {
			Response []byte
			Error    error
		}
	}
}

func NewNotify() *Notify {
	return &Notify{}
}

func (n *Notify) Execute(connection db.ConnectionInterface, req *http.Request, context stack.Context,
	guid string, strategy services.StrategyInterface, validator notify.ValidatorInterface, vcapRequestID string) ([]byte, error) {

	n.ExecuteCall.Receives.Connection = connection
	n.ExecuteCall.Receives.Request = req
	n.ExecuteCall.Receives.Context = context
	n.ExecuteCall.Receives.GUID = guid
	n.ExecuteCall.Receives.Strategy = strategy
	n.ExecuteCall.Receives.Validator = validator
	n.ExecuteCall.Receives.VCAPRequestID = vcapRequestID

	return n.ExecuteCall.Returns.Response, n.ExecuteCall.Returns.Error
}
