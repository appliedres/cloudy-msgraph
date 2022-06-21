package cloudymsgraph

import (
	"context"
)

type LicenseManager struct {
	graph *MSGraph
}

func (lm *LicenseManager) AssignLicense(ctx context.Context, userId string, licenseSku string) error {
	return lm.graph.AssignLicenses(ctx, userId, licenseSku)
}

func (lm *LicenseManager) RemoveLicense(ctx context.Context, userId string, licenseSku string) error {
	return lm.graph.RemoveLicenses(ctx, userId, licenseSku)
}

func (lm *LicenseManager) GetLicenses(ctx context.Context, userId string, licenseSku string) ([]string, error) {
	return nil, nil
}

func (lm *LicenseManager) GetAssigned(ctx context.Context, licenseSku string) ([]string, error) {

	return nil, nil
}
