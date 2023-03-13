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
	tenantID := "108d0f68-f06e-435d-bc00-555ef5ecc3b1"
	ClientID := "04a8c1c9-7248-4bab-addb-50c888998257"
	ClientSecret := "46h6N3V5iRXr7VNLy6._0wLF_FOVCF~xtX"

	//tenantID := cloudy.ForceEnv("TenantID", "")
	//ClientID := cloudy.ForceEnv("ClientID", "")
	//ClientSecret := cloudy.ForceEnv("ClientSecret", "")

	cfg := &MsGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	cfg.SetInstance(&USGovernment)

	inviteUser := &cloudymodels.User{
		UserName:    "bill.testuser@collider.onmicrosoft.us",
		DisplayName: "bill testuser",
		Email:       "williamflentje@gmail.com",
	}

	im, err := NewMsGraphInviteManager(ctx, cfg)
	assert.Nil(t, err)

	url := "https://dashboard.afrlcollider.us/signin"
	im.CreateInvitation(ctx, inviteUser, true, url)

}
