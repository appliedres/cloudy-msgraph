package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/appliedres/cloudy"
	mapset "github.com/deckarep/golang-set"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/assignlicense"
)

var GCCHighTeamsPhone = "985fcb26-7b94-475b-b512-89356697be71"
var GCCHighVisioPlan1 = "50a4284d-f0b2-4779-8de6-72c5f9b4349f"
var GCCHighTeamsAudio = "4dee1f32-0808-4fd2-a2ed-fdd575e3a45f"
var GCCHighOffice365E3 = "aea38a85-9bd5-4981-aa00-616b411205bf"
var GCCHighProjectPlan3 = "64758d81-92b7-4855-bcac-06617becb3e8"

func NewGraph(ctx context.Context, tenantID string, clientID string, clientSecret string) (*MSGraph, error) {
	scopes := []string{"https://graph.microsoft.us/.default"}

	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: policy.ClientOptions{
				Cloud: cloud.AzureGovernment,
			},
		})

	if err != nil {
		fmt.Printf("Error authentication provider: %v\n", err)
		return nil, err
	}
	auth, err := a.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		fmt.Printf("Error authentication provider: %v\n", err)
		return nil, err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return nil, err
	}
	adapter.SetBaseUrl("https://graph.microsoft.us/v1.0")

	return &MSGraph{
		Adapter: adapter,
		Client:  msgraphsdk.NewGraphServiceClient(adapter),
	}, nil
}

type MSGraph struct {
	Client  *msgraphsdk.GraphServiceClient
	Adapter *msgraphsdk.GraphRequestAdapter
}

func (graph *MSGraph) DebugSerialize(v serialization.Parsable) {
	writerRegistry := graph.Adapter.GetSerializationWriterFactory() // Returns SerializationWriterFactoryRegistry
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

func (graph *MSGraph) GetUserByID(id string) (models.Userable, error) {
	fields := []string{
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

	requestCfg := &item.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &item.UserItemRequestBuilderGetQueryParameters{
			Select: fields,
		},
	}

	result, err := graph.Client.UsersById("johnbauer@skyborg.onmicrosoft.us").
		GetWithRequestConfigurationAndResponseHandler(requestCfg, nil)

	return result, err
}

func (graph *MSGraph) SetLicenses(ctx context.Context, userId string, licenseSkus []string) error {
	u, err := graph.GetUserByID(ctx, userId)
	if err != nil {
		return err
	}

	var licenses []models.AssignedLicenseable

	var toAdd []string
	var toRemove []string

	assigned := u.GetLicenseAssignmentStates()
	licenseSet := mapset.NewSet[string](licenseSkus...)

	for _, lic := range assigned {
		sku := lic.GetSkuId()
		// This license is meant to be there. Since we
		// can only have each sku once remove it
		if cloudy.ArrayIncludes(licenses, sku) {
			licenseSet.Remove(sku)
		} else {
			toRemove = append(toRemove, sku)
		}
	}
	// Now check the set.. all the remaining need to be added
	for add := range licenseSet {
		al := models.NewAssignedLicense()
		al.SetSkuId(&l)
		licenses = append(licenses, al)
	}

	body := assignlicense.NewAssignLicenseRequestBody()
	body.SetAddLicenses(licenses)
	body.SetRemoveLicenses(toRemove)

	graph.DebugSerialize(body)

	_, err := graph.Client.UsersById(userId).AssignLicense().Post(body)
}

func (graph *MSGraph) AssignLicenses(ctx context.Context, userId string, licenseSkus ...string) error {
	body := assignlicense.NewAssignLicenseRequestBody()

	var licenses []models.AssignedLicenseable
	for _, l := range licenseSkus {
		al := models.NewAssignedLicense()
		al.SetSkuId(&l)
		licenses = append(licenses, al)
	}

	body.SetAddLicenses(licenses)
	body.SetRemoveLicenses([]string{})

	graph.DebugSerialize(body)

	_, err := graph.Client.UsersById(userId).AssignLicense().Post(body)

	return err
}

func (graph *MSGraph) RemoveLicenses(ctx context.Context, userId string, licenseSkus ...string) error {
	body := assignlicense.NewAssignLicenseRequestBody()

	body.SetAddLicenses([]models.AssignedLicenseable{})
	body.SetRemoveLicenses(licenseSkus)
	_, err := graph.Client.UsersById(userId).AssignLicense().Post(body)

	return err
}

func (graph *MSGraph) GetGroups(ctx context.Context, userId string) {

}

func (graph *MSGraph) GetAllGroups(ctx context.Context) {

}

func (graph *MSGraph) GetGroupMembers(ctx context.Context, groupId string) ([]string, error) {
	return nil, nil
}

func (graph *MSGraph) CreateGroup(ctx context.Context, groupId string, groupName string) error {
	return nil
}

func (graph *MSGraph) CreateUser(ctx context.Context, userId string) error {

	upn := "TESTFIRST.TESTLAST@skyborg.onmicrosoft.us"

	profile := models.NewPasswordProfile()
	profile.SetForceChangePasswordNextSignIn(boolP(true))
	profile.SetPassword(stringP("abcde1234$%#^ASDF"))

	body := models.NewUser()
	body.SetAccountEnabled(boolP(true))
	body.SetUserPrincipalName(&upn)
	body.SetDisplayName(stringP("TESTFIRST TESTLAST"))
	body.SetGivenName(stringP("TESTFIRST"))
	body.SetSurname(stringP("TESTLAST"))
	body.SetCompanyName(stringP("TESTCOMPANY"))
	body.SetJobTitle(stringP("TESTJOBTITLE"))
	body.SetMailNickname(stringP("TESTFIRST.TESTLAST"))
	body.SetPasswordProfile(profile)
	body.SetBusinessPhones([]string{"+1 937-111-1111"})
	body.SetMobilePhone(stringP("+1 937-111-1111"))

	user, err := graph.Client.Users().Post(body)
	if err != nil {
		return err
	}

	fmt.Printf("CREATED %v\n", *user.GetUserPrincipalName())
	return nil
}

func stringP(s string) *string {
	return &s
}

func boolP(v bool) *bool {
	return &v
}
