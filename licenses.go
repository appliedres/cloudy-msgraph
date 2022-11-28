package cloudymsgraph

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/license"
	cloudymodels "github.com/appliedres/cloudy/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/assignlicense"
)

func init() {
	license.LicenseProviders.Register(MSGraphName, &MsGraphLicenseManagerFactory{})
}

type MSGraphLicenseManager struct {
	*MSGraph
}

var GCCHighTeamsPhone = "985fcb26-7b94-475b-b512-89356697be71"
var GCCHighVisioPlan1 = "50a4284d-f0b2-4779-8de6-72c5f9b4349f"
var GCCHighTeamsAudio = "4dee1f32-0808-4fd2-a2ed-fdd575e3a45f"
var GCCHighOffice365E3 = "aea38a85-9bd5-4981-aa00-616b411205bf"
var GCCHighProjectPlan3 = "64758d81-92b7-4855-bcac-06617becb3e8"

type MsGraphLicenseManagerFactory struct {
	MSGraph
}

func NewMSGraphLicenseManager(ctx context.Context, cfg *MSGraphConfig) (*MSGraphLicenseManager, error) {
	gm := &MSGraphLicenseManager{
		MSGraph: &MSGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (ms *MsGraphLicenseManagerFactory) Create(cfg interface{}) (license.LicenseManager, error) {
	return NewMSGraphLicenseManager(context.Background(), cfg.(*MSGraphConfig))
}

func (ms *MsGraphLicenseManagerFactory) FromEnv(env *cloudy.Environment) (interface{}, error) {
	cfg := fromEnvironment(env)
	return cfg, nil
}

func NewMsGraphLicenseManager(ctx context.Context, cfg *MSGraphConfig) (*MSGraphLicenseManager, error) {
	gm := &MSGraphLicenseManager{
		MSGraph: &MSGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (lm *MSGraphLicenseManager) AssignLicense(ctx context.Context, userId string, licenseSkus ...string) error {
	body := &assignlicense.AssignLicensePostRequestBody{}

	var licenses []models.AssignedLicenseable
	for _, l := range licenseSkus {
		al := models.NewAssignedLicense()
		al.SetSkuId(&l)
		licenses = append(licenses, al)
	}

	body.SetAddLicenses(licenses)
	body.SetRemoveLicenses([]string{})

	lm.DebugSerialize(body)

	_, err := lm.Client.UsersById(userId).AssignLicense().Post(ctx, body, nil)

	return err
}

func (lm *MSGraphLicenseManager) RemoveLicense(ctx context.Context, userId string, licenseSkus ...string) error {
	body := &assignlicense.AssignLicensePostRequestBody{}
	body.SetAddLicenses([]models.AssignedLicenseable{})
	body.SetRemoveLicenses(licenseSkus)
	_, err := lm.Client.UsersById(userId).AssignLicense().Post(ctx, body, nil)

	return err
}

func (lm *MSGraphLicenseManager) GetUserAssigned(ctx context.Context, uid string) ([]*license.LicenseDescription, error) {
	fields := []string{"assignedLicenses"}

	result, err := lm.Client.UsersById(uid).Get(ctx,
		&item.UserItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &item.UserItemRequestBuilderGetQueryParameters{
				Select: fields,
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

	assigned := result.GetAssignedLicenses()
	rtn := make([]*license.LicenseDescription, len(assigned))

	for i, l := range assigned {
		rtn[i] = &license.LicenseDescription{
			SKU: *l.GetSkuId(),
		}
	}

	return rtn, nil

}

// GetAssigned gets a list of all the users with licenses
// https://graph.microsoft.com/v1.0/users?$filter=assignedLicenses/any(s:s/skuId eq 184efa21-98c3-4e5d-95ab-d07053a96e67)
// SEE : https://docs.microsoft.com/en-us/graph/query-parameters#filter-parameter
func (lm *MSGraphLicenseManager) GetAssigned(ctx context.Context, licenseSku string) ([]*cloudymodels.User, error) {
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
	pageIterator, err := msgraphcore.NewPageIterator(result, lm.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, err
	}

	err = pageIterator.Iterate(ctx, func(pageItem interface{}) bool {
		u := pageItem.(models.Userable)
		rtn = append(rtn, UserToCloudy(u))
		return true
	})
	if err != nil {
		return nil, err
	}

	return rtn, nil
}

// ListLicenses List all the managed licenses
func (lm *MSGraphLicenseManager) ListLicenses(ctx context.Context) ([]*license.LicenseDescription, error) {
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
			SKU:      *lic.GetSkuId(),
			Name:     *lic.GetSkuPartNumber(),
			Assigned: int(used),
			Total:    int(cnt),
		}
	}

	return rtn, nil
}
