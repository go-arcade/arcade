package router

import (
	http2 "net/http"

	identitymodel "github.com/go-arcade/arcade/internal/engine/model/identity_integration"
	identityservice "github.com/go-arcade/arcade/internal/engine/service/identity_integration"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rt *Router) identityIntegrationRouter(r fiber.Router, auth fiber.Handler) {
	identityGroup := r.Group("/identity")
	{
		// Provider resource management (authentication required)
		identityGroup.Get("/providers", auth, rt.listProviders)               // GET /identity/providers - list all providers (supports ?type=xxx filter)
		identityGroup.Post("/providers", auth, rt.createProvider)             // POST /identity/providers - create provider
		identityGroup.Get("/providers/types", auth, rt.listProviderTypes)     // GET /identity/providers/types - list all provider types
		identityGroup.Get("/providers/:name", auth, rt.getProvider)           // GET /identity/providers/:name - get specific provider
		identityGroup.Put("/providers/:name", auth, rt.updateProvider)        // PUT /identity/providers/:name - update provider
		identityGroup.Put("/providers/:name/toggle", auth, rt.toggleProvider) // PUT /identity/providers/:name/toggle - toggle enabled status
		identityGroup.Delete("/providers/:name", auth, rt.deleteProvider)     // DELETE /identity/providers/:name - delete provider

		// Authentication flow (no authentication required)
		identityGroup.Get("/authorize/:provider", rt.authorize)   // GET /identity/authorize/:provider - initiate authorization (OAuth/OIDC)
		identityGroup.Get("/callback/:provider", rt.callback)     // GET /identity/callback/:provider - authorization callback
		identityGroup.Post("/ldap/login/:provider", rt.ldapLogin) // POST /identity/ldap/login/:provider - LDAP login
	}
}

// authorize initiates authorization (OAuth/OIDC)
func (rt *Router) authorize(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	url, err := identityService.Authorize(providerName)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	return c.Redirect(url, http2.StatusTemporaryRedirect)
}

// callback handles OAuth/OIDC authorization callback
func (rt *Router) callback(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	providerName := c.Params("provider")
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" || providerName == "" {
		return http.WithRepErrMsg(c, http.InvalidStatusParameter.Code, http.InvalidStatusParameter.Msg, c.Path())
	}

	userInfo, err := identityService.Callback(providerName, state, code)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, userInfo)
	return nil
}

// listProviders lists all providers (supports ?type=xxx filter)
func (rt *Router) listProviders(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	// support filtering by type through query parameter
	providerType := c.Query("type")

	var err error
	var integrations any

	if providerType != "" {
		integrations, err = identityService.GetProviderByType(providerType)
	} else {
		integrations, err = identityService.GetProviderList()
	}

	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	// build response without timestamps
	type ProviderResponse struct {
		ProviderId   string `json:"providerId"`
		Name         string `json:"name"`
		ProviderType string `json:"providerType"`
		Description  string `json:"description"`
		Priority     int    `json:"priority"`
		IsEnabled    int    `json:"isEnabled"`
	}

	var response []ProviderResponse
	switch v := integrations.(type) {
	case []identitymodel.IdentityIntegration:
		for _, integration := range v {
			response = append(response, ProviderResponse{
				ProviderId:   integration.ProviderId,
				Name:         integration.Name,
				ProviderType: integration.ProviderType,
				Description:  integration.Description,
				Priority:     integration.Priority,
				IsEnabled:    integration.IsEnabled,
			})
		}
	}

	c.Locals(middleware.DETAIL, response)
	return nil
}

// getProvider gets a specific provider
func (rt *Router) getProvider(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	name := c.Params("name")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	provider, err := identityService.GetProvider(name)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, provider)
	return nil
}

// listProviderTypes lists all provider types
func (rt *Router) listProviderTypes(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	providerTypes, err := identityService.GetProviderTypeList()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, providerTypes)
	return nil
}

// ldapLogin handles LDAP login (authentication methods requiring username and password)
func (rt *Router) ldapLogin(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	var req identityservice.LDAPLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, http.BadRequest.Msg, c.Path())
	}

	if req.Username == "" || req.Password == "" {
		return http.WithRepErrMsg(c, http.UsernameArePasswordIsRequired.Code, http.UsernameArePasswordIsRequired.Msg, c.Path())
	}

	userInfo, err := identityService.LDAPLogin(providerName, req.Username, req.Password)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, userInfo)
	return nil
}

// createProvider creates an identity integration provider
func (rt *Router) createProvider(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	var provider identitymodel.IdentityIntegration
	if err := c.BodyParser(&provider); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	// validate required fields
	if provider.Name == "" || provider.ProviderType == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "name and providerType are required fields", c.Path())
	}

	if err := identityService.CreateProvider(&provider); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, provider)
	c.Locals(middleware.OPERATION, "create identity integration provider")
	return nil
}

// updateProvider updates an identity integration provider
func (rt *Router) updateProvider(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	name := c.Params("name")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	var provider identitymodel.IdentityIntegration
	if err := c.BodyParser(&provider); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	if err := identityService.UpdateProvider(name, &provider); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update identity integration provider")
	return nil
}

// toggleProvider toggles the enabled status of an identity integration provider
func (rt *Router) toggleProvider(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	name := c.Params("name")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	if err := identityService.ToggleProvider(name); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "toggle identity integration provider status")
	return nil
}

// deleteProvider deletes an identity integration provider
func (rt *Router) deleteProvider(c *fiber.Ctx) error {
	identityService := rt.Services.IdentityIntegration

	name := c.Params("name")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	if err := identityService.DeleteProvider(name); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "delete identity integration provider")
	return nil
}
