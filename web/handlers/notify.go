package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/cloudfoundry-incubator/notifications/postal"
	"github.com/cloudfoundry-incubator/notifications/postal/strategies"
	"github.com/cloudfoundry-incubator/notifications/web/params"
	"github.com/cloudfoundry-incubator/notifications/web/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/ryanmoran/stack"
)

type NotifyInterface interface {
	Execute(models.ConnectionInterface, *http.Request, stack.Context, string, strategies.StrategyInterface, ValidatorInterface, string) ([]byte, error)
}

type Notify struct {
	finder    services.NotificationsFinderInterface
	registrar services.RegistrarInterface
}

func NewNotify(finder services.NotificationsFinderInterface, registrar services.RegistrarInterface) Notify {
	return Notify{
		finder:    finder,
		registrar: registrar,
	}
}

type ValidatorInterface interface {
	Validate(*params.Notify) bool
}

func (handler Notify) Execute(connection models.ConnectionInterface, req *http.Request, context stack.Context,
	guid string, strategy strategies.StrategyInterface, validator ValidatorInterface, vcapRequestID string) ([]byte, error) {
	parameters, err := params.NewNotify(req.Body)
	if err != nil {
		return []byte{}, err
	}

	if !validator.Validate(&parameters) {
		return []byte{}, params.ValidationError(parameters.Errors)
	}

	requestReceivedTime, ok := context.Get(RequestReceivedTime).(time.Time)
	if !ok {
		panic("programmer error: missing handlers.RequestReceivedTime in http context")
	}
	token := context.Get("token").(*jwt.Token) // TODO: (rm) get rid of the context object, just pass in the token
	clientID := token.Claims["client_id"].(string)

	client, kind, err := handler.finder.ClientAndKind(context.Get("database").(models.DatabaseInterface), clientID, parameters.KindID)
	if err != nil {
		return []byte{}, err
	}

	if kind.Critical && !handler.hasCriticalNotificationsWriteScope(token.Claims["scope"]) {
		return []byte{}, postal.NewCriticalNotificationError(kind.ID)
	}

	err = handler.registrar.Register(connection, client, []models.Kind{kind})
	if err != nil {
		return []byte{}, err
	}

	var responses []strategies.Response

	responses, err = strategy.Dispatch(clientID, guid, vcapRequestID, requestReceivedTime, parameters.ToOptions(client, kind), connection)
	if err != nil {
		return []byte{}, err
	}

	output, err := json.Marshal(responses)
	if err != nil {
		panic(err)
	}

	return output, nil
}

func (handler Notify) hasCriticalNotificationsWriteScope(elements interface{}) bool {
	for _, elem := range elements.([]interface{}) {
		if elem.(string) == "critical_notifications.write" {
			return true
		}
	}
	return false
}
