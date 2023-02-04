package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestGroupManager(t *testing.T) {
	ctx := cloudy.StartContext()

	_ = testutil.LoadEnv("collider.env")
	tenantID := cloudy.ForceEnv("TENANT_ID", "")
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

	groups, _ := gm.ListGroups(ctx)

	for _, group := range groups {
		_, _ = gm.GetGroupMembers(ctx, group.ID)
	}

	// um, err := NewMsGraphUserManager(ctx, cfg)
	// if err != nil {
	// 	log.Fatalf("Error %v", err)
	// }

	// testutil.TestGroupManager(t, gm, um)

}
