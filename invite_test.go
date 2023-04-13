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

	_ = testutil.LoadEnv("collider.env")
	tenantID := cloudy.ForceEnv("TenantID", "")
	ClientID := cloudy.ForceEnv("ClientID", "")
	ClientSecret := cloudy.ForceEnv("ClientSecret", "")

	cfg := &MsGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
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
