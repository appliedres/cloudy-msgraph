package cloudymsgraph

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item"
)

var ErrInvalidInstanceName = errors.New("invalid instance name")

const MSGraphName = "msgraph"

var DefaultUserSelectFields = []string{
	"businessPhones",
	"displayName",
	"givenName",
	"id",
	"jobTitle",
	"mail",
	"mobilePhone",
	"officeLocation",
	"surname",
	"userPrincipalName",
	"assignedLicenses",
	"companyName",
}

func init() {
	cloudy.UserProviders.Register(MSGraphName, &MsGraphUserManagerFactory{})
}

type MsGraphUserManagerFactory struct{}

func (ms *MsGraphUserManagerFactory) Create(cfg interface{}) (cloudy.UserManager, error) {
	azum := &AzureUserManager{}
	err := azum.Configure(cfg)
	return azum, err
}

func (ms *MsGraphUserManagerFactory) ToConfig(config map[string]interface{}) (interface{}, error) {
	cfg := cfgFromMap(config)
	return cfg, nil
}

type AzureUserManager struct {
	Client  *msgraphsdk.GraphServiceClient
	Adapter *msgraphsdk.GraphRequestAdapter
	Cfg     *MSGraphConfig
}

func NewMsGraphUserManager(ctx context.Context, cfg *MSGraphConfig) (*AzureUserManager, error) {
	azum := &AzureUserManager{}
	err := azum.Configure(cfg)

	return azum, err
}

func (azum *AzureUserManager) Configure(cfg interface{}) error {
	azCfg := cfg.(*MSGraphConfig)

	if azCfg.Version == "" {
		azCfg.Version = MSGraphVersionV1
	}
	if azCfg.Region == "" {
		azCfg.SetInstance(&AzurePublic)
	}
	if azCfg.SelectFields == nil {
		azCfg.SelectFields = DefaultUserSelectFields
	}

	scopes := []string{"https://graph.microsoft.us/.default"}

	cred, err := azidentity.NewClientSecretCredential(azCfg.TenantID, azCfg.ClientID, azCfg.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: policy.ClientOptions{
				Cloud: cloud.AzureGovernment,
			},
		})

	if err != nil {
		fmt.Printf("Error authentication provider: %v\n", err)
		return err
	}
	auth, err := a.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		fmt.Printf("Error authentication provider: %v\n", err)
		return err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return err
	}
	adapter.SetBaseUrl("https://graph.microsoft.us/v1.0")

	azum.Cfg = azCfg
	azum.Adapter = adapter
	azum.Client = msgraphsdk.NewGraphServiceClient(adapter)

	return err
}

func cfgFromMap(cfgMap map[string]interface{}) *MSGraphConfig {
	cfg := &MSGraphConfig{}

	cfg.TenantID, _ = cloudy.EnvKeyStr(cfgMap, "TenantID")
	cfg.ClientID, _ = cloudy.EnvKeyStr(cfgMap, "ClientID")
	cfg.ClientSecret, _ = cloudy.EnvKeyStr(cfgMap, "ClientSecret")
	cfg.Region, _ = cloudy.EnvKeyStr(cfgMap, "Region")
	cfg.APIBase, _ = cloudy.EnvKeyStr(cfgMap, "APIBase")

	return cfg
}

func (azum *AzureUserManager) DebugSerialize(v serialization.Parsable) {
	writerRegistry := azum.Adapter.GetSerializationWriterFactory() // Returns SerializationWriterFactoryRegistry
	myWriter, err := writerRegistry.GetSerializationWriter("application/json")
	if err != nil {
		fmt.Printf("Unable to retrieve writer: %v\n", err)
	}
	// err2 := body.Serialize(myWriter)
	err2 := myWriter.WriteObjectValue("", v)

	if err2 != nil {
		fmt.Printf("Unable to serialize: %v\n", err2)
	}
	bodyBytes, _ := myWriter.GetSerializedContent() // ([]byte, error)
	fmt.Println(string(bodyBytes))
}

func UserToAzure(user *cloudymodels.User) *models.User {
	u := models.NewUser()

	u.SetUserPrincipalName(&user.UserName)
	u.SetDisplayName(&user.DisplayName)
	emailNickname := cloudy.TrimDomain(*u.GetUserPrincipalName())

	u.SetGivenName(&user.FirstName)
	u.SetSurname(&user.LastName)
	if user.Company != "" {
		u.SetCompanyName(&user.Company)
	}
	if user.JobTitle != "" {
		u.SetJobTitle(&user.JobTitle)
	}
	u.SetMailNickname(&emailNickname)
	if user.OfficePhone != "" {
		u.SetBusinessPhones([]string{user.OfficePhone})
	}
	if user.MobilePhone != "" {
		u.SetMobilePhone(&user.MobilePhone)
	}
	if user.Department != "" {
		u.SetDepartment(&user.Department)
	}
	if user.MustChangePassword || user.Password != "" {
		profile := models.NewPasswordProfile()
		profile.SetForceChangePasswordNextSignIn(cloudy.BoolP(user.MustChangePassword))
		profile.SetPassword(&user.Password)
		u.SetPasswordProfile(profile)
	}

	return u
}

