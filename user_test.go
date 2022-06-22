package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestUserManager(t *testing.T) {
	ctx := cloudy.StartContext()

	testutil.LoadEnv("test.env")
	tenantID := cloudy.ForceEnv("TenantID", "")
	ClientID := cloudy.ForceEnv("ClientID", "")
	ClientSecret := cloudy.ForceEnv("ClientSecret", "")

	cfg := &MSGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	cfg.SetInstance(&USGovernment)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestUserManager(t, um)

}
