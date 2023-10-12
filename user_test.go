package cloudymsgraph

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/appliedres/cloudy/testutil"
	"github.com/stretchr/testify/assert"

	_ "github.com/appliedres/cloudy-azure"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func TestUserManager(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")
	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	msgraphCreds := env.LoadCredentials("MSGRAPH")
	um, err := cloudy.UserProviders.NewFromEnv(env.SegmentWithCreds(msgraphCreds, "USER"), "DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the user manager. %v", err)
	}

	testutil.TestUserManager(t, um)

}

func TestGetUser(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	ctx := cloudy.StartContext()

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	msgraphCreds := env.LoadCredentials("MSGRAPH")
	um, err := cloudy.UserProviders.NewFromEnv(env.SegmentWithCreds(msgraphCreds, "USER"), "DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the user manager. %v", err)
	}
	u, err := um.GetUser(ctx, "test.user@collider.onmicrosoft.us")
	assert.Nil(t, err)
	assert.NotNil(t, u)
}

func TestGetUserProfilePicture(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	ctx := cloudy.StartContext()

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	msgraphCreds := env.LoadCredentials("MSGRAPH")
	um, err := cloudy.UserProviders.NewFromEnv(env.SegmentWithCreds(msgraphCreds, "USER"), "DRIVER")
	if err != nil {
		log.Fatalf("Could not instantiate the user manager. %v", err)
	}

	uid := "test.user@collider.onmicrosoft.us"

	pic, err := um.(*MsGraphUserManager).GetProfilePicture(ctx, uid)
	assert.Nil(t, err)
	assert.NotNil(t, pic)
}

func TestGetUserByEmail(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")
	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "USERAPI_PREFIX", "USER_API")
	cloudy.SetDefaultEnvironment(env)

	msgraphCreds := env.LoadCredentials("MSGRAPH")
	um, err := cloudy.UserProviders.NewFromEnv(env.SegmentWithCreds(msgraphCreds, "USER"), "DRIVER")
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	ctx := cloudy.StartContext()

	u, err := um.GetUserByEmail(ctx, "unittest@collider.onmicrosoft.us",
		&cloudy.UserOptions{IncludeLastSignIn: cloudy.BoolP(true)})
	assert.Nil(t, err)
	assert.NotNil(t, u)

}

func TestGetUserToAzure(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := testutil.CreateTestEnvironment()
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	u, err := um.GetUser(ctx, "unittest@collider.onmicrosoft.us")
	assert.Nil(t, err)
	assert.NotNil(t, u)

	azUser := UserToAzure(u)
	assert.NotNil(t, azUser)

}

func TestGetUserWithCustomSecurityAttributes(t *testing.T) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := testutil.CreateTestEnvironment()
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	u, err := um.GetUser(ctx, "unittest@collider.onmicrosoft.us")
	assert.Nil(t, err)
	assert.NotNil(t, u)

}

func TestUpdateUser(t *testing.T) {
	ctx, um := testUM()

	u, err := um.GetUser(ctx, "unittest@collider.onmicrosoft.us")
	assert.Nil(t, err)

	data := time.Now().Format(time.RFC1123Z)
	u.ContractNumber = data
	u.ContractDate = "Whenever"
	u.Citizenship = "USA"
	u.AccountType = "DOD Contractor"
	err = um.UpdateUser(ctx, u)
	assert.Nil(t, err)

	// Eventually consistent... give it 5 sec
	time.Sleep(10 * time.Second)

	u2, err := um.GetUser(ctx, "unittest@collider.onmicrosoft.us")
	assert.Nil(t, err)

	assert.Equal(t, data, u2.ContractNumber)
}

func testUM() (context.Context, *MsGraphUserManager) {
	_ = testutil.LoadEnv("../arkloud-conf/arkloud.env")

	env := testutil.CreateTestEnvironment()
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	return ctx, um
}

func TestUserModel(t *testing.T) {
	cloudyU1 := &cloudymodels.User{
		UPN:                "a",
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
	}

	azureU2 := models.NewUser()
	azureU2.SetId(&cloudyU1.ID)
	azureU2.SetUserPrincipalName(&cloudyU1.UPN)
	azureU2.SetDisplayName(&cloudyU1.DisplayName)
	azureU2.SetMailNickname(&cloudyU1.UPN)
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
