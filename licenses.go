package cloudymsgraph

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/license"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/google/uuid"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/users"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
)

func init() {
	license.LicenseProviders.Register(MsGraphName, &MsGraphLicenseManagerFactory{})
}

type MsGraphLicenseManager struct {
	*MsGraph
}

var GCCHighTeamsPhone = "985fcb26-7b94-475b-b512-89356697be71"
var GCCHighVisioPlan1 = "50a4284d-f0b2-4779-8de6-72c5f9b4349f"
var GCCHighTeamsAudio = "4dee1f32-0808-4fd2-a2ed-fdd575e3a45f"
var GCCHighOffice365E3 = "aea38a85-9bd5-4981-aa00-616b411205bf"
var GCCHighProjectPlan3 = "64758d81-92b7-4855-bcac-06617becb3e8"

type MsGraphLicenseManagerFactory struct {
	MsGraph
}

func NewMsGraphLicenseManager(ctx context.Context, cfg *MsGraphConfig) (*MsGraphLicenseManager, error) {
	lm := &MsGraphLicenseManager{
		MsGraph: &MsGraph{},
	}
	err := lm.Configure(cfg)

	return lm, err
}

func (lm *MsGraphLicenseManagerFactory) Create(cfg interface{}) (license.LicenseManager, error) {
	return NewMsGraphLicenseManager(context.Background(), cfg.(*MsGraphConfig))
}

func (lm *MsGraphLicenseManagerFactory) FromEnv(env *cloudy.Environment) (interface{}, error) {
	cfg := fromEnvironment(env)
	return cfg, nil
}

func (lm *MsGraphLicenseManager) AssignLicense(ctx context.Context, userId string, licenseSkus ...string) error {
	//body := users.NewItemMicrosoftGraphAssignLicenseAssignLicensePostRequestBody()
	body := users.NewItemAssignLicensePostRequestBody()

	assignedLicenses := []models.AssignedLicenseable{}
	for _, sku := range licenseSkus {
		skuId, err := uuid.Parse(sku)
		if err != nil {
			return cloudy.Error(ctx, "AssignLicense Invalid license: %s %v", sku, err)
		}

		assignedLicense := models.NewAssignedLicense()
		assignedLicense.SetSkuId(&skuId)
		assignedLicenses = append(assignedLicenses, assignedLicense)
	}

	body.SetAddLicenses(assignedLicenses)
	body.SetRemoveLicenses([]uuid.UUID{})

	_, err := lm.Client.UsersById(userId).AssignLicense().Post(ctx, body, nil)

	return err
}

func (lm *MsGraphLicenseManager) RemoveLicense(ctx context.Context, userId string, licenseSkus ...string) error {
	body := users.NewItemAssignLicensePostRequestBody()

	body.SetAddLicenses([]models.AssignedLicenseable{})

	removedLicenses := []uuid.UUID{}
	for _, sku := range licenseSkus {
		skuId, err := uuid.Parse(sku)
		if err != nil {
			return cloudy.Error(ctx, "RemoveLicense Invalid license: %s %v", sku, err)
		}

		removedLicenses = append(removedLicenses, skuId)
	}

	body.SetRemoveLicenses(removedLicenses)

	_, err := lm.Client.UsersById(userId).AssignLicense().Post(ctx, body, nil)

	return err
}

func (lm *MsGraphLicenseManager) GetUserAssigned(ctx context.Context, uid string) ([]*license.LicenseDescription, error) {
	result, err := lm.Client.UsersById(uid).Get(ctx,
		&users.UserItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
				Select: []string{"assignedLicenses"},
			},
		})

	if err != nil {
		oerr := err.(*odataerrors.ODataError)
		code := *oerr.GetError().GetCode()

		if code == "Request_ResourceNotFound" {
			return nil, nil
		}

		return nil, err
	}

	rtn := []*license.LicenseDescription{}
	for _, l := range result.GetAssignedLicenses() {
		rtn = append(rtn,
			&license.LicenseDescription{
				SKU: l.GetSkuId().String(),
			})
	}

	return rtn, nil

}

// GetAssigned gets a list of all the users with licenses
// https://graph.microsoft.com/v1.0/users?$filter=assignedLicenses/any(s:s/skuId eq 184efa21-98c3-4e5d-95ab-d07053a96e67)
// SEE : https://docs.microsoft.com/en-us/graph/query-parameters#filter-parameter
func (lm *MsGraphLicenseManager) GetAssigned(ctx context.Context, licenseSku string) ([]*cloudymodels.User, error) {
	filter := fmt.Sprintf("assignedLicenses/any(s:s/skuId eq %v)", licenseSku)
	fields := DefaultUserSelectFields

	result, err := lm.Client.Users().Get(ctx,
		&users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Select: fields,
				Filter: &filter,
			},
		})
	if err != nil {
		return nil, err
	}

	var rtn []*cloudymodels.User
	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](result, lm.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, err
	}

	err = pageIterator.Iterate(ctx, func(pageItem models.Userable) bool {
		rtn = append(rtn, UserToCloudy(pageItem))
		return true
	})
	if err != nil {
		return nil, err
	}

	return rtn, nil
}

// ListLicenses List all the managed licenses
func (lm *MsGraphLicenseManager) ListLicenses(ctx context.Context) ([]*license.LicenseDescription, error) {
	result, err := lm.Client.SubscribedSkus().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	all := result.GetValue()
	rtn := make([]*license.LicenseDescription, len(all))
	for i, lic := range all {
		var cnt int32
		var used int32

		if lic.GetConsumedUnits() != nil {
			used = *lic.GetConsumedUnits()
		}
		if lic.GetPrepaidUnits() != nil && lic.GetPrepaidUnits().GetEnabled() != nil {
			cnt = *lic.GetPrepaidUnits().GetEnabled()
		}

		rtn[i] = &license.LicenseDescription{
			ID:       *lic.GetId(),
			SKU:      lic.GetSkuId().String(),
			Name:     *lic.GetSkuPartNumber(),
			Assigned: int(used),
			Total:    int(cnt),
		}
	}

	return rtn, nil
}
