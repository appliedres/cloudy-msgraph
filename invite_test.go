package cloudymsgraph

import (
	"testing"

	cloudymodels "github.com/appliedres/cloudy/models"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInviteManager(t *testing.T) {
	ctx := cloudy.StartContext()

	em := testutil.CreateTestEnvMgr()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnvMgr(em).(*MsGraphConfig)
	cfg.SetInstance(&USGovernment)

	inviteUser := &cloudymodels.User{
		UPN:         "some.testuser@collider.onmicrosoft.us",
		DisplayName: "some testuser",
		Email:       "sometestuser@gmail.com",
	}

	im, err := NewMsGraphInviteManager(ctx, cfg)
	assert.Nil(t, err)

	url := "https://dashboard.afrlcollider.us/signin"
	im.CreateInvitation(ctx, inviteUser, true, url)

}
