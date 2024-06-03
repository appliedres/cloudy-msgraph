package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGroupManager(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	em := cloudy.GetDefaultEnvManager()
	em.LoadSources("test")

	um, err := cloudy.UserProviders.NewFromEnvMgr(em, "USER_DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the user manager. %v", err)
	}

	// Get the default group Manager
	gm, err := cloudy.GroupProviders.NewFromEnvMgr(em, "GROUP_DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the group manager. %v", err)
	}

	testutil.TestGroupManager(t, gm, um)

}

func TestListGroups(t *testing.T) {

	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	em := cloudy.GetDefaultEnvManager()
	em.LoadSources("test")

	ctx := cloudy.StartContext()

	gm, err := cloudy.GroupProviders.NewFromEnvMgr(em, "GROUP_DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the group manager. %v", err)
	}

	groups, _ := gm.ListGroups(ctx)

	for _, group := range groups {
		_, _ = gm.GetGroupMembers(ctx, group.ID)
	}
}

func TestListUserGroups(t *testing.T) {

	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	email := "test.user@collider.onmicrosoft.us"

	em := cloudy.GetDefaultEnvManager()
	em.LoadSources("test")

	ctx := cloudy.StartContext()

	gm, err := cloudy.GroupProviders.NewFromEnvMgr(em, "GROUP_DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the group manager. %v", err)
	}

	groups, _ := gm.GetUserGroups(ctx, email)

	for _, group := range groups {
		cloudy.Info(ctx, "user: %s member of %s", email, group.Name)
	}
}

func TestGetGroupId(t *testing.T) {

	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	em := cloudy.GetDefaultEnvManager()
	em.LoadSources("test")

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnvMgr(em).(*MsGraphConfig)

	gm, err := NewMsGraphGroupManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	groupName := "IL5"
	groupId, err := gm.GetGroupId(ctx, groupName)
	assert.Nil(t, err)
	assert.NotNil(t, groupId)

	_ = groupId
}