func UserToCloudy(user models.Userable) *cloudymodels.User {
	u := &cloudymodels.User{}

	u.ID = *user.GetId()
	if user.GetUserPrincipalName() != nil {
		u.UserName = *user.GetUserPrincipalName()
	}
	if user.GetGivenName() != nil {
		u.FirstName = *user.GetGivenName()
	}
	if user.GetSurname() != nil {
		u.LastName = *user.GetSurname()
	}
	if user.GetCompanyName() != nil {
		u.Company = *user.GetCompanyName()
	}
	if user.GetJobTitle() != nil {
		u.JobTitle = *user.GetJobTitle()
	}
	if user.GetDisplayName() != nil {
		u.DisplayName = *user.GetDisplayName()
	}
	if user.GetDepartment() != nil {
		u.Department = *user.GetDepartment()
	}
	if user.GetMobilePhone() != nil {
		u.MobilePhone = *user.GetMobilePhone()
	}

	if len(user.GetBusinessPhones()) == 1 {
		u.OfficePhone = user.GetBusinessPhones()[0]
	}

	return u
}

func (azum *AzureUserManager) NewUser(ctx context.Context, newUser *cloudymodels.User) (*cloudymodels.User, error) {

	body := UserToAzure(newUser)
	body.SetAccountEnabled(cloudy.BoolP(true))

	user, err := azum.Client.Users().Post(body)
	if err != nil {
		return nil, err
	}

	created := UserToCloudy(user)
	return created, nil
}

func (azum *AzureUserManager) GetUser(ctx context.Context, uid string) (*cloudymodels.User, error) {
	fields := DefaultUserSelectFields

	requestCfg := &item.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &item.UserItemRequestBuilderGetQueryParameters{
			Select: fields,
		},
	}

	result, err := azum.Client.UsersById(uid).GetWithRequestConfigurationAndResponseHandler(requestCfg, nil)
	if err != nil {
		oerr := err.(*odataerrors.ODataError)
		code := *oerr.GetError().GetCode()

		if code == "Request_ResourceNotFound" {
			return nil, nil
		}

		return nil, err
	}
	return UserToCloudy(result), nil
}

func (azum *AzureUserManager) ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*cloudymodels.User, interface{}, error) {
	fields := DefaultUserSelectFields

	result, err := azum.Client.Users().GetWithRequestConfigurationAndResponseHandler(
		&users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Select: fields,
			},
		}, nil)
	if err != nil {
		return nil, nil, err
	}

	var rtn []*cloudymodels.User
	pageIterator, err := msgraphcore.NewPageIterator(result, azum.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, nil, err
	}

	err = pageIterator.Iterate(func(pageItem interface{}) bool {
		u := pageItem.(models.Userable)
		rtn = append(rtn, UserToCloudy(u))
		return true
	})
	if err != nil {
		return nil, nil, err
	}

	// items := result.GetValue()
	// rtn := make([]*cloudymodels.User, len(items))
	// for i, u := range items {
	// 	rtn[i] = UserToCloudy(u)
	// }

	return rtn, nil, nil
}

func (azum *AzureUserManager) UpdateUser(ctx context.Context, usr *cloudymodels.User) error {
	azUser := UserToAzure(usr)

	_, err := azum.Client.Users().Post(azUser)
	return err
}

func (azum *AzureUserManager) Enable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(true))

	err := azum.Client.UsersById(uid).Patch(u)
	return err
}

func (azum *AzureUserManager) Disable(ctx context.Context, uid string) error {
	u := models.NewUser()
	u.SetAccountEnabled(cloudy.BoolP(false))
	err := azum.Client.UsersById(uid).Patch(u)
	return err
}

func (azum *AzureUserManager) DeleteUser(ctx context.Context, uid string) error {
	err := azum.Client.UsersById(uid).Delete()
	return err
}

func (azum *AzureUserManager) ForceUserName(ctx context.Context, name string) (string, bool, error) {
	u, err := azum.GetUser(ctx, name)
	if err != nil {
		return name, false, err
	}

	if u != nil {
		return name, true, nil
	}

	return name, false, nil
}
