package cloudymsgraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	cloudymodels "github.com/appliedres/cloudy/models"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/groups"
	graphmodels "github.com/microsoftgraph/msgraph-beta-sdk-go/models"
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
	allGroups, err := gm.Client.Groups().Get(ctx, nil)
	if err != nil {
		_ = cloudy.Error(ctx, "MsGraphGroupManager error: %v", err)
		return nil, err
	}

	cloudy.Info(ctx, "MsGraphGroupManager Creating Group array")
	groups := allGroups.GetValue()
	rtn := []*models.Group{}
	for _, g := range groups {
		rtn = append(rtn, GroupToCloudy(g))
	}

	cloudy.Info(ctx, "MsGraphGroupManager Creating Group array complete")

	return rtn, nil
}

// Get all the groups for a single user
func (gm *MsGraphGroupManager) GetUserGroups(ctx context.Context, uid string) ([]*cloudymodels.Group, error) {

	cloudy.Info(ctx, "GetUserGroups: %s", uid)

	results, err := gm.Client.Users().ByUserId(uid).MemberOf().Get(context.Background(), nil)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			return nil, cloudy.Error(ctx, "GetUserGroups Error: %s - ResourceNotFound - %s", uid, message)
		}

		return nil, cloudy.Error(ctx, "GetUserGroups Error: %s %s", uid, message)
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
	return gm.Client.Groups().ByGroupId(groupId).Delete(ctx, nil)
}

func (gm *MsGraphGroupManager) GetGroup(ctx context.Context, id string) (*models.Group, error) {
	result, err := gm.Client.Groups().ByGroupId(id).Get(ctx, nil)
	if err != nil {
		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			return nil, cloudy.Error(ctx, "GetGroup Error: %s - ResourceNotFound - %s", id, message)
		}

		return nil, cloudy.Error(ctx, "GetGroup Error: %s %s", id, message)
	}

	return GroupToCloudy(result), nil
}

func (gm *MsGraphGroupManager) GetGroupId(ctx context.Context, name string) (string, error) {
	cloudy.Info(ctx, "MsGraphGroupManager get group id by display name %v", name)
	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")

	requestFilter := fmt.Sprintf("displayName eq '%v'", name)
	requestParameters := &groups.GroupsRequestBuilderGetQueryParameters{
		Filter: &requestFilter,
	}

	configuration := &groups.GroupsRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	result, err := gm.Client.Groups().Get(ctx, configuration)
	if err != nil {

		code, message := GetErrorCodeAndMessage(ctx, err)

		if strings.EqualFold(code, ResourceNotFoundCode) {
			return "", cloudy.Error(ctx, "GetGroupId Error: %s - ResourceNotFound - %s", name, message)
		}

		return "", cloudy.Error(ctx, "GetGroupId Error: %s %s", name, message)
	}

	var rtn []*cloudymodels.Group

	groups := result.GetValue()
	for _, g := range groups {
		rtn = append(rtn, GroupToCloudy(g))
	}

	return rtn[0].ID, nil

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

	_, err := gm.Client.Groups().ByGroupId(grp.ID).Patch(ctx, g, nil)
	return true, err
}

// Get all the members of a group. This returns partial users only,
// typically just the user id, name and email fields
func (gm *MsGraphGroupManager) GetGroupMembers(ctx context.Context, grpId string) ([]*models.User, error) {

	cloudy.Info(ctx, "MsGraphGroupManager GetGroupMembers grpId: %s", grpId)

	result, err := gm.Client.Groups().ByGroupId(grpId).Members().Get(ctx,
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

	cloudy.Info(ctx, "MsGraphGroupManager GetGroupMembers grpId: %s found %d", grpId, len(rtn))

	return rtn, nil
}

// Remove members from a group
func (gm *MsGraphGroupManager) RemoveMembers(ctx context.Context, groupId string, userIds []string) error {
	cloudy.Info(ctx, "MsGraphGroupManager RemoveMembers")

	err := cloudy.MultiError()
	for _, userId := range userIds {
		oneErr := gm.Client.Groups().ByGroupId(groupId).Members().ByDirectoryObjectId(userId).Ref().Delete(ctx, nil)
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
		_ = cloudy.Error(context.Background(), "MsGraphGroupManager Temp Recover: %v", err)
	}
}

// Add member(s) to a group
func (gm *MsGraphGroupManager) AddMembers(ctx context.Context, groupId string, userIds []string) error {
	cloudy.Info(ctx, "MsGraphGroupManager AddMembers")

	defer gm.tempRecover()

	// Something's wrong with the bulk operation
	// newMembers := []string{}
	// for _, userId := range userIds {
	// 	newMembers = append(newMembers, "https://graph.microsoft.com/v1.0/directoryObjects/"+userId)
	// }

	// requestBody := graphmodels.NewGroup()
	// additionalData := requestBody.GetAdditionalData()
	// additionalData["members@odata.bind"] = newMembers
	// requestBody.SetAdditionalData(additionalData)

	// _, err := gm.Client.GroupsById(groupId).Patch(ctx, requestBody, nil)

	errs := cloudy.MultiError()
	for _, userId := range userIds {
		requestBody := graphmodels.NewReferenceCreate()
		odataId := "https://graph.microsoft.com/v1.0/directoryObjects/" + userId
		requestBody.SetOdataId(&odataId)

		err := gm.Client.Groups().ByGroupId(groupId).Members().Ref().Post(ctx, requestBody, nil)

		if err != nil {
			errs.Append(err)
		}
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func GroupToCloudy(g graphmodels.Groupable) *models.Group {
	cg := &models.Group{}
	cg.ID = *g.GetId()
	cg.Name = *g.GetDisplayName()

	return cg
}

func GroupToAzure(cg *models.Group) *graphmodels.Group {

	group := graphmodels.NewGroup()
	group.SetId(&cg.ID)
	group.SetDisplayName(&cg.Name)
	group.SetMailEnabled(cloudy.BoolP(false))
	group.SetMailNickname(&cg.Name)
	group.SetSecurityEnabled(cloudy.BoolP(true))

	return group
}
