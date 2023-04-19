package cloudymsgraph

import (
	b64 "encoding/base64"
	"encoding/json"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/go-openapi/strfmt"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

var DefaultUserSelectFields = []string{
	"accountEnabled",
	"customSecurityAttributes",
	// "signInActivity",
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

var SigninActivityField = "signInActivity"

type UserCustomSecurityAttributes struct {
	AccountType            string `json:"AccountType,omitempty"`
	Citizenship            string `json:"Citizenship,omitempty"`
	ContractNumber         string `json:"ContractNumber,omitempty"`
	ContractExpirationDate string `json:"ContractExpirationDate,omitempty"`
	Justification          string `json:"Justification,omitempty"`
	ProgramRole            string `json:"ProgramRole,omitempty"`
	Sponsor                string `json:"Sponsor,omitempty"`
	StatusReason           string `json:"StatusReason,omitempty"`
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
	if user.AccountType != "" || user.Citizenship != "" || user.ContractDate != "" || user.ContractNumber != "" {
		cloudyattr := make(map[string]interface{})

		odata := "#Microsoft.DirectoryServices.CustomSecurityAttributeValue"
		cloudyattr["@odata.type"] = &odata
		if user.AccountType != "" {
			cloudyattr["AccountType"] = &user.AccountType
		}
		if user.Citizenship != "" {
			cloudyattr["Citizenship"] = &user.Citizenship
		}
		if user.ContractDate != "" {
			cloudyattr["ContractExpirationDate"] = &user.ContractDate
		}
		if user.ContractNumber != "" {
			cloudyattr["ContractNumber"] = &user.ContractNumber
		}

		customSecurityAttributes := models.NewCustomSecurityAttributeValue()
		customSecurityAttributes.GetAdditionalData()["cloudy"] = cloudyattr
		u.SetCustomSecurityAttributes(customSecurityAttributes)
	}

	return u
}

func UserToCloudy(user models.Userable) *cloudymodels.User {
	u := &cloudymodels.User{}

	if user.GetId() != nil {
		u.ID = *user.GetId()
	}

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

	if user.GetAccountEnabled() != nil {
		u.Enabled = *user.GetAccountEnabled()
	}

	if user.GetSignInActivity() != nil && user.GetSignInActivity().GetLastSignInDateTime() != nil {
		lastSignIn := *user.GetSignInActivity().GetLastSignInDateTime()
		u.LastSignInDate = strfmt.DateTime(lastSignIn)
	}

	if user.GetPasswordProfile() != nil {
		if user.GetPasswordProfile().GetForceChangePasswordNextSignIn() != nil {
			u.MustChangePassword = *user.GetPasswordProfile().GetForceChangePasswordNextSignIn()
		}

		if user.GetPasswordProfile().GetPassword() != nil {
			u.Password = *user.GetPasswordProfile().GetPassword()
		}
	}

	if user.GetCustomSecurityAttributes() != nil && user.GetCustomSecurityAttributes().GetAdditionalData() != nil {

		// Read the Contract Number
		contractNumber := readCustomAttributeStr(user, "cloudy", "ContractNumber")
		if contractNumber != nil {
			u.ContractNumber = *contractNumber
		}

		// Read the Contract Date
		contractDate := readCustomAttributeStr(user, "cloudy", "ContractExpirationDate")
		if contractDate != nil {
			u.ContractDate = *contractDate
		}

		// Read the Account Type
		acctType := readCustomAttributeStr(user, "cloudy", "AccountType")
		if acctType != nil {
			u.AccountType = *acctType
		}

		// Read the Citizenship
		citizenship := readCustomAttributeStr(user, "cloudy", "Citizenship")
		if citizenship != nil {
			u.Citizenship = *citizenship
		}

	} else if user.GetStreetAddress() != nil {
		// TODO: When Microsoft fixes the bug with Custom Security Attributes this will need to be changed to user.GetCustomSecurityAttributes and tested
		// also change cloudy user model CustomSecurityAttributes from string to object and implement interface

		sDec, _ := b64.StdEncoding.DecodeString(*user.GetStreetAddress())
		csa := UserCustomSecurityAttributes{}
		json.Unmarshal(sDec, &csa)

		u.AccountType = csa.AccountType
		u.Citizenship = csa.Citizenship
		u.ContractNumber = csa.ContractNumber
		u.ContractDate = csa.ContractExpirationDate
	}

	return u
}

func readCustomAttributeStr(user models.Userable, category string, attrName string) *string {
	if user.GetCustomSecurityAttributes() != nil && user.GetCustomSecurityAttributes().GetAdditionalData() != nil {
		attr := user.GetCustomSecurityAttributes().GetAdditionalData()
		attrCatMap := attr[category]
		if attrCatMap != nil && attrCatMap.(map[string]interface{}) != nil {
			cloudyAttrMap := attrCatMap.(map[string]interface{})

			val := cloudyAttrMap[attrName]
			if val != nil && val.(*string) != nil {
				return val.(*string)
			}
		}
	}
	return nil
}
