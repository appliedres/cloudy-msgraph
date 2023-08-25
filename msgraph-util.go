package cloudymsgraph

import (
	"context"

	"github.com/appliedres/cloudy"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models/odataerrors"
)

const ResourceNotFoundCode = "Request_ResourceNotFound"
const ImageNotFoundCode = "ImageNotFound"

func GetErrorCodeAndMessage(ctx context.Context, err error) (string, string) {
	oDataErr, ok := err.(*odataerrors.ODataError)

	if !ok {
		_ = cloudy.Error(ctx, "err is wrong type %v", err)
		return "", ""
	}

	errEscaped := oDataErr.GetErrorEscaped()

	return *errEscaped.GetCode(), *errEscaped.GetMessage()
}
