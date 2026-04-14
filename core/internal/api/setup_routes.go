package api

func (r *Router) setupRoutes() {
	r.createHealthRouter()
	r.createLoginRouter()
	r.createLoginCallback()
}
