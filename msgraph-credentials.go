package cloudymsgraph

import "github.com/appliedres/cloudy"

func init() {
	cloudy.CredentialSources[MSGraphCredentialsKey] = &MSGraphCredentialLoader{}
}

const MSGraphCredentialsKey = "msgraph"

type MSGraphCredentialLoader struct{}

func (loader *MSGraphCredentialLoader) ReadFromEnv(env *cloudy.Environment) interface{} {
	cfg := &MsGraphConfig{}

	cfg.TenantID = env.Force("AZ_TENANT_ID")
	cfg.ClientID = env.Force("AZ_CLIENT_ID")
	cfg.ClientSecret = env.Force("AZ_CLIENT_SECRET")
	cfg.Region = env.Default("AZ_REGION", "usgovvirginia")
	cfg.APIBase = env.Default("AZ_API_BASE", "https://graph.microsoft.us/v1.0")

	if cfg.TenantID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil
	}

	return cfg
}
