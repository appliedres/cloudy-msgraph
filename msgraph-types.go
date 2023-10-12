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
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

const MsGraphName = "msgraph"
const MsGraphVersionV1 = "v1.0"
const MsGraphVersionBeta = "beta"

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
	Version      string
	SelectFields []string
}

func (azcfg *MsGraphConfig) SetInstanceName(name string) error {
	if strings.EqualFold(name, USGovernment.Name) {
		azcfg.SetInstance(&USGovernment)
		return nil
	} else if strings.EqualFold(name, AzurePublic.Name) {
		azcfg.SetInstance(&AzurePublic)
		return nil
	}

	return ErrInvalidInstanceName
}

func (azcfg *MsGraphConfig) SetInstance(instance *MsGraphInstance) {
	azcfg.APIBase = instance.Base
	azcfg.Region = instance.Login
}

type MsGraph struct {
	Client  *msgraphsdk.GraphServiceClient
	Adapter *msgraphsdk.GraphRequestAdapter
	Cfg     *MsGraphConfig
}

func (azum *MsGraph) Configure(azCfg *MsGraphConfig) error {
	if azCfg == nil || azCfg.ClientID == "" {
		return cloudy.ErrInvalidConfiguration
	}

	if azCfg.Version == "" {
		azCfg.Version = MsGraphVersionV1
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
		fmt.Printf("MsGraph Configure Error authentication provider: %v\n", err)
		return err
	}
	auth, err := a.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
	if err != nil {
		fmt.Printf("MsGraph Configure Error authentication provider: %v\n", err)
		return err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return err
	}

	adapter.SetBaseUrl(azCfg.APIBase)

	azum.Cfg = azCfg
	azum.Adapter = adapter
	azum.Client = msgraphsdk.NewGraphServiceClient(adapter)

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
	auth, err := a.NewAzureIdentityAuthenticationProviderWithScopes(cred, scopes)
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
