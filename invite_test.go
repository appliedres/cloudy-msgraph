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

	env := testutil.CreateTestEnvironment()
	cloudy.SetDefaultEnvironment(env)

	testEnv := env.Segment("TEST")
	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(testEnv).(*MsGraphConfig)
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
