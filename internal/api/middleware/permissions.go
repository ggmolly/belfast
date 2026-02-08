package middleware

import (
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/orm"
)

const authzCacheKey = "authz.effective"

func effectivePermissions(ctx iris.Context) (map[string]authz.Capability, error) {
	if cached := ctx.Values().Get(authzCacheKey); cached != nil {
		if perms, ok := cached.(map[string]authz.Capability); ok {
			return perms, nil
		}
	}
	account, ok := GetAccount(ctx)
	if !ok {
		return map[string]authz.Capability{}, nil
	}
	perms, err := orm.LoadEffectivePermissions(account.ID)
	if err != nil {
		return nil, err
	}
	ctx.Values().Set(authzCacheKey, perms)
	return perms, nil
}

func RequirePermission(key string, op authz.Operation) iris.Handler {
	return func(ctx iris.Context) {
		ctx.Values().Set(authzKeyKey, key)
		ctx.Values().Set(authzOpKey, string(op))
		if IsAuthDisabled(ctx) {
			ctx.Next()
			return
		}
		_, ok := GetAccount(ctx)
		if !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		perms, err := effectivePermissions(ctx)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load permissions", nil))
			return
		}
		cap := perms[key]
		if !cap.Allowed(op) {
			ctx.StatusCode(iris.StatusForbidden)
			_ = ctx.JSON(response.Error("permissions.denied", "permission denied", nil))
			return
		}
		ctx.Next()
	}
}

func RequirePermissionAny(key string) iris.Handler {
	return func(ctx iris.Context) {
		op := authz.OperationForMethod(ctx.Method(), authz.ReadAny, authz.WriteAny)
		RequirePermission(key, op)(ctx)
	}
}

func RequirePermissionSelf(key string) iris.Handler {
	return func(ctx iris.Context) {
		op := authz.OperationForMethod(ctx.Method(), authz.ReadSelf, authz.WriteSelf)
		RequirePermission(key, op)(ctx)
	}
}

func RequirePermissionForMethod(key string, readOp authz.Operation, writeOp authz.Operation) iris.Handler {
	return func(ctx iris.Context) {
		if ctx.Method() == http.MethodOptions {
			ctx.Next()
			return
		}
		op := authz.OperationForMethod(ctx.Method(), readOp, writeOp)
		RequirePermission(key, op)(ctx)
	}
}

// RequirePermissionAnyOrSelf enforces:
// - {read_any,write_any} when the request is not scoped to the caller's commander
// - {read_self,write_self} when the path param {id} matches the caller's CommanderID
//
// If the route does not include an {id} param, it always requires *_any.
func RequirePermissionAnyOrSelf(key string) iris.Handler {
	return func(ctx iris.Context) {
		ctx.Values().Set(authzKeyKey, key)
		if ctx.Method() == http.MethodOptions {
			ctx.Next()
			return
		}
		if IsAuthDisabled(ctx) {
			ctx.Next()
			return
		}
		account, ok := GetAccount(ctx)
		if !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		perms, err := effectivePermissions(ctx)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load permissions", nil))
			return
		}
		cap := perms[key]

		idParam := ctx.Params().Get("id")
		isSelf := false
		if idParam != "" && account.CommanderID != nil {
			parsed, err := strconv.ParseUint(idParam, 10, 32)
			if err != nil {
				ctx.StatusCode(iris.StatusBadRequest)
				_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
				return
			}
			isSelf = uint32(parsed) == *account.CommanderID
		}

		var op authz.Operation
		if isSelf {
			op = authz.OperationForMethod(ctx.Method(), authz.ReadSelf, authz.WriteSelf)
		} else {
			op = authz.OperationForMethod(ctx.Method(), authz.ReadAny, authz.WriteAny)
		}
		ctx.Values().Set(authzOpKey, string(op))

		if !cap.Allowed(op) {
			ctx.StatusCode(iris.StatusForbidden)
			_ = ctx.JSON(response.Error("permissions.denied", "permission denied", nil))
			return
		}
		ctx.Next()
	}
}
