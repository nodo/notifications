package application

import (
	"crypto/rand"
	"database/sql"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/notifications/cf"
	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/gobble"
	"github.com/cloudfoundry-incubator/notifications/mail"
	"github.com/cloudfoundry-incubator/notifications/uaa"
	v1models "github.com/cloudfoundry-incubator/notifications/v1/models"
	"github.com/cloudfoundry-incubator/notifications/v1/services"
	v2models "github.com/cloudfoundry-incubator/notifications/v2/models"
	"github.com/cloudfoundry-incubator/notifications/v2/queue"
	"github.com/cloudfoundry-incubator/notifications/v2/util"
	"github.com/pivotal-golang/lager"
)

type Mother struct {
	sqlDB *sql.DB
	mutex sync.Mutex
}

func NewMother() *Mother {
	return &Mother{}
}

func (m *Mother) GobbleDatabase() gobble.DatabaseInterface {
	return gobble.NewDatabase(m.SQLDatabase())
}

func (m *Mother) Queue() gobble.QueueInterface {
	env := NewEnvironment()

	return gobble.NewQueue(m.GobbleDatabase(), gobble.Config{
		WaitMaxDuration: time.Duration(env.GobbleWaitMaxDuration) * time.Millisecond,
	})
}

func (m *Mother) V2Enqueuer() queue.JobEnqueuer {
	return queue.NewJobEnqueuer(m.Queue(), v2models.NewMessagesRepository(util.NewClock(), v2models.NewGUIDGenerator(rand.Reader).Generate))
}

func (m *Mother) UserStrategy() services.UserStrategy {
	return services.NewUserStrategy(m.Enqueuer(), m.V2Enqueuer())
}

func (m *Mother) SpaceStrategy() services.SpaceStrategy {
	env := NewEnvironment()
	uaaClient := uaa.NewZonedUAAClient(env.UAAClientID, env.UAAClientSecret, env.VerifySSL, UAAPublicKey)
	cloudController := cf.NewCloudController(env.CCHost, !env.VerifySSL)

	tokenLoader := uaa.NewTokenLoader(uaaClient)
	spaceLoader := services.NewSpaceLoader(cloudController)
	organizationLoader := services.NewOrganizationLoader(cloudController)
	enqueuer := m.Enqueuer()
	findsUserGUIDs := services.NewFindsUserGUIDs(cloudController, uaaClient)

	return services.NewSpaceStrategy(tokenLoader, spaceLoader, organizationLoader, findsUserGUIDs, enqueuer, m.V2Enqueuer())
}

func (m *Mother) OrganizationStrategy() services.OrganizationStrategy {
	env := NewEnvironment()
	cloudController := cf.NewCloudController(env.CCHost, !env.VerifySSL)

	uaaClient := uaa.NewZonedUAAClient(env.UAAClientID, env.UAAClientSecret, env.VerifySSL, UAAPublicKey)
	tokenLoader := uaa.NewTokenLoader(uaaClient)
	organizationLoader := services.NewOrganizationLoader(cloudController)
	findsUserGUIDs := services.NewFindsUserGUIDs(cloudController, uaaClient)
	enqueuer := m.Enqueuer()

	return services.NewOrganizationStrategy(tokenLoader, organizationLoader, findsUserGUIDs, enqueuer, m.V2Enqueuer())
}

func (m *Mother) EveryoneStrategy() services.EveryoneStrategy {
	env := NewEnvironment()
	uaaClient := uaa.NewZonedUAAClient(env.UAAClientID, env.UAAClientSecret, env.VerifySSL, UAAPublicKey)
	tokenLoader := uaa.NewTokenLoader(uaaClient)
	allUsers := services.NewAllUsers(uaaClient)
	enqueuer := m.Enqueuer()

	return services.NewEveryoneStrategy(tokenLoader, allUsers, enqueuer, m.V2Enqueuer())
}

func (m *Mother) UAAScopeStrategy() services.UAAScopeStrategy {
	env := NewEnvironment()
	uaaClient := uaa.NewZonedUAAClient(env.UAAClientID, env.UAAClientSecret, env.VerifySSL, UAAPublicKey)
	cloudController := cf.NewCloudController(env.CCHost, !env.VerifySSL)

	tokenLoader := uaa.NewTokenLoader(uaaClient)
	findsUserGUIDs := services.NewFindsUserGUIDs(cloudController, uaaClient)
	enqueuer := m.Enqueuer()

	return services.NewUAAScopeStrategy(tokenLoader, findsUserGUIDs, enqueuer, m.V2Enqueuer(), env.DefaultUAAScopes)
}

func (m *Mother) EmailStrategy() services.EmailStrategy {
	return services.NewEmailStrategy(m.Enqueuer(), m.V2Enqueuer())
}

func (m *Mother) Enqueuer() services.Enqueuer {
	return services.NewEnqueuer(m.Queue(), m.MessagesRepo())
}

func (m *Mother) MailClient() *mail.Client {
	env := NewEnvironment()
	mailConfig := mail.Config{
		User:           env.SMTPUser,
		Pass:           env.SMTPPass,
		Host:           env.SMTPHost,
		Port:           env.SMTPPort,
		Secret:         env.SMTPCRAMMD5Secret,
		TestMode:       env.TestMode,
		SkipVerifySSL:  !env.VerifySSL,
		DisableTLS:     !env.SMTPTLS,
		LoggingEnabled: env.SMTPLoggingEnabled,
	}

	switch env.SMTPAuthMechanism {
	case SMTPAuthNone:
		mailConfig.AuthMechanism = mail.AuthNone
	case SMTPAuthPlain:
		mailConfig.AuthMechanism = mail.AuthPlain
	case SMTPAuthCRAMMD5:
		mailConfig.AuthMechanism = mail.AuthCRAMMD5
	}

	return mail.NewClient(mailConfig)
}

func (m *Mother) Logger() lager.Logger {
	logger := lager.NewLogger("notifications")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	return logger
}

func (m *Mother) SQLDatabase() *sql.DB {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.sqlDB != nil {
		return m.sqlDB
	}

	env := NewEnvironment()

	var err error
	m.sqlDB, err = sql.Open("mysql", env.DatabaseURL)
	if err != nil {
		panic(err)
	}

	if err := m.sqlDB.Ping(); err != nil {
		panic(err)
	}

	m.sqlDB.SetMaxOpenConns(env.DBMaxOpenConns)

	return m.sqlDB
}

func (m *Mother) Database() db.DatabaseInterface {
	env := NewEnvironment()
	database := v1models.NewDatabase(m.SQLDatabase(), v1models.Config{
		DefaultTemplatePath: path.Join(env.RootPath, "templates", "default.json"),
	})

	if env.DBLoggingEnabled {
		database.TraceOn("[DB]", log.New(os.Stdout, "", 0))
	}

	return database
}

func (m *Mother) KindsRepo() v1models.KindsRepo {
	return v1models.NewKindsRepo()
}

func (m *Mother) UnsubscribesRepo() v1models.UnsubscribesRepo {
	return v1models.NewUnsubscribesRepo()
}

func (m *Mother) GlobalUnsubscribesRepo() v1models.GlobalUnsubscribesRepo {
	return v1models.NewGlobalUnsubscribesRepo()
}

func (m *Mother) MessagesRepo() v1models.MessagesRepo {
	return v1models.NewMessagesRepo(v2models.NewGUIDGenerator(rand.Reader).Generate)
}

func (m *Mother) ReceiptsRepo() v1models.ReceiptsRepo {
	return v1models.NewReceiptsRepo()
}
