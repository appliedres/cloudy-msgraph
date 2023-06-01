package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestLicenseManager(t *testing.T) {
	ctx := cloudy.StartContext()

	env := testutil.CreateTestEnvironment()
	cloudy.SetDefaultEnvironment(env)

	testEnv := env.Segment("TEST")
	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(testEnv).(*MsGraphConfig)
	cfg.SetInstance(&USGovernment)

	TestUser := "unittest@collider.onmicrosoft.us"
	TestSku := GCCHighAADP2

	lm, err := NewMsGraphLicenseManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestLicenseManager(t, ctx, lm, TestUser, TestSku)
}
