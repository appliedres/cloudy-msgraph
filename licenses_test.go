package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestLicenseManager(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "KEYVAULT")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()
	tenantID := env.Force("AZ_TENANT_ID")
	ClientID := env.Force("AZ_CLIENT_ID")
	ClientSecret := env.Force("AZ_CLIENT_SECRET")
	TestUser := env.Force("TEST_USER")
	TestSku := env.Force("TEST_SKU")

	cfg := &MsGraphConfig{
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
