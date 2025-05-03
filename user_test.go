package cloudymsgraph

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/appliedres/cloudy"
	cloudymodels "github.com/appliedres/cloudy/models"
	"github.com/appliedres/cloudy/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestEmailNickname1(t *testing.T) {
	check1 := GenerateMailNickname("   2John Doe #254#@@?")
	require.Equal(t, "2johndoe254", check1)
}

func TestUserToAzure(t *testing.T) {
	u := &cloudymodels.User{
		Username:  "2test@abc.onmicrosoft.com",
		Email:     "2test@abc.com",
		FirstName: "2Test",
		LastName:  "TEster",
	}

	azureUser := UserToAzure(u, "abc.onmicrosoft.com")
	require.NotEmpty(t, azureUser.GetMailNickname())
}

func newTestUM(t *testing.T) *MsGraphUserManager {
	ctx := cloudy.StartContext()
	err := cloudy.LoadEnv(".env.local")
	require.Nil(t, err)

	clientId := os.Getenv("AZ_CLIENT_ID")
	clientSecret := os.Getenv("AZ_CLIENT_SECRET")
	tenantId := os.Getenv("AZ_TENANT_ID")
	domain := os.Getenv("DEFAULT_DOMAIN")

	require.NotEmpty(t, clientId)
	require.NotEmpty(t, clientSecret)
	require.NotEmpty(t, tenantId)
	require.NotEmpty(t, domain)

	cfg := &MsGraphConfig{
		TenantID:      tenantId,
		ClientID:      clientId,
		ClientSecret:  clientSecret,
		DefaultDomain: domain,
	}
	cfg.SetInstance(&USGovernment)

	um, err := NewMsGraphUserManager(ctx, cfg)
	require.Nil(t, err)
	require.NotNil(t, um)
	return um
}

func TestNewUser2(t *testing.T) {
	ctx := cloudy.StartContext()
	um := newTestUM(t)

	u := &cloudymodels.User{
		FirstName: "Firsty",
		LastName:  "Lasty",
		Email:     "flasty@appliedres.com",
	}

	created, err := um.NewUser(ctx, u)
	require.Nil(t, err)
	require.NotNil(t, created)
	require.NotEmpty(t, created.UID)

	err = um.DeleteUser(ctx, created.UID)
	require.Nil(t, err)
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
	// _ = testutil.LoadEnv("../../arkloud-conf/arkloud.env")
	_ = testutil.LoadEnv("/home/john/arkloud/arkloud-conf/arkloud.env")

	env := cloudy.CreateCompleteEnvironment("ARKLOUD_ENV", "", "")
	cloudy.SetDefaultEnvironment(env)

	ctx := cloudy.StartContext()

	loader := MSGraphCredentialLoader{}
	cfg := loader.ReadFromEnv(env).(*MsGraphConfig)

	um, err := NewMsGraphUserManager(ctx, cfg)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	u, err := um.GetUser(ctx, "adam.dyer@collider.onmicrosoft.us")
	assert.Nil(t, err)
	assert.NotNil(t, u)

	azUser := UserToAzure(u, "collider.onmicrosoft.us")
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
	u.Attributes["ContractNumber"] = data
	u.Attributes["ContractDate"] = "Whenever"
	u.Attributes["Citizenship"] = "USA"
	u.Attributes["AccountType"] = "DOD Contractor"
	err = um.UpdateUser(ctx, u)
	assert.Nil(t, err)

	// Eventually consistent... give it 5 sec
	time.Sleep(10 * time.Second)

	u2, err := um.GetUser(ctx, "unittest@collider.onmicrosoft.us")
	assert.Nil(t, err)

	assert.Equal(t, data, u2.Attributes["ContractNumber"])
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
		Username:    "a",
		DisplayName: "b",
		FirstName:   "d",
		LastName:    "e",
		Email:       "h",
		UID:         "i",
	}

	cloudyU1.Attributes = make(map[string]string)
	cloudyU1.Attributes["Company"] = "f"
	cloudyU1.Attributes["Department"] = "g"
	cloudyU1.Attributes["JobTitle"] = "j"
	cloudyU1.Attributes["MobilePhone"] = "k"
	cloudyU1.Attributes["MustChangePassword"] = "true"
	cloudyU1.Attributes["OfficePhone"] = "l"
	cloudyU1.Attributes["Password"] = "m"

	azureU2 := models.NewUser()
	azureU2.SetId(&cloudyU1.UID)
	azureU2.SetUserPrincipalName(&cloudyU1.Username)
	azureU2.SetDisplayName(&cloudyU1.DisplayName)
	azureU2.SetMailNickname(&cloudyU1.Username)
	azureU2.SetMail(&cloudyU1.Email)
	azureU2.SetGivenName(&cloudyU1.FirstName)
	azureU2.SetSurname(&cloudyU1.LastName)

	company := cloudyU1.Attributes["Company"]
	azureU2.SetCompanyName(&company)
	jobTitle := cloudyU1.Attributes["JobTitle"]
	azureU2.SetJobTitle(&jobTitle)
	azureU2.SetBusinessPhones([]string{cloudyU1.Attributes["OfficePhone"]})
	mobilePhone := cloudyU1.Attributes["MobilePhone"]
	azureU2.SetMobilePhone(&mobilePhone)
	department := cloudyU1.Attributes["Department"]
	azureU2.SetDepartment(&department)
	passwordProfile := models.NewPasswordProfile()
	mustChangePw := true
	passwordProfile.SetForceChangePasswordNextSignIn(&mustChangePw)
	pw := cloudyU1.Attributes["Password"]
	passwordProfile.SetPassword(&pw)
	azureU2.SetPasswordProfile(passwordProfile)

	azureU1 := UserToAzure(cloudyU1, "collider.onmicrosoft.us")
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
