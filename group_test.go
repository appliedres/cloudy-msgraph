package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestGroupManager(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "KEYVAULT")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()
	tenantID := env.Force("AZ_TENANT_ID")
	ClientID := env.Force("AZ_CLIENT_ID")
	ClientSecret := env.Force("AZ_CLIENT_SECRET")

	cfg := &MsGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	cfg.SetInstance(&USGovernment)

	gm, err := NewMsGraphGroupManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestGroupManager(t, gm, um)

}

func TestListGroups(t *testing.T) {

	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "KEYVAULT")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()
	tenantID := env.Force("AZ_TENANT_ID")
	ClientID := env.Force("AZ_CLIENT_ID")
	ClientSecret := env.Force("AZ_CLIENT_SECRET")

	cfg := &MsGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}

	cfg.SetInstance(&USGovernment)

	gm, err := NewMsGraphGroupManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	groups, _ := gm.ListGroups(ctx)

	for _, group := range groups {
		_, _ = gm.GetGroupMembers(ctx, group.ID)
	}
}
