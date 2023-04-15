package cloudymsgraph

import (
	b64 "encoding/base64"
	"encoding/json"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
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
	"streetAddress",
}

type UserCustomSecurityAttributes struct {
	AccountStatus          string `json:"AccountStatus"` // approved, pending, rejected, offboarded
	AccountType            string `json:"AccountType"`
	Citizenship            string `json:"Citizenship"`
	ContractNumber         string `json:"ContractNumber"`
	ContractExpirationDate string `json:"ContractExpirationDate"`
	Justification          string `json:"Justification"`
	ProgramRole            string `json:"ProgramRole"`
	Sponsor                string `json:"Sponsor"`
	StatusReason           string `json:"StatusReason"`
}

func UserToAzure(user *cloudymodels.User) *models.User {
	u := models.NewUser()
	u.SetId(&user.ID)

	u.SetUserPrincipalName(&user.UPN)
	u.SetDisplayName(&user.DisplayName)

	emailNickname := cloudy.TrimDomain(user.UPN)
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

	// TODO: When Microsoft fixes the bug with Custom Security Attributes this will need to be changed
	if user.AccountType != "" || user.Citizenship != "" || user.ContractDate != "" || user.ContractNumber != "" || user.Justification != "" ||
		user.ProgramRole != "" || user.Sponsor != "" || user.Status != "" {
		csa := &UserCustomSecurityAttributes{
			AccountStatus:          user.AccountStatus,
			AccountType:            user.AccountType,
			Citizenship:            user.Citizenship,
			ContractExpirationDate: user.ContractDate,
			ContractNumber:         user.ContractNumber,
			Justification:          user.Justification,
			Sponsor:                user.Sponsor,
			StatusReason:           user.Status,
		}
		jsonStr, err := json.Marshal(&csa)
		if err == nil {
			sEnc := b64.StdEncoding.EncodeToString(jsonStr)
			u.SetStreetAddress(&sEnc)
		}
	}

	return u
}

func UserToCloudy(user models.Userable) *cloudymodels.User {
	u := &cloudymodels.User{}

	u.ID = *user.GetId()
	if user.GetUserPrincipalName() != nil {
		u.UPN = *user.GetUserPrincipalName()
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

	// TODO: When Microsoft fixes the bug with Custom Security Attributes this will need to be changed to user.GetCustomSecurityAttributes and tested
	// also change cloudy user model CustomSecurityAttributes from string to object and implement interface
	if user.GetStreetAddress() != nil {
		sDec, _ := b64.StdEncoding.DecodeString(*user.GetStreetAddress())
		csa := UserCustomSecurityAttributes{}
		json.Unmarshal(sDec, &csa)

		u.AccountType = csa.AccountType
		u.Citizenship = csa.Citizenship
		u.ContractNumber = csa.ContractNumber
		u.ContractDate = csa.ContractExpirationDate
		u.Justification = csa.Justification
		u.Sponsor = csa.Sponsor
		u.Status = csa.StatusReason
	}

	return u
}
