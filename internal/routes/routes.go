package routes

import (
	"net/http"

	"golang-basic/api/internal/config"
	"golang-basic/api/internal/controller"
	appmiddleware "golang-basic/api/internal/middleware"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRoutes() *chi.Mux {
	// Initialize repositories
	userRepo := repository.NewUserRepository(config.DB)
	profileRepo := repository.NewProfileRepository(config.DB)
	authorizationRepo := repository.NewAuthorizationRepository(config.DB)

	// Initialize services
	userService := service.NewUserService(userRepo)
	profileService := service.NewProfileService(profileRepo, userRepo)
	authorizationService := service.NewAuthorizationService(authorizationRepo)
	projectService := service.NewProjectService(nil)

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(config.DB)
	projectService = service.NewProjectService(projectRepo)

	// Initialize controllers
	userController := controller.NewUserController(userService)
	profileController := controller.NewProfileController(profileService)
	permissionController := controller.NewPermissionController(nil) // TODO: Add permission service when implemented
	projectController := controller.NewProjectController(projectService)

	// Initialize authorization middleware
	authMiddleware := appmiddleware.NewAuthorizationMiddleware(authorizationService)

	// Create Chi router
	r := chi.NewRouter()

	// ========== GLOBAL MIDDLEWARE ==========
	r.Use(appmiddleware.RequestLogger) // Custom logger from common.go
	r.Use(middleware.Recoverer)        // Built-in panic recovery

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))

	r.Use(appmiddleware.SetUserIDInContext()) // Extract user ID from header to context

	// ========== HEALTH CHECK ==========
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========== API v1 ROUTES ==========
	r.Route("/api/v1", func(r chi.Router) {

		// ----- USERS -----
		r.Route("/users", func(r chi.Router) {
			// TODO: Re-enable authorization middleware after fixing repository type mismatch
			r.Get("/", userController.GetAllUsers)
			r.Post("/", userController.CreateUser)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", userController.GetUserByID)
				r.Put("/", userController.UpdateUser)
			})
		})

		// ----- PROFILES -----
		r.Route("/profiles", func(r chi.Router) {
			r.Get("/", profileController.GetAllProfiles)
			r.Post("/", profileController.CreateProfile)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", profileController.GetProfileByID)
				r.Put("/", profileController.UpdateProfile)
				r.Delete("/", profileController.DeleteProfile)
			})
		})

		// ----- PROJECTS -----
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectController.ListProjects)
			r.With(authMiddleware.RequirePermission("projects", "write")).Post("/", projectController.CreateProject)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", projectController.GetProjectByID)
				r.With(authMiddleware.RequirePermission("projects", "write")).Put("/", projectController.UpdateProject)
				r.With(authMiddleware.RequirePermission("projects", "delete")).Delete("/", projectController.DeleteProject)
			})
		})

		// ----- PERMISSIONS (Nested Routes) -----
		r.Route("/permissions", func(r chi.Router) {
			// Attributes
			r.Route("/attributes", func(r chi.Router) {
				r.Get("/", permissionController.ListAttributes)
				r.Post("/", permissionController.CreateAttribute)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetAttributeByID)
					r.Delete("/", permissionController.DeleteAttribute)
				})
			})

			// Resources
			r.Route("/resources", func(r chi.Router) {
				r.Get("/", permissionController.ListResources)
				r.Post("/", permissionController.CreateResource)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetResourceByID)
					r.Delete("/", permissionController.DeleteResource)
				})
			})

			// Permissions
			r.Route("/permissions", func(r chi.Router) {
				r.Get("/", permissionController.ListPermissions)
				r.Post("/", permissionController.CreatePermission)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetPermissionByID)
					r.Delete("/", permissionController.DeletePermission)
				})
			})
		})
	})

	return r
}
