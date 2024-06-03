package cloudymsgraph

import (
	"context"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

func init() {
	cloudy.InviteProviders.Register(MsGraphName, &MsGraphInviteManagerFactory{}, []cloudy.EnvDefinition{})
}

type MsGraphInviteManagerFactory struct {
	MsGraph
}

func NewMsGraphInviteManager(ctx context.Context, cfg *MsGraphConfig) (*MsGraphInviteManager, error) {
	gm := &MsGraphInviteManager{
		MsGraph: &MsGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (ms *MsGraphInviteManagerFactory) Create(cfg interface{}) (cloudy.InviteManager, error) {
	return NewMsGraphInviteManager(context.Background(), cfg.(*MsGraphConfig))
}

func (ms *MsGraphInviteManagerFactory) FromEnvMgr(em *cloudy.EnvManager, prefix string) (interface{}, error) {
	cfg := fromEnvironment(em)
	return cfg, nil
}

type MsGraphInviteManager struct {
	*MsGraph
}

func (im *MsGraphInviteManager) CreateInvitation(ctx context.Context, user *cloudymodels.User, emailInvite bool, inviteRedirectUrl string) error {
	requestBody := graphmodels.NewInvitation()
	requestBody.SetInvitedUserEmailAddress(&user.Email)
	requestBody.SetInvitedUserDisplayName(&user.DisplayName)
	requestBody.SetSendInvitationMessage(&emailInvite)
	requestBody.SetInviteRedirectUrl(&inviteRedirectUrl)
	_, err := im.Client.Invitations().Post(ctx, requestBody, nil)
	return err
}
