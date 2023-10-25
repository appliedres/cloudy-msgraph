package cloudymsgraph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/appliedres/cloudy"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	msauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

const MsGraphName = "msgraph"

var ErrInvalidInstanceName = errors.New("invalid instance name")

type MsGraphInstance struct {
	Name  string
	Login string
	Base  string
}

var USGovernment = MsGraphInstance{
	Name:  "USGovernment",
	Login: "https://login.microsoftonline.us/",
	Base:  "https://graph.microsoft.us/",
}

var AzurePublic = MsGraphInstance{
	Name:  "Public",
	Login: "https://login.microsoftonline.com/",
	Base:  "https://graph.microsoft.com/",
}

type MsGraphConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	Region       string
	APIBase      string
	SelectFields []string
}

func (azConfig *MsGraphConfig) SetInstanceName(name string) error {
	if strings.EqualFold(name, USGovernment.Name) {
		azConfig.SetInstance(&USGovernment)
		return nil
	} else if strings.EqualFold(name, AzurePublic.Name) {
		azConfig.SetInstance(&AzurePublic)
		return nil
	}

	return ErrInvalidInstanceName
}

func (azConfig *MsGraphConfig) SetInstance(instance *MsGraphInstance) {
	azConfig.APIBase = instance.Base
	azConfig.Region = instance.Login
}

type MsGraph struct {
	Client  *msgraphsdk.GraphServiceClient
	Adapter *msgraphsdk.GraphRequestAdapter
	Cfg     *MsGraphConfig
}

func (azUM *MsGraph) Configure(azConfig *MsGraphConfig) error {
	if azConfig == nil || azConfig.ClientID == "" {
		return cloudy.ErrInvalidConfiguration
	}

	if azConfig.Region == "" {
		azConfig.SetInstance(&AzurePublic)
	}
	if azConfig.SelectFields == nil {
		azConfig.SelectFields = DefaultUserSelectFields
	}

	scopes := []string{"https://graph.microsoft.us/.default"}

	cred, err := azidentity.NewClientSecretCredential(azConfig.TenantID, azConfig.ClientID, azConfig.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: policy.ClientOptions{
				Cloud: cloud.AzureGovernment,
			},
		})

	if err != nil {
		fmt.Printf("MsGraph Configure Error authentication provider: %v\n", err)
		return err
	}
	auth, err := msauth.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		fmt.Printf("MsGraph Configure Error authentication provider: %v\n", err)
		return err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return err
	}

	adapter.SetBaseUrl(azConfig.APIBase)

	azUM.Cfg = azConfig
	azUM.Adapter = adapter
	azUM.Client = msgraphsdk.NewGraphServiceClient(adapter)

	return err
}

func (graph *MsGraph) DebugSerialize(v serialization.Parsable) {
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

func NewGraph(ctx context.Context, tenantID string, clientID string, clientSecret string) (*MsGraph, error) {
	scopes := []string{"https://graph.microsoft.us/.default"}

	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: policy.ClientOptions{
				Cloud: cloud.AzureGovernment,
			},
		})

	if err != nil {
		fmt.Printf("NewGraph Error authentication provider: %v\n", err)
		return nil, err
	}
	auth, err := msauth.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		fmt.Printf("NewGraph Error authentication provider: %v\n", err)
		return nil, err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return nil, err
	}

	return &MsGraph{
		Adapter: adapter,
		Client:  msgraphsdk.NewGraphServiceClient(adapter),
	}, nil
}
