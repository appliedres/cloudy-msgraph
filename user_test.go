package cloudymsgraph

import (
	"log"
	"testing"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/appliedres/cloudy/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func TestUserManager(t *testing.T) {
	ctx := cloudy.StartContext()

	testutil.LoadEnv("test.env")
	tenantID := cloudy.ForceEnv("TenantID", "")
	ClientID := cloudy.ForceEnv("ClientID", "")
	ClientSecret := cloudy.ForceEnv("ClientSecret", "")

	cfg := &MSGraphConfig{
		TenantID:     tenantID,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	}
	cfg.SetInstance(&USGovernment)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	testutil.TestUserManager(t, um)

}

func TestUserModel(t *testing.T) {
	cloudyU1 := &cloudymodels.User{
		UserName:           "a",
		DisplayName:        "b",
		FirstName:          "d",
		LastName:           "e",
		Company:            "f",
		Department:         "g",
		Email:              "h",
		ID:                 "i",
		JobTitle:           "j",
		MobilePhone:        "k",
		MustChangePassword: true,
		OfficePhone:        "l",
		Password:           "m",
		Status:             "",
	}

	azureU2 := models.NewUser()
	azureU2.SetId(&cloudyU1.ID)
	azureU2.SetUserPrincipalName(&cloudyU1.UserName)
	azureU2.SetDisplayName(&cloudyU1.DisplayName)
	azureU2.SetMailNickname(&cloudyU1.UserName)
	azureU2.SetMail(&cloudyU1.Email)
	azureU2.SetGivenName(&cloudyU1.FirstName)
	azureU2.SetSurname(&cloudyU1.LastName)
	azureU2.SetCompanyName(&cloudyU1.Company)
	azureU2.SetJobTitle(&cloudyU1.JobTitle)
	azureU2.SetBusinessPhones([]string{cloudyU1.OfficePhone})
	azureU2.SetMobilePhone(&cloudyU1.MobilePhone)
	azureU2.SetDepartment(&cloudyU1.Department)
	passwordProfile := models.NewPasswordProfile()
	passwordProfile.SetForceChangePasswordNextSignIn(cloudy.BoolP(cloudyU1.MustChangePassword))
	passwordProfile.SetPassword(&cloudyU1.Password)
	azureU2.SetPasswordProfile(passwordProfile)

	azureU1 := UserToAzure(cloudyU1)
	assert.Equal(t, azureU1.GetId(), azureU2.GetId())
	assert.Equal(t, azureU1.GetUserPrincipalName(), azureU2.GetUserPrincipalName())
	assert.Equal(t, azureU1.GetDisplayName(), azureU2.GetDisplayName())
	assert.Equal(t, azureU1.GetMailNickname(), azureU2.GetMailNickname())
	assert.Equal(t, azureU1.GetGivenName(), azureU2.GetGivenName())
	assert.Equal(t, azureU1.GetSurname(), azureU2.GetSurname())
	assert.Equal(t, azureU1.GetCompanyName(), azureU2.GetCompanyName())
	assert.Equal(t, azureU1.GetJobTitle(), azureU2.GetJobTitle())
	assert.Equal(t, azureU1.GetBusinessPhones()[0], azureU2.GetBusinessPhones()[0])
	assert.Equal(t, azureU1.GetMobilePhone(), azureU2.GetMobilePhone())
	assert.Equal(t, azureU1.GetDepartment(), azureU2.GetDepartment())

	cloudyU2 := UserToCloudy(azureU2)
	assert.Equal(t, cloudyU1, cloudyU2)

}
