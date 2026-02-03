package handlers

import (
	"errors"
	"sort"
	"strings"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/orm"
)

const (
	userPermissionResourcesRead   = "self.resources.read"
	userPermissionResourcesUpdate = "self.resources.update"
	userPermissionShipsGive       = "self.ships.give"
	userPermissionItemsGive       = "self.items.give"
	userPermissionSkinsGive       = "self.skins.give"
)

var userPermissionActions = []string{
	userPermissionResourcesRead,
	userPermissionResourcesUpdate,
	userPermissionShipsGive,
	userPermissionItemsGive,
	userPermissionSkinsGive,
}

var userPermissionActionSet = func() map[string]struct{} {
	lookup := make(map[string]struct{}, len(userPermissionActions))
	for _, action := range userPermissionActions {
		lookup[action] = struct{}{}
	}
	return lookup
}()

func requireUserPermission(ctx iris.Context, action string) bool {
	allowed, err := orm.UserPermissionAllowed(action)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load permission policy", nil))
		return false
	}
	if !allowed {
		ctx.StatusCode(iris.StatusForbidden)
		_ = ctx.JSON(response.Error("permissions.denied", "permission denied", nil))
		return false
	}
	return true
}

func normalizeUserPermissionActions(actions []string) ([]string, error) {
	seen := map[string]struct{}{}
	output := make([]string, 0, len(actions))
	for _, action := range actions {
		trimmed := strings.TrimSpace(action)
		if trimmed == "" {
			return nil, errors.New("permission action required")
		}
		if _, ok := userPermissionActionSet[trimmed]; !ok {
			return nil, errors.New("unknown permission action")
		}
		if _, ok := seen[trimmed]; ok {
			return nil, errors.New("duplicate permission action")
		}
		seen[trimmed] = struct{}{}
		output = append(output, trimmed)
	}
	sort.Strings(output)
	return output, nil
}
