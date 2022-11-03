package cloudymsgraph

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	"github.com/microsoftgraph/msgraph-sdk-go/groups/item/members"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/getmembergroups"
)

func init() {
	cloudy.GroupProviders.Register(MSGraphName, &MsGraphGroupManagerFactory{})
}

type MsGraphGroupManagerFactory struct {
	MSGraph
}

func NewMSGraphGroupManager(ctx context.Context, cfg *MSGraphConfig) (*MSGraphGroupManager, error) {
	gm := &MSGraphGroupManager{
		MSGraph: &MSGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (ms *MsGraphGroupManagerFactory) Create(cfg interface{}) (cloudy.GroupManager, error) {
	return NewMSGraphGroupManager(context.Background(), cfg.(*MSGraphConfig))
}

func (ms *MsGraphGroupManagerFactory) FromEnv(env *cloudy.SegmentedEnvironment) (interface{}, error) {
	cfg := fromEnvironment(env)
	return cfg, nil
}

type MSGraphGroupManager struct {
	*MSGraph
}

// List all the groups available
func (gm *MSGraphGroupManager) ListGroups(ctx context.Context) ([]*models.Group, error) {
	cloudy.Info(ctx, "Listing Groups")
	grps, err := gm.Client.Groups().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	cloudy.Info(ctx, "Creating Group array")
	groups := grps.GetValue()
	rtn := make([]*models.Group, len(groups))
	for i, g := range groups {
		rtn[i] = gm.ToCloudy(g)
	}

	return rtn, nil
}

// Get all the groups for a single user
func (gm *MSGraphGroupManager) GetUserGroups(ctx context.Context, uid string) ([]*models.Group, error) {

	data := getmembergroups.NewGetMemberGroupsPostRequestBody()
	data.SetSecurityEnabledOnly(cloudy.BoolP(false))

	results, err := gm.Client.UsersById(uid).GetMemberGroups().Post(ctx, data, nil)
	if err != nil {
		return nil, err
	}
	groupIds := results.GetValue()

	groups, err := gm.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	rtn := make([]*models.Group, len(groupIds))
	for i, gid := range groupIds {
		for _, grp := range groups {
			if grp.ID == gid {
				rtn[i] = grp
				break
			}
		}
	}

	return rtn, nil
}

func (gm *MSGraphGroupManager) DeleteGroup(ctx context.Context, groupId string) error {
	return gm.Client.GroupsById(groupId).Delete(ctx, nil)
}

func (gm *MSGraphGroupManager) GetGroup(ctx context.Context, id string) (*models.Group, error) {
	result, err := gm.Client.GroupsById(id).Get(ctx, nil)
	if err != nil {
		oerr := err.(*odataerrors.ODataError)
		code := *oerr.GetError().GetCode()

		if code == "Request_ResourceNotFound" {
			return nil, nil
		}

		return nil, err
	}
	return gm.ToCloudy(result), nil
}

// Create a new Group
func (gm *MSGraphGroupManager) NewGroup(ctx context.Context, grp *models.Group) (*models.Group, error) {
	g := gm.ToAzure(grp)

	result, err := gm.Client.Groups().Post(ctx, g, nil)
	newGrp := gm.ToCloudy(result)
	cloudy.Info(ctx, "New group created, %+v", result)

	return newGrp, err
}

// Update a group. This is generally just the name of the group.
func (gm *MSGraphGroupManager) UpdateGroup(ctx context.Context, grp *models.Group) (bool, error) {
	g := &graphmodels.Group{}
	g.SetId(&grp.ID)
	g.SetDisplayName(&grp.Name)

	_, err := gm.Client.GroupsById(grp.ID).Patch(ctx, g, nil)
	return true, err
}

// Get all the members of a group. This returns partial users only,
// typically just the user id, name and email fields
func (gm *MSGraphGroupManager) GetGroupMembers(ctx context.Context, grpId string) ([]*models.User, error) {

	result, err := gm.Client.GroupsById(grpId).Members().Get(ctx,
		&members.MembersRequestBuilderGetRequestConfiguration{
			QueryParameters: &members.MembersRequestBuilderGetQueryParameters{
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
		return nil, err
	}

	dirObjects := result.GetValue()
	rtn := make([]*models.User, len(dirObjects))
	for i, obj := range dirObjects {
		userable := obj.(graphmodels.Userable)
		rtn[i] = UserToCloudy(userable)
	}

	return rtn, nil
}

// Remove members from a group
func (gm *MSGraphGroupManager) RemoveMembers(ctx context.Context, groupId string, userIds []string) error {
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
func (gm *MSGraphGroupManager) tempRecover() {
	err := recover()
	if err != nil {
		fmt.Println(err)
	}
}

// Add member(s) to a group
func (gm *MSGraphGroupManager) AddMembers(ctx context.Context, groupId string, userIds []string) error {
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

	// for _, userId := range userIds {
	// 	ref := ref.NewRef()
	// 	ref.SetAdditionalData(map[string]interface{}{
	// 		"@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/" + userId,
	// 	})

	// 	oneErr := gm.Client.GroupsById(groupId).Members().Ref().Post(ctx, ref, nil)
	// 	if oneErr != nil {
	// 		err.Append(oneErr)
	// 	}
	// }

	// if err.HasError() {
	// 	return err
	// }
	// return nil
}

func (gm *MSGraphGroupManager) ToCloudy(g graphmodels.Groupable) *models.Group {
	cg := &models.Group{}
	cg.ID = *g.GetId()
	cg.Name = *g.GetDisplayName()

	return cg
}

func (gm *MSGraphGroupManager) ToAzure(cg *models.Group) *graphmodels.Group {
	g := &graphmodels.Group{}
	g.SetId(&cg.ID)
	g.SetDisplayName(&cg.Name)
	g.SetMailEnabled(cloudy.BoolP(false))
	g.SetMailNickname(&cg.Name)
	g.SetSecurityEnabled(cloudy.BoolP(true))

	return g
}
