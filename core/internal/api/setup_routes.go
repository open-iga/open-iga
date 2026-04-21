package api

func (r *Router) setupRoutes() {
	r.addHealthRoute()
	r.addLoginRoute()
	r.addLoginCallbackRoute()
}
