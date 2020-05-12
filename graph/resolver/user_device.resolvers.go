// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
package resolver

import (
	"context"
	"strings"

	"github.com/nedrocks/delphisbe/graph/generated"
	"github.com/nedrocks/delphisbe/graph/model"
)

func (r *userDeviceResolver) Platform(ctx context.Context, obj *model.UserDevice) (model.Platform, error) {
	switch strings.ToLower(obj.Platform) {
	case "ios":
		return model.PlatformIos, nil
	case "android":
		return model.PlatformAndroid, nil
	case "web":
		return model.PlatformWeb, nil
	default:
		return model.PlatformUnknown, nil
	}
}

func (r *Resolver) UserDevice() generated.UserDeviceResolver { return &userDeviceResolver{r} }

type userDeviceResolver struct{ *Resolver }