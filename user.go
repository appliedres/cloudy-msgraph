package cloudymsgraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
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

	cloudy.Info(ctx, "MsGraphUserManager NewUser")

	body := UserToAzure(newUser)
	body.SetAccountEnabled(cloudy.BoolP(true))

	user, err := um.Client.Users().Post(ctx, body, nil)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, BadRequest) {
			return nil, cloudy.Error(ctx, "NewUser: %s - BadRequest - %s", newUser.UPN, message)
		} else {
			return nil, cloudy.Error(ctx, "NewUser Error %s %v", message, err)
		}
	}

	created := UserToCloudy(user)
	return created, nil
}

func (um *MsGraphUserManager) GetUser(ctx context.Context, uid string) (*cloudymodels.User, error) {
	cloudy.Info(ctx, "GetUser: %s", uid)
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	fields := DefaultUserSelectFields

	result, err := um.Client.Users().ByUserId(uid).Get(ctx,
		&users.UserItemRequestBuilderGetRequestConfiguration{
			Headers: headers,
			QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
				Select: fields,
			},
		})
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			cloudy.Info(ctx, "GetUser: %s - ResourceNotFound - %s", uid, message)
			return nil, nil
		}

		return nil, cloudy.Error(ctx, "GetUser: %s - error: %v", uid, message)
	}

	return UserToCloudy(result), nil
}

func (um *MsGraphUserManager) GetUserByEmail(ctx context.Context, email string, opts *cloudy.UserOptions) (*cloudymodels.User, error) {

	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	requestFilter := fmt.Sprintf("mail eq '%v'", email)
	count := true
	requestParameters := &users.UsersRequestBuilderGetQueryParameters{
		Filter: &requestFilter,
		Select: append(DefaultUserSelectFields, SigninActivityField),
		Count:  &count,
	}
	configuration := &users.UsersRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	result, err := um.Client.Users().Get(ctx, configuration)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			return nil, cloudy.Error(ctx, "GetUserByEmail Error: %s - ResourceNotFound - %s", email, message)
		}

		return nil, cloudy.Error(ctx, "GetUserByEmail Error: %s %s", email, message)
	}

	var rtn []*cloudymodels.User
	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](result, um.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, err
	}

	err = pageIterator.Iterate(ctx, func(pageItem models.Userable) bool {
		rtn = append(rtn, UserToCloudy(pageItem))
		return true
	})
	if err != nil {
		return nil, err
	}
	if len(rtn) == 0 {
		return nil, nil
	}

	if opts != nil && opts.IncludeLastSignIn != nil && *opts.IncludeLastSignIn {
		//OK ... this is really strange... we need to request "JUST" the "signinactivity" since it fig
	}

	return rtn[0], nil
}

func (um *MsGraphUserManager) ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*cloudymodels.User, interface{}, error) {
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")
	// requestCount := true
	result, err := um.Client.Users().Get(ctx,
		&users.UsersRequestBuilderGetRequestConfiguration{
			Headers: headers,
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				// Count:  &requestCount,
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

	_, err := um.Client.Users().ByUserId(usr.ID).Patch(ctx, azUser, nil)

	if err != nil {
		_, message := GetErrorCodeAndMessage(ctx, err)

		return cloudy.Error(ctx, "UpdateUser Error %s", message)
	}

	return err
}

func (um *MsGraphUserManager) Enable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(true))

	_, err := um.Client.Users().ByUserId(uid).Patch(ctx, u, nil)
	return err
}

func (um *MsGraphUserManager) UploadProfilePicture(ctx context.Context, uid string, picture []byte) error {
	u, err := um.Client.Users().ByUserId(uid).Get(ctx, nil)
	if err != nil {
		return err
	}
	id := *u.GetId()

	_, err = um.Client.Users().ByUserId(id).Photo().Content().Put(ctx, picture, nil)
	return err
}

func (um *MsGraphUserManager) GetProfilePicture(ctx context.Context, uid string) ([]byte, error) {
	cloudy.Info(ctx, "GetProfilePicture for %s", uid)

	u, err := um.Client.Users().ByUserId(uid).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	if u == nil {
		cloudy.Info(ctx, "GetProfilePicture for %s - No user found", uid)
		return nil, nil
	}

	photo, err := um.Client.Users().ByUserId(*u.GetId()).Photo().Content().Get(ctx, nil)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ImageNotFoundCode) {
			cloudy.Warn(ctx, "GetProfilePicture  %s - ImageNotFoundCode - %s", uid, message)
			return nil, nil
		}

		return nil, cloudy.Error(ctx, "GetProfilePicture Error: %s %s", uid, message)
	}

	return photo, err
}

// Associates a certificate ID as a second factor authentication
func (um *MsGraphUserManager) GetCertificateMFA(ctx context.Context, uid string) ([]string, error) {
	azUser, err := um.Client.Users().ByUserId(uid).Get(ctx,
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

	azUser, err := um.Client.Users().ByUserId(uid).Get(ctx,
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

	_, err = um.Client.Users().ByUserId(uid).Patch(ctx, azUser, nil)
	return err
}

func (um *MsGraphUserManager) Disable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(false))
	_, err := um.Client.Users().ByUserId(uid).Patch(ctx, u, nil)
	return err
}

func (um *MsGraphUserManager) DeleteUser(ctx context.Context, uid string) error {
	err := um.Client.Users().ByUserId(uid).Delete(ctx, nil)
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

func (um *MsGraphUserManager) getUserWithCSA(ctx context.Context, uid string) (*cloudymodels.User, error) {
	cloudy.Info(ctx, "GetUser: %s", uid)

	selectFields := append(DefaultUserSelectFields, "customSecurityAttributes")
	headers := abstractions.NewRequestHeaders()

	headers.Add("ConsistencyLevel", "eventual")
	requestParameters := &users.UserItemRequestBuilderGetQueryParameters{
		Select: selectFields,
	}
	configuration := &users.UserItemRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	result, err := um.Client.Users().ByUserId(uid).Get(ctx, configuration)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			return nil, cloudy.Error(ctx, "getUserWithCSA Error: %s - ResourceNotFound - %s", uid, message)
		}

		return nil, cloudy.Error(ctx, "getUserWithCSA Error: %s %s", uid, message)
	}

	return UserToCloudy(result), nil
}
