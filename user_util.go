package cloudymsgraph

import (
	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

var DefaultUserSelectFields = []string{
	"businessPhones",
	"displayName",
	"givenName",
	"id",
	"jobTitle",
	"mail",
	"mobilePhone",
	"officeLocation",
	"surname",
	"userPrincipalName",
	"assignedLicenses",
	"companyName",
	"authorizationInfo",
}

func UserToAzure(user *cloudymodels.User) *models.User {
	u := models.NewUser()
	u.SetId(&user.ID)

	u.SetUserPrincipalName(&user.UserName)
	u.SetDisplayName(&user.DisplayName)

	emailNickname := cloudy.TrimDomain(user.UserName)
	u.SetMailNickname(&emailNickname)

	if user.Email != "" {
		u.SetMail(&user.Email)
	}

	u.SetGivenName(&user.FirstName)
	u.SetSurname(&user.LastName)

	if user.Company != "" {
		u.SetCompanyName(&user.Company)
	}

	if user.JobTitle != "" {
		u.SetJobTitle(&user.JobTitle)
	}

	if user.OfficePhone != "" {
		u.SetBusinessPhones([]string{user.OfficePhone})
	}

	if user.MobilePhone != "" {
		u.SetMobilePhone(&user.MobilePhone)
	}

	if user.Department != "" {
		u.SetDepartment(&user.Department)
	}

	if user.MustChangePassword || user.Password != "" {
		profile := models.NewPasswordProfile()
		profile.SetForceChangePasswordNextSignIn(cloudy.BoolP(user.MustChangePassword))
		profile.SetPassword(&user.Password)
		u.SetPasswordProfile(profile)
	}

	return u
}

func UserToCloudy(user models.Userable) *cloudymodels.User {
	u := &cloudymodels.User{}

	u.ID = *user.GetId()
	if user.GetUserPrincipalName() != nil {
		u.UserName = *user.GetUserPrincipalName()
	}

	if user.GetGivenName() != nil {
		u.FirstName = *user.GetGivenName()
	}

	if user.GetSurname() != nil {
		u.LastName = *user.GetSurname()
	}

	if user.GetMail() != nil {
		u.Email = *user.GetMail()
	}

	if user.GetCompanyName() != nil {
		u.Company = *user.GetCompanyName()
	}

	if user.GetJobTitle() != nil {
		u.JobTitle = *user.GetJobTitle()
	}

	if user.GetDisplayName() != nil {
		u.DisplayName = *user.GetDisplayName()
	}

	if user.GetDepartment() != nil {
		u.Department = *user.GetDepartment()
	}

	if user.GetMobilePhone() != nil {
		u.MobilePhone = *user.GetMobilePhone()
	}

	if len(user.GetBusinessPhones()) >= 1 {
		u.OfficePhone = user.GetBusinessPhones()[0]
	}

	if user.GetPasswordProfile() != nil {
		if user.GetPasswordProfile().GetForceChangePasswordNextSignIn() != nil {
			u.MustChangePassword = *user.GetPasswordProfile().GetForceChangePasswordNextSignIn()
		}

		if user.GetPasswordProfile().GetPassword() != nil {
			u.Password = *user.GetPasswordProfile().GetPassword()
		}
	}

	if user.GetStreetAddress() != nil {
		u.CustomSecurityAttributes = *user.GetStreetAddress()
	}

	return u
}
