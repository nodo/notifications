package handlers_test

import (
    "errors"
    "testing"

    "github.com/cloudfoundry-incubator/notifications/cf"
    "github.com/cloudfoundry-incubator/notifications/config"
    "github.com/cloudfoundry-incubator/notifications/mail"
    "github.com/dgrijalva/jwt-go"
    "github.com/pivotal-cf/uaa-sso-golang/uaa"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

func TestWebHandlersSuite(t *testing.T) {
    RegisterFastTokenSigningMethod()

    RegisterFailHandler(Fail)
    RunSpecs(t, "Web Handlers Suite")
}

const (
    UAAPrivateKey = "PRIVATE-KEY"
    UAAPublicKey  = "PUBLIC-KEY"
)

type SigningMethodFast struct{}

func (m SigningMethodFast) Alg() string {
    return "FAST"
}

func (m SigningMethodFast) Sign(signingString string, key []byte) (string, error) {
    signature := jwt.EncodeSegment([]byte(signingString + "SUPERFAST"))
    return signature, nil
}

func (m SigningMethodFast) Verify(signingString, signature string, key []byte) (err error) {
    if signature != jwt.EncodeSegment([]byte(signingString+"SUPERFAST")) {
        return errors.New("Signature is invalid")
    }

    return nil
}

func RegisterFastTokenSigningMethod() {
    jwt.RegisterSigningMethod("FAST", func() jwt.SigningMethod {
        return SigningMethodFast{}
    })
}

func BuildToken(header map[string]interface{}, claims map[string]interface{}) string {
    config.UAAPublicKey = UAAPublicKey

    alg := header["alg"].(string)
    signingMethod := jwt.GetSigningMethod(alg)
    token := jwt.New(signingMethod)
    token.Header = header
    token.Claims = claims

    signed, err := token.SignedString([]byte(UAAPrivateKey))
    if err != nil {
        panic(err)
    }

    return signed
}

type FakeMailClient struct {
    messages       []mail.Message
    errorOnSend    bool
    errorOnConnect bool
}

func (fake *FakeMailClient) Connect() error {
    if fake.errorOnConnect {
        return errors.New("BOOM!")
    }
    return nil
}

func (fake *FakeMailClient) Send(msg mail.Message) error {
    err := fake.Connect()
    if err != nil {
        return err
    }

    if fake.errorOnSend {
        return errors.New("BOOM!")
    }

    fake.messages = append(fake.messages, msg)
    return nil
}

type FakeUAAClient struct {
    ClientToken      uaa.Token
    UsersByID        map[string]uaa.User
    ErrorForUserByID error
}

func (fake FakeUAAClient) AuthorizeURL() string {
    return ""
}

func (fake FakeUAAClient) LoginURL() string {
    return ""
}

func (fake FakeUAAClient) SetToken(token string) {}

func (fake FakeUAAClient) Exchange(code string) (uaa.Token, error) {
    return uaa.Token{}, nil
}

func (fake FakeUAAClient) Refresh(token string) (uaa.Token, error) {
    return uaa.Token{}, nil
}

func (fake FakeUAAClient) GetClientToken() (uaa.Token, error) {
    return fake.ClientToken, nil
}

func (fake FakeUAAClient) GetTokenKey() (string, error) {
    return "", nil
}

func (fake FakeUAAClient) UserByID(id string) (uaa.User, error) {
    return fake.UsersByID[id], fake.ErrorForUserByID
}

type FakeCloudController struct {
    UsersBySpaceGuid         map[string][]cf.CloudControllerUser
    CurrentToken             string
    GetUsersBySpaceGuidError error
    Spaces                   map[string]cf.CloudControllerSpace
    Orgs                     map[string]cf.CloudControllerOrganization
}

func NewFakeCloudController() *FakeCloudController {
    return &FakeCloudController{
        UsersBySpaceGuid: make(map[string][]cf.CloudControllerUser),
    }
}

func (fake *FakeCloudController) GetUsersBySpaceGuid(guid, token string) ([]cf.CloudControllerUser, error) {
    fake.CurrentToken = token

    if users, ok := fake.UsersBySpaceGuid[guid]; ok {
        return users, fake.GetUsersBySpaceGuidError
    } else {
        return make([]cf.CloudControllerUser, 0), fake.GetUsersBySpaceGuidError
    }
}

func (fake *FakeCloudController) LoadSpace(guid, token string) (cf.CloudControllerSpace, error) {
    if space, ok := fake.Spaces[guid]; ok {
        return space, nil
    } else {
        return cf.CloudControllerSpace{}, nil
    }
}

func (fake *FakeCloudController) LoadOrganization(guid, token string) (cf.CloudControllerOrganization, error) {
    if org, ok := fake.Orgs[guid]; ok {
        return org, nil
    } else {
        return cf.CloudControllerOrganization{}, nil
    }
}
