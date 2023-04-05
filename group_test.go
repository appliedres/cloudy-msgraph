package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
)

func TestGroupManager(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

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

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

	gm, err := NewMsGraphGroupManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	groups, _ := gm.ListGroups(ctx)

	for _, group := range groups {
		_, _ = gm.GetGroupMembers(ctx, group.ID)
	}
}
