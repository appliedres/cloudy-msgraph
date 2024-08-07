package cloudymsgraph

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func UserToAzure(user *cloudymodels.User) *models.User {
	u := models.NewUser()

	if !strings.EqualFold(user.UID, "") {
		u.SetId(&user.UID)
	}

	u.SetUserPrincipalName(&user.Username)
	u.SetDisplayName(&user.DisplayName)

	emailNickname := cloudy.TrimDomain(user.Username)
	u.SetMailNickname(&emailNickname)

	if user.Email != "" {
		u.SetMail(&user.Email)
	}

	u.SetGivenName(&user.FirstName)
	u.SetSurname(&user.LastName)

	if user.Attributes["Company"] != "" {
		company := user.Attributes["Company"]
		u.SetCompanyName(&company)
	}

	if user.Attributes["JobTitle"] != "" {
		jobTitle := user.Attributes["JobTitle"]
		u.SetJobTitle(&jobTitle)
	}

	if user.Attributes["OfficePhone"] != "" {
		u.SetBusinessPhones([]string{user.Attributes["OfficePhone"]})
	}

	if user.Attributes["MobilePhone"] != "" {
		mobilePhone := user.Attributes["MobilePhone"]
		u.SetMobilePhone(&mobilePhone)
	}

	if user.Attributes["Department"] != "" {
		dept := user.Attributes["Department"]
		u.SetDepartment(&dept)
	}

	if user.Attributes["MustChangePassword"] == "true" || user.Attributes["Password"] != "" {
		profile := models.NewPasswordProfile()
		mustChangePw := true
		profile.SetForceChangePasswordNextSignIn(&mustChangePw)
		pw := user.Attributes["Password"]
		profile.SetPassword(&pw)
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

	if user.Attributes["AccountType"] != "" {
		acctType := user.Attributes["AccountType"]
		cloudyattr["AccountType"] = &acctType
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["Citizenship"] != "" {
		citizenship := user.Attributes["Citizenship"]
		cloudyattr["Citizenship"] = &citizenship
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["ContractDate"] != "" {
		contractDate := user.Attributes["ContractDate"]
		cloudyattr["ContractExpirationDate"] = &contractDate
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["ContractNumber"] != "" {
		contractNumber := user.Attributes["ContractNumber"]
		cloudyattr["ContractNumber"] = &contractNumber
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["Organization"] != "" {
		organization := user.Attributes["Organization"]
		cloudyattr["Organization"] = &organization
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["Project"] != "" {
		project := user.Attributes["Project"]
		cloudyattr["Project"] = &project
		hasCustomSecurityAttributes = true
	}

	if user.Attributes["ProgramRole"] != "" {
		programRole := user.Attributes["ProgramRole"]
		cloudyattr["ProgramRole"] = &programRole
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
	u.SetId(&user.UID)

	if user.FirstName != currentUser.FirstName {
		u.SetGivenName(&user.FirstName)
	}

	if user.LastName != currentUser.LastName {
		u.SetSurname(&user.LastName)
	}

	if user.Attributes["JobTitle"] != currentUser.Attributes["JobTitle"] {
		jobTitle := user.Attributes["JobTitle"]
		u.SetJobTitle(&jobTitle)
	}

	if user.Attributes["MobilePhone"] != currentUser.Attributes["MobilePhone"] {
		mobilePhone := user.Attributes["MobilePhone"]
		u.SetMobilePhone(&mobilePhone)
	}

	if user.Attributes["Department"] != currentUser.Attributes["Department"] {
		department := user.Attributes["Department"]
		u.SetDepartment(&department)
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
		u.UID = *user.GetId()
	}

	if user.GetUserPrincipalName() != nil {
		u.Username = *user.GetUserPrincipalName()
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

	if user.GetDisplayName() != nil {
		u.DisplayName = *user.GetDisplayName()
	}

	if user.GetAccountEnabled() != nil {
		u.Enabled = *user.GetAccountEnabled()
	}

	u.Attributes = make(map[string]string)
	if user.GetCompanyName() != nil {
		u.Attributes["Company"] = *user.GetCompanyName()
	}

	if user.GetJobTitle() != nil {
		u.Attributes["JobTitle"] = *user.GetJobTitle()
	}

	if user.GetDepartment() != nil {
		u.Attributes["Department"] = *user.GetDepartment()
	}

	if user.GetMobilePhone() != nil {
		u.Attributes["MobilePhone"] = *user.GetMobilePhone()
	}

	if len(user.GetBusinessPhones()) >= 1 {
		u.Attributes["OfficePhone"] = user.GetBusinessPhones()[0]
	}

	if user.GetSignInActivity() != nil && user.GetSignInActivity().GetLastSignInDateTime() != nil {
		u.Attributes["LastSignInDate"] = user.GetSignInActivity().GetLastSignInDateTime().String()
	}

	if user.GetPasswordProfile() != nil {
		if user.GetPasswordProfile().GetForceChangePasswordNextSignIn() != nil {
			u.Attributes["MustChangePassword"] = strconv.FormatBool(*user.GetPasswordProfile().GetForceChangePasswordNextSignIn())
		}

		if user.GetPasswordProfile().GetPassword() != nil {
			u.Attributes["Password"] = *user.GetPasswordProfile().GetPassword()
		}
	}

	allAttributes := readAllCustomSecurityAttributes(user, "cloudy")
	if allAttributes != nil {

		// Read the Contract Number
		contractNumber, exists := allAttributes["ContractNumber"]
		if exists && contractNumber != nil {
			u.Attributes["ContractNumber"] = *contractNumber
		}

		// Read the Contract Date
		contractDate, exists := allAttributes["ContractExpirationDate"]
		if exists && contractDate != nil {
			u.Attributes["ContractDate"] = *contractDate
		}

		// Read the Account Type
		accountType, exists := allAttributes["AccountType"]
		if exists && accountType != nil {
			u.Attributes["AccountType"] = *accountType
		}

		// Read the Citizenship
		citizenship, exists := allAttributes["Citizenship"]
		if exists && citizenship != nil {
			u.Attributes["Citizenship"] = *citizenship
		}

		// Read the Citizenship
		organization, exists := allAttributes["Organization"]
		if exists && organization != nil {
			u.Attributes["Organization"] = *organization
		}

		// Read the Citizenship
		programRole, exists := allAttributes["ProgramRole"]
		if exists && programRole != nil {
			u.Attributes["ProgramRole"] = *programRole
		}

	} else if user.GetStreetAddress() != nil {
		// TODO: When Microsoft fixes the bug with Custom Security Attributes this will need to be changed to user.GetCustomSecurityAttributes and tested
		// also change cloudy user model CustomSecurityAttributes from string to object and implement interface

		sDec, _ := b64.StdEncoding.DecodeString(*user.GetStreetAddress())
		csa := UserCustomSecurityAttributes{}
		json.Unmarshal(sDec, &csa)

		u.Attributes["AccountType"] = csa.AccountType
		u.Attributes["Citizenship"] = csa.Citizenship
		u.Attributes["ContractNumber"] = csa.ContractNumber
		u.Attributes["ContractDate"] = csa.ContractExpirationDate
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

	if azUser.GetId() == nil || !strings.EqualFold(*azUser.GetId(), cUser.UID) {
		azUser.SetId(&cUser.UID)
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
