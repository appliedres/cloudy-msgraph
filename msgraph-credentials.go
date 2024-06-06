package cloudymsgraph

import "github.com/appliedres/cloudy"

func init() {
	cloudy.CredentialSources[MSGraphCredentialsKey] = &MSGraphCredentialLoader{}
}

const MSGraphCredentialsKey = "msgraph"

type MSGraphCredentialLoader struct{}

func (loader *MSGraphCredentialLoader) ReadFromEnvMgr(em *cloudy.EnvManager) interface{} {
	cfg := &MsGraphConfig{}

	cfg.TenantID = em.GetVar("MSGRAPH_AZ_TENANT_ID")
	cfg.ClientID = em.GetVar("MSGRAPH_AZ_CLIENT_ID")
	cfg.ClientSecret = em.GetVar("MSGRAPH_AZ_CLIENT_SECRET")
	cfg.Region = em.GetVar("MSGRAPH_AZ_REGION")
	cfg.APIBase = em.GetVar("MSGRAPH_AZ_API_BASE")

	if cfg.TenantID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil
	}

	return cfg
}
