package cloudymsgraph

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

func init() {
	cloudy.GroupProviders.Register(MsGraphName, &MsGraphGroupManagerFactory{})
}

type MsGraphGroupManagerFactory struct {
	MsGraph
}

func NewMsGraphGroupManager(ctx context.Context, cfg *MsGraphConfig) (*MsGraphGroupManager, error) {
	gm := &MsGraphGroupManager{
		MsGraph: &MsGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (ms *MsGraphGroupManagerFactory) Create(cfg interface{}) (cloudy.GroupManager, error) {
	return NewMsGraphGroupManager(context.Background(), cfg.(*MsGraphConfig))
}

func (ms *MsGraphGroupManagerFactory) FromEnv(env *cloudy.Environment) (interface{}, error) {
	cfg := fromEnvironment(env)
	return cfg, nil
}

type MsGraphGroupManager struct {
	*MsGraph
}

// List all the groups available
func (gm *MsGraphGroupManager) ListGroups(ctx context.Context) ([]*models.Group, error) {
	cloudy.Info(ctx, "MsGraphGroupManager Listing Groups")
	grps, err := gm.Client.Groups().Get(ctx, nil)
	if err != nil {
		_ = cloudy.Error(ctx, "MsGraphGroupManager error: %v", err)
		return nil, err
	}

	cloudy.Info(ctx, "MsGraphGroupManager Creating Group array")
	groups := grps.GetValue()
	rtn := []*models.Group{}
	for _, g := range groups {
		rtn = append(rtn, GroupToCloudy(g))
	}

	return rtn, nil
}

// Get all the groups for a single user
func (gm *MsGraphGroupManager) GetUserGroups(ctx context.Context, uid string) ([]*cloudymodels.Group, error) {

	// Does not apply to Get
	// data := users.NewItemMicrosoftGraphGetMemberGroupsGetMemberGroupsPostRequestBody()
	// data.SetSecurityEnabledOnly(cloudy.BoolP(false))

	results, err := gm.Client.UsersById(uid).MemberOf().Get(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	rtn := []*cloudymodels.Group{}
	for _, msGroup := range results.GetValue() {
		switch groupData := msGroup.(type) {
		case graphmodels.Groupable:
			rtn = append(rtn, GroupToCloudy(groupData))
		}
	}

	return rtn, nil
}

func (gm *MsGraphGroupManager) DeleteGroup(ctx context.Context, groupId string) error {
	return gm.Client.GroupsById(groupId).Delete(ctx, nil)
}

func (gm *MsGraphGroupManager) GetGroup(ctx context.Context, id string) (*models.Group, error) {
	result, err := gm.Client.GroupsById(id).Get(ctx, nil)
	if err != nil {
		oerr := err.(*odataerrors.ODataError)
		code := *oerr.GetError().GetCode()

		if code == "Request_ResourceNotFound" {
			return nil, nil
		}

		return nil, err
	}
	return GroupToCloudy(result), nil
}

// Create a new Group
func (gm *MsGraphGroupManager) NewGroup(ctx context.Context, grp *models.Group) (*models.Group, error) {
	g := GroupToAzure(grp)

	result, err := gm.Client.Groups().Post(ctx, g, nil)
	newGrp := GroupToCloudy(result)
	cloudy.Info(ctx, "New group created, %+v", result)

	return newGrp, err
}

// Update a group. This is generally just the name of the group.
func (gm *MsGraphGroupManager) UpdateGroup(ctx context.Context, grp *models.Group) (bool, error) {
	g := &graphmodels.Group{}
	g.SetId(&grp.ID)
	g.SetDisplayName(&grp.Name)

	_, err := gm.Client.GroupsById(grp.ID).Patch(ctx, g, nil)
	return true, err
}

// Get all the members of a group. This returns partial users only,
// typically just the user id, name and email fields
func (gm *MsGraphGroupManager) GetGroupMembers(ctx context.Context, grpId string) ([]*models.User, error) {

	cloudy.Info(ctx, "MsGraphGroupManager GetGroupMembers grpId: %s", grpId)

	result, err := gm.Client.GroupsById(grpId).Members().Get(ctx,
		&groups.ItemMembersRequestBuilderGetRequestConfiguration{
			QueryParameters: &groups.ItemMembersRequestBuilderGetQueryParameters{
				Select: []string{
					"id",
					"displayName",
					"givenName",
					"surname",
					"userPrincipalName",
				},
			},
		})
	if err != nil {
		return nil, cloudy.Error(ctx, "GetGroupMembers (%s) Failed %v", grpId, err)
	}

	dirObjects := result.GetValue()
	rtn := []*models.User{}
	for _, dirObj := range dirObjects {
		switch data := dirObj.(type) {
		case graphmodels.Userable:
			rtn = append(rtn, UserToCloudy(data))
			// default:
			// 	cloudy.Info(ctx, "Non-User directory object: %T", dirObj)
		}

	}

	return rtn, nil
}

// Remove members from a group
func (gm *MsGraphGroupManager) RemoveMembers(ctx context.Context, groupId string, userIds []string) error {
	err := cloudy.MultiError()
	for _, userId := range userIds {
		oneErr := gm.Client.GroupsById(groupId).MembersById(userId).Ref().Delete(ctx, nil)
		if oneErr != nil {
			err.Append(oneErr)
		}
	}

	if err.HasError() {
		return err
	}
	return nil
}

// SEE : https://github.com/microsoftgraph/msgraph-sdk-go/issues/155#issuecomment-1156264835
func (gm *MsGraphGroupManager) tempRecover() {
	err := recover()
	if err != nil {
		fmt.Println(err)
	}
}

// Add member(s) to a group
func (gm *MsGraphGroupManager) AddMembers(ctx context.Context, groupId string, userIds []string) error {
	defer gm.tempRecover()

	newMembers := []string{}
	for _, userId := range userIds {
		newMembers = append(newMembers, "https://graph.microsoft.com/v1.0/directoryObjects/"+userId)
	}

	requestBody := graphmodels.NewGroup()
	additionalData := map[string]interface{}{"members@odata.bind": newMembers}
	requestBody.SetAdditionalData(additionalData)
	_, err := gm.Client.GroupsById(groupId).Patch(ctx, requestBody, nil)

	return err
}

func GroupToCloudy(g graphmodels.Groupable) *models.Group {
	cg := &models.Group{}
	cg.ID = *g.GetId()
	cg.Name = *g.GetDisplayName()

	return cg
}

func GroupToAzure(cg *models.Group) *graphmodels.Group {
	g := &graphmodels.Group{}
	g.SetId(&cg.ID)
	g.SetDisplayName(&cg.Name)
	g.SetMailEnabled(cloudy.BoolP(false))
	g.SetMailNickname(&cg.Name)
	g.SetSecurityEnabled(cloudy.BoolP(true))

	return g
}
