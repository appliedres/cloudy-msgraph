package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestLicenseManager(t *testing.T) {
	ctx := cloudy.StartContext()

	testutil.LoadEnv("test.env")
	tenantID := cloudy.ForceEnv("TenantID", "")
	ClientID := cloudy.ForceEnv("ClientID", "")
	ClientSecret := cloudy.ForceEnv("ClientSecret", "")
	TestUser := cloudy.ForceEnv("TestUser", "")
	TestSku := cloudy.ForceEnv("TestSku", "")

	cfg := &MSGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	cfg.SetInstance(&USGovernment)

	lm, err := NewMsGraphLicenseManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestLicenseManager(t, ctx, lm, TestUser, TestSku)
}
