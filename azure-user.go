package cloudymsgraph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
)

var InvalidInstanceName = errors.New("invalid instance name")

const Azure = "azure"

const MSGraphVersionV1 = "v1.0"
const MSGraphVersionBeta = "beta"

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

type MSGraphInstance struct {
	Name  string
	Login string
	Base  string
}

var USGovernment = MSGraphInstance{
	Name:  "USGovernment",
	Login: "https://login.microsoftonline.us/",
	Base:  "https://graph.microsoft.us/",
}

var AzurePublic = MSGraphInstance{
	Name:  "Public",
	Login: "https://login.microsoftonline.com/",
	Base:  "https://graph.microsoft.com/",
}

func init() {
	cloudy.UserProviders.Register("azure", func(cfg interface{}) (cloudy.Users, error) {
		azum := &AzureUserManager{}
		err := azum.Configure(cfg)
		return azum, err
	})
}

type AzureUserManager struct {
	Client  *msgraphsdk.GraphServiceClient
	Adapter *msgraphsdk.GraphRequestAdapter
	Cfg     *AzureUserConfig
}

type AzureUserConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	Region       string
	APIBase      string
	Version      string
	SelectFields []string
}

func (azcfg *AzureUserConfig) SetInstanceName(name string) error {
	if strings.EqualFold(name, USGovernment.Name) {
		azcfg.SetInstance(USGovernment)
		return nil
	} else if strings.EqualFold(name, AzurePublic.Name) {
		azfcg.SetInstance(AzurePublic)
		return nil
	}

	return InvalidInstanceName
}

func (azcfg *AzureUserConfig) SetInstance(instance *MSGraphInstance) {
	azcfg.APIBase = instance.Base
	azcfg.Region = instance.Login
}

func (azum *AzureUserManager) Configure(cfg interface{}) error {
	azCfg := cfg.(AzureUserConfig)

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

	azum.Cfg = &azCfg
	azum.Adapter = adapter
	azum.Client = msgraphsdk.NewGraphServiceClient(adapter)
}

func cfgFromMap(cfgMap map[string]interface{}) *AzureUserConfig {
	cfg := &AzureUserConfig{}

	cfg.TenantID, _ = cloudy.MapKeyStr(cfgMap, "TenantID", true)
	cfg.ClientID, _ = cloudy.MapKeyStr(cfgMap, "ClientID", true)
	cfg.ClientSecret, _ = cloudy.MapKeyStr(cfgMap, "ClientSecret", true)
	cfg.Region, _ = cloudy.MapKeyStr(cfgMap, "Region", true)
	cfg.APIBase, _ = cloudy.MapKeyStr(cfgMap, "APIBase", true)

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

func (azum *AzureUserManager) ToAzure(user *cloudymodels.User) *models.User {
	u := models.NewUser()

	//TODO : Finish
	if user.UserName != "" {
		u.UserName = &user.UserName
	}

	return u
}

func (azum *AzureUserManager) ToCloudy(user models.Userable) *cloudymodels.User {
	u := &cloudymodels.User{}

	if user.GetUserPrincipalName() != nil {
		u.UserName = *user.GetUserPrincipalName()
	}

	return u
}

func (azum *AzureUserManager) NewUser(ctx context.Context, newUser *cloudymodels.User) (*cloudymodels.User, error) {
	// upn := "TESTFIRST.TESTLAST@skyborg.onmicrosoft.us"
	// profile := models.NewPasswordProfile()
	// profile.SetForceChangePasswordNextSignIn(boolP(true))
	// profile.SetPassword(stringP("abcde1234$%#^ASDF"))

	// body := models.NewUser()
	// body.SetAccountEnabled(boolP(true))
	// body.SetUserPrincipalName(&upn)
	// body.SetDisplayName(stringP("TESTFIRST TESTLAST"))
	// body.SetGivenName(stringP("TESTFIRST"))
	// body.SetSurname(stringP("TESTLAST"))
	// body.SetCompanyName(stringP("TESTCOMPANY"))
	// body.SetJobTitle(stringP("TESTJOBTITLE"))
	// body.SetMailNickname(stringP("TESTFIRST.TESTLAST"))
	// body.SetPasswordProfile(profile)
	// body.SetBusinessPhones([]string{"+1 937-111-1111"})
	// body.SetMobilePhone(stringP("+1 937-111-1111"))

	body := azum.ToAzure(newUser)

	user, err := graph.Client.Users().Post(body)
	if err != nil {
		return err
	}
	return newUser, nil
}

func (azum *AzureUserManager) GetUser(ctx context.Context, uid string) (*models.User, error) {
	return nil, nil
}

func (azum *AzureUserManager) ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*models.User, interface{}, error) {
	return nil, nil, nil
}

func (azum *AzureUserManager) UpdateUser(ctx context.Context, usr *models.User) (bool, error) {
	return false, nil
}

func (azum *AzureUserManager) Enable(ctx context.Context, uid string) (bool, error) {
	return false, nil
}

func (azum *AzureUserManager) Disable(ctx context.Context, uid string) (bool, error) {
	return false, nil
}

func (azum *AzureUserManager) DeleteUser(ctx context.Context, uid string) (bool, error) {
	return false, nil
}

/// Other thinsgs that MS Graph can do

func (azum *AzureUserManager) AddRemoveLicenses(ctx context.Context, uid string, skusToAdd []string, skusToRemove []string) error {
	return false, nil
}

func (azum *AzureUserManager) SetLicenses(ctx context.Context, uid string, skus []string) error {
	// Get the user

	licenses := user.Get

	return false, nil
}
