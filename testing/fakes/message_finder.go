package fakes

import (
	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/v1/services"
)

type MessageFinder struct {
	FindCall struct {
		Receives struct {
			Database  db.DatabaseInterface
			MessageID string
		}
		Returns struct {
			Message services.Message
			Error   error
		}
	}
}

func NewMessageFinder() *MessageFinder {
	return &MessageFinder{}
}

func (f *MessageFinder) Find(database db.DatabaseInterface, messageID string) (services.Message, error) {
	f.FindCall.Receives.Database = database
	f.FindCall.Receives.MessageID = messageID

	return f.FindCall.Returns.Message, f.FindCall.Returns.Error
}
