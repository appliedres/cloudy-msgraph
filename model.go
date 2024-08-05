package cloudymsgraph

var DefaultUserSelectFields = []string{
	"accountEnabled",
	"displayName",
	"givenName",
	"id",
	"mail",
	"surname",
	"userPrincipalName",
}

var AdditionalAttributes = []string{
	"customSecurityAttributes",
	"businessPhones",
	"jobTitle",
	"mobilePhone",
	"officeLocation",
	"assignedLicenses",
	"companyName",
	"authorizationInfo",
	"streetAddress",
	// "signInActivity",
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
