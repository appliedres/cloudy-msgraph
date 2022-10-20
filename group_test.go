package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestGroupManager(t *testing.T) {
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

	gm, err := NewMSGraphGroupManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestGroupManager(t, gm, um)

}
