package restapi

import (
	"net/http"

	identity_controllers "github.com/cjlapao/common-go-identity/controllers"
	"github.com/cjlapao/common-go-identity/database"
	"github.com/cjlapao/common-go-identity/interfaces"
	"github.com/cjlapao/common-go-identity/middleware"
	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/cjlapao/common-go/execution_context"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func (l *HttpListener) AddAuthorizedController(c controllers.Controller, path string, methods ...string) {
	l.Controllers = append(l.Controllers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}
	adapters := make([]controllers.Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)
	adapters = append(adapters, middleware.TokenAuthorizationMiddlewareAdapter([]string{}, []string{}))

	if l.Options.ApiPrefix != "" {
		path = http_helper.JoinUrl(l.Options.ApiPrefix, path)
	}

	subRouter.HandleFunc(path,
		controllers.Adapt(
			http.HandlerFunc(c),
			adapters...).ServeHTTP)
}

func (l *HttpListener) WithDefaultAuthentication() *HttpListener {
	context := database.NewMemoryUserAdapter()
	return l.WithAuthentication(context)
}

func (l *HttpListener) WithAuthentication(context interfaces.UserContextAdapter) *HttpListener {
	ctx := execution_context.Get()
	if ctx.Authorization != nil {
		defaultAuthControllers := identity_controllers.NewAuthorizationControllers(context)

		l.AddController(defaultAuthControllers.Token(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "token"), "POST")
		l.AddController(defaultAuthControllers.Token(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "token"), "POST")
		l.AddController(defaultAuthControllers.Introspection(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "token", "introspect"), "POST")
		l.AddController(defaultAuthControllers.Introspection(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "{tenantId}", "token", "introspect"), "POST")
		l.AddAuthorizedControllerWithRoles(defaultAuthControllers.Register(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "register"), []string{"_su,_admin"}, "POST")
		l.AddAuthorizedControllerWithRoles(defaultAuthControllers.Register(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "{tenantId}", "register"), []string{"_su,_admin"}, "POST")
		l.AddAuthorizedControllerWithRoles(defaultAuthControllers.Revoke(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "revoke"), []string{"_su,_admin"}, "POST")
		l.AddAuthorizedControllerWithRoles(defaultAuthControllers.Revoke(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "{tenantId}", "revoke"), []string{"_su,_admin"}, "POST")

		l.AddController(defaultAuthControllers.Configuration(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, ".well-known", "openid-configuration"), "GET")
		l.AddController(defaultAuthControllers.Configuration(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "{tenantId}", ".well-known", "openid-configuration"), "GET")
		l.AddController(defaultAuthControllers.Jwks(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, ".well-known", "openid-configuration", "jwks"), "GET")
		l.AddController(defaultAuthControllers.Jwks(), http_helper.JoinUrl(ctx.Authorization.Options.ControllerPrefix, "{tenantId}", ".well-known", "openid-configuration", "jwks"), "GET")
		l.DefaultAdapters = append([]controllers.Adapter{middleware.EndAuthorizationMiddlewareAdapter()}, l.DefaultAdapters...)
		l.Options.EnableAuthentication = true
	} else {
		l.Logger.Error("No authorization context found, ignoring")
	}
	return l
}

func (l *HttpListener) AddAuthorizedControllerWithRoles(c controllers.Controller, path string, roles []string, methods ...string) {
	l.AddAuthorizedControllerWithRolesAndClaims(c, path, roles, []string{}, methods...)
}

func (l *HttpListener) AddAuthorizedControllerWithClaims(c controllers.Controller, path string, claims []string, methods ...string) {
	l.AddAuthorizedControllerWithRolesAndClaims(c, path, []string{}, claims, methods...)
}

func (l *HttpListener) AddAuthorizedControllerWithRolesAndClaims(c controllers.Controller, path string, roles []string, claims []string, methods ...string) {
	l.Controllers = append(l.Controllers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}
	adapters := make([]controllers.Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)
	adapters = append(adapters, middleware.TokenAuthorizationMiddlewareAdapter(roles, claims))

	if l.Options.ApiPrefix != "" {
		path = http_helper.JoinUrl(l.Options.ApiPrefix, path)
	}

	subRouter.HandleFunc(path,
		controllers.Adapt(
			http.HandlerFunc(c),
			adapters...).ServeHTTP)
}
