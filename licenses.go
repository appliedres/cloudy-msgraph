package cloudymsgraph

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy/license"
	cloudymodels "github.com/appliedres/cloudy/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/assignlicense"
)

type LicenseManager struct {
	*MSGraph
}

var GCCHighTeamsPhone = "985fcb26-7b94-475b-b512-89356697be71"
var GCCHighVisioPlan1 = "50a4284d-f0b2-4779-8de6-72c5f9b4349f"
var GCCHighTeamsAudio = "4dee1f32-0808-4fd2-a2ed-fdd575e3a45f"
var GCCHighOffice365E3 = "aea38a85-9bd5-4981-aa00-616b411205bf"
var GCCHighProjectPlan3 = "64758d81-92b7-4855-bcac-06617becb3e8"

func NewMsGraphLicenseManager(ctx context.Context, cfg *MSGraphConfig) (*LicenseManager, error) {
	gm := &LicenseManager{
		MSGraph: &MSGraph{},
	}
	err := gm.Configure(cfg)

	return gm, err
}

func (lm *LicenseManager) AssignLicense(ctx context.Context, userId string, licenseSkus ...string) error {
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

	_, err := lm.Client.UsersById(userId).AssignLicense().Post(body)

	return err
}

func (lm *LicenseManager) RemoveLicense(ctx context.Context, userId string, licenseSkus ...string) error {
	body := &assignlicense.AssignLicensePostRequestBody{}
	body.SetAddLicenses([]models.AssignedLicenseable{})
	body.SetRemoveLicenses(licenseSkus)
	_, err := lm.Client.UsersById(userId).AssignLicense().Post(body)

	return err
}

func (lm *LicenseManager) GetUserAssigned(ctx context.Context, uid string) ([]*license.LicenseDescription, error) {
	fields := []string{"assignedLicenses"}

	requestCfg := &item.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &item.UserItemRequestBuilderGetQueryParameters{
			Select: fields,
		},
	}

	result, err := lm.Client.UsersById(uid).GetWithRequestConfigurationAndResponseHandler(requestCfg, nil)
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

//GetAssigned gets a list of all the users with licenses
// https://graph.microsoft.com/v1.0/users?$filter=assignedLicenses/any(s:s/skuId eq 184efa21-98c3-4e5d-95ab-d07053a96e67)
// SEE : https://docs.microsoft.com/en-us/graph/query-parameters#filter-parameter
func (lm *LicenseManager) GetAssigned(ctx context.Context, licenseSku string) ([]*cloudymodels.User, error) {
	filter := fmt.Sprintf("assignedLicenses/any(s:s/skuId eq %v)", licenseSku)
	fields := DefaultUserSelectFields

	result, err := lm.Client.Users().GetWithRequestConfigurationAndResponseHandler(
		&users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Select: fields,
				Filter: &filter,
			},
		}, nil)
	if err != nil {
		return nil, err
	}

	var rtn []*cloudymodels.User
	pageIterator, err := msgraphcore.NewPageIterator(result, lm.Adapter, models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, err
	}

	err = pageIterator.Iterate(func(pageItem interface{}) bool {
		u := pageItem.(models.Userable)
		rtn = append(rtn, UserToCloudy(u))
		return true
	})
	if err != nil {
		return nil, err
	}

	return rtn, nil
}

//ListLicenses List all the managed licenses
func (lm *LicenseManager) ListLicenses(ctx context.Context) ([]*license.LicenseDescription, error) {
	result, err := lm.Client.SubscribedSkus().Get()
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
