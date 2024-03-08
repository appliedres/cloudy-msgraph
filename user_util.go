package cloudymsgraph

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"strings"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/go-openapi/strfmt"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
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

	if !strings.EqualFold(user.ID, "") {
		u.SetId(&user.ID)
	}

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

	customSecurityAttributes := ParseUserCustomSecurityAttributes(user)

	if customSecurityAttributes != nil {
		u.SetCustomSecurityAttributes(customSecurityAttributes)
	}

	return u
}

func ParseUserCustomSecurityAttributes(user *cloudymodels.User) *models.CustomSecurityAttributeValue {
	hasCustomSecurityAttributes := false

	// TODO: When Microsoft fixes the bug with Custom Security Attributes this will need to be changed
	cloudyattr := make(map[string]interface{})

	if user.AccountType != "" {
		cloudyattr["AccountType"] = &user.AccountType
		hasCustomSecurityAttributes = true
	}

	if user.Citizenship != "" {
		cloudyattr["Citizenship"] = &user.Citizenship
		hasCustomSecurityAttributes = true
	}

	if user.ContractDate != "" {
		cloudyattr["ContractExpirationDate"] = &user.ContractDate
		hasCustomSecurityAttributes = true
	}

	if user.ContractNumber != "" {
		cloudyattr["ContractNumber"] = &user.ContractNumber
		hasCustomSecurityAttributes = true
	}

	if user.Organization != "" {
		cloudyattr["Organization"] = &user.Organization
		hasCustomSecurityAttributes = true
	}

	if user.Project != "" {
		cloudyattr["Project"] = &user.Project
		hasCustomSecurityAttributes = true
	}

	if user.ProgramRole != "" {
		cloudyattr["ProgramRole"] = &user.ProgramRole
		hasCustomSecurityAttributes = true
	}

	if hasCustomSecurityAttributes {

		customSecurityAttributes := models.NewCustomSecurityAttributeValue()

		odata := "#microsoft.graph.customSecurityAttributeValue"
		// odata := "#Microsoft.DirectoryServices.CustomSecurityAttributeValue"
		cloudyattr["@odata.type"] = &odata

		customSecurityAttributes.GetAdditionalData()["cloudy"] = cloudyattr

		return customSecurityAttributes
	}

	return nil
}

func UserToPatch(user *cloudymodels.User, currentUser *cloudymodels.User) *models.User {

	u := models.NewUser()
	u.SetId(&user.ID)

	if user.FirstName != currentUser.FirstName {
		u.SetGivenName(&user.FirstName)
	}

	if user.LastName != currentUser.LastName {
		u.SetSurname(&user.LastName)
	}

	if user.JobTitle != currentUser.JobTitle {
		u.SetJobTitle(&user.JobTitle)
	}

	if user.MobilePhone != currentUser.MobilePhone {
		u.SetMobilePhone(&user.MobilePhone)
	}

	if user.Department != currentUser.Department {
		u.SetDepartment(&user.Department)
	}

	customSecurityAttributes := ParseUserCustomSecurityAttributes(user)

	if customSecurityAttributes != nil {
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

	allAttributes := readAllCustomSecurityAttributes(user, "cloudy")
	if allAttributes != nil {

		// Read the Contract Number
		contractNumber, exists := allAttributes["ContractNumber"]
		if exists && contractNumber != nil {
			u.ContractNumber = *contractNumber
		}

		// Read the Contract Date
		contractDate, exists := allAttributes["ContractExpirationDate"]
		if exists && contractDate != nil {
			u.ContractDate = *contractDate
		}

		// Read the Account Type
		accountType, exists := allAttributes["AccountType"]
		if exists && accountType != nil {
			u.AccountType = *accountType
		}

		// Read the Citizenship
		citizenship, exists := allAttributes["Citizenship"]
		if exists && citizenship != nil {
			u.Citizenship = *citizenship
		}

		// Read the Citizenship
		organization, exists := allAttributes["Organization"]
		if exists && organization != nil {
			u.Organization = *organization
		}

		// Read the Citizenship
		programRole, exists := allAttributes["ProgramRole"]
		if exists && programRole != nil {
			u.ProgramRole = *programRole
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

func readAllCustomSecurityAttributes(user models.Userable, attributeSet string) map[string]*string {
	allAttributes := make(map[string]*string)

	if user.GetCustomSecurityAttributes() != nil && user.GetCustomSecurityAttributes().GetAdditionalData() != nil {
		attrs := user.GetCustomSecurityAttributes().GetAdditionalData()

		attributeSetMap := attrs[attributeSet]
		if attributeSetMap != nil && attributeSetMap.(map[string]interface{}) != nil {
			attributeSetMap := attributeSetMap.(map[string]interface{})
			for attributeName := range attributeSetMap {

				attributeValue := attributeSetMap[attributeName]
				if attributeValue != nil && attributeValue.(*string) != nil {
					allAttributes[attributeName] = attributeValue.(*string)
				} else {
					allAttributes[attributeName] = nil
				}
			}

			return allAttributes
		}
	}

	return nil
}

func readCustomAttributeStr(user models.Userable, attributeSet string, attributeName string) *string {

	allAttributes := readAllCustomSecurityAttributes(user, attributeSet)
	if allAttributes != nil {

		attributeValue, exists := allAttributes[attributeName]
		if exists && attributeValue != nil {
			return attributeValue
		}
	}

	return nil
}

func UpdateAzUser(ctx context.Context, azUser models.Userable, cUser *cloudymodels.User) {

	if azUser.GetId() == nil || !strings.EqualFold(*azUser.GetId(), cUser.ID) {
		azUser.SetId(&cUser.ID)
	}

	if azUser.GetSurname() == nil || !strings.EqualFold(*azUser.GetSurname(), cUser.FirstName) {
		if azUser.GetSurname() == nil {
			silly := ""
			azUser.SetSurname(&silly)
		}
		azUser.SetSurname(&cUser.FirstName)
	}

	if azUser.GetGivenName() == nil || !strings.EqualFold(*azUser.GetGivenName(), cUser.LastName) {
		azUser.SetGivenName(&cUser.LastName)
	}

}
