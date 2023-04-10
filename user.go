package cloudymsgraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

func init() {
	cloudy.UserProviders.Register(MsGraphName, &MsGraphUserManagerFactory{})
}

type MsGraphUserManagerFactory struct {
	MsGraph
}

func (umf *MsGraphUserManagerFactory) Create(cfg interface{}) (cloudy.UserManager, error) {
	return NewMsGraphUserManager(context.Background(), cfg.(*MsGraphConfig))
}

func (umf *MsGraphUserManagerFactory) FromEnv(env *cloudy.Environment) (interface{}, error) {
	cfg := fromEnvironment(env)
	return cfg, nil
}

type MsGraphUserManager struct {
	*MsGraph
}

func NewMsGraphUserManager(ctx context.Context, cfg *MsGraphConfig) (*MsGraphUserManager, error) {
	um := &MsGraphUserManager{
		MsGraph: &MsGraph{},
	}
	err := um.Configure(cfg)

	return um, err
}

func NewMsGraphUserManagerFromEnv(ctx context.Context, env *cloudy.Environment) (*MsGraphUserManager, error) {
	fact := &MsGraphUserManagerFactory{}
	cfg, _ := fact.FromEnv(env)
	return NewMsGraphUserManager(ctx, cfg.(*MsGraphConfig))
}

func fromEnvironment(env *cloudy.Environment) *MsGraphConfig {
	creds := env.GetCredential(MSGraphCredentialsKey)
	if creds != nil {
		return creds.(*MsGraphConfig)
	}

	cfg := &MsGraphConfig{}

	cfg.TenantID = env.Force("AZ_TENANT_ID")
	cfg.ClientID = env.Force("AZ_CLIENT_ID")
	cfg.ClientSecret = env.Force("AZ_CLIENT_SECRET")
	cfg.Region = env.Default("AZ_REGION", "usgovvirginia")
	cfg.APIBase = env.Default("AZ_API_BASE", "https://graph.microsoft.us/v1.0")

	return cfg
}

func (um *MsGraphUserManager) NewUser(ctx context.Context, newUser *cloudymodels.User) (*cloudymodels.User, error) {

	body := UserToAzure(newUser)
	body.SetAccountEnabled(cloudy.BoolP(true))

	user, err := um.Client.Users().Post(ctx, body, nil)
	if err != nil {
		return nil, err
	}

	created := UserToCloudy(user)
	return created, nil
}

func (um *MsGraphUserManager) GetUser(ctx context.Context, uid string) (*cloudymodels.User, error) {
	cloudy.Info(ctx, "GetUser: %s", uid)

	fields := DefaultUserSelectFields

	result, err := um.Client.UsersById(uid).Get(ctx,
		&users.UserItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
				Select: fields,
			},
		})
	if err != nil {

		oerr := err.(*odataerrors.ODataError)
		code := *oerr.GetError().GetCode()
		message := *oerr.GetError().GetMessage()

		if code == "Request_ResourceNotFound" {
			cloudy.Info(ctx, "GetUser: %s - Request_ResourceNotFound - %s", uid, message)
			return nil, nil
		}

		return nil, cloudy.Error(ctx, "GetUser: %s - error: %v", uid, message)
	}
	return UserToCloudy(result), nil
}

func (um *MsGraphUserManager) ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*cloudymodels.User, interface{}, error) {
	result, err := um.Client.Users().Get(ctx,
		&users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Select: DefaultUserSelectFields,
			},
		})
	if err != nil {
		return nil, nil, err
	}

	var rtn []*cloudymodels.User
	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](result, um.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, nil, err
	}

	err = pageIterator.Iterate(ctx, func(pageItem models.Userable) bool {
		rtn = append(rtn, UserToCloudy(pageItem))
		return true
	})
	if err != nil {
		return nil, nil, err
	}

	return rtn, nil, nil
}

func (um *MsGraphUserManager) UpdateUser(ctx context.Context, usr *cloudymodels.User) error {
	azUser := UserToAzure(usr)

	_, err := um.Client.UsersById(usr.ID).Patch(ctx, azUser, nil)
	return err
}

func (um *MsGraphUserManager) Enable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(true))

	_, err := um.Client.UsersById(uid).Patch(ctx, u, nil)
	return err
}

func (um *MsGraphUserManager) UploadProfilePicture(ctx context.Context, uid string, picture []byte) error {
	u, err := um.Client.UsersById(uid).Get(ctx, nil)
	if err != nil {
		return err
	}
	id := *u.GetId()

	_, err = um.Client.UsersById(id).Photo().Content().Put(ctx, picture, nil)
	return err
}

func (um *MsGraphUserManager) GetProfilePicture(ctx context.Context, uid string) ([]byte, error) {

	u, err := um.Client.UsersById(uid).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	id := *u.GetId()

	photo, err := um.Client.UsersById(id).Photo().Content().Get(ctx, nil)
	return photo, err
}

// Associates a certificate ID as a second factor authentication
func (um *MsGraphUserManager) GetCertificateMFA(ctx context.Context, uid string) ([]string, error) {
	azUser, err := um.Client.UsersById(uid).Get(ctx,
		&users.UserItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
				Select: DefaultUserSelectFields,
			},
		})

	if err != nil {
		return nil, err
	}

	info := azUser.GetAuthorizationInfo()
	certIds := info.GetCertificateUserIds()

	return certIds, nil
}

// Associates a certificate ID as a second factor authentication
func (um *MsGraphUserManager) AssocateCerificateMFA(ctx context.Context, uid string, certId string, replace bool) error {
	if !strings.HasPrefix(certId, "X509:<PN>") {
		certId = fmt.Sprintf("X509:<PN>%v", certId)
	}

	azUser, err := um.Client.UsersById(uid).Get(ctx,
		&users.UserItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
				Select: DefaultUserSelectFields,
			},
		})

	if err != nil {
		return err
	}

	info := azUser.GetAuthorizationInfo()
	certIds := info.GetCertificateUserIds()

	var newCertIds []string
	if replace {
		newCertIds = append(newCertIds, certId)
	} else {
		newCertIds = append(certIds, certId)
	}

	info.SetCertificateUserIds(newCertIds)

	_, err = um.Client.UsersById(uid).Patch(ctx, azUser, nil)
	return err
}

func (um *MsGraphUserManager) Disable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(false))
	_, err := um.Client.UsersById(uid).Patch(ctx, u, nil)
	return err
}

func (um *MsGraphUserManager) DeleteUser(ctx context.Context, uid string) error {
	err := um.Client.UsersById(uid).Delete(ctx, nil)
	return err
}

func (um *MsGraphUserManager) ForceUserName(ctx context.Context, name string) (string, bool, error) {
	u, err := um.GetUser(ctx, name)
	if err != nil {
		return name, false, err
	}

	if u != nil {
		return name, true, nil
	}

	return name, false, nil
}
