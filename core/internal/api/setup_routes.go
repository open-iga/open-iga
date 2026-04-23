package api

import "github.com/danielgtaylor/huma/v2"

func (r *Router) setupRoutes() {
	huma.Get(r.api, "/health", r.healthHandler)
	huma.Get(r.api, "/login/{provider}", r.loginHandler)
	huma.Get(r.api, "/login/{provider}/callback", r.loginCallBackHandler)
}
