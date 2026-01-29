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
	projectRepo := repository.NewProjectRepository(config.DB)
	restAPIRepo := repository.NewRestAPIRepository(config.DB)
	graphqlAPIRepo := repository.NewGraphQLAPIRepository(config.DB)
	grpcAPIRepo := repository.NewGrpcAPIRepository(config.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	profileService := service.NewProfileService(profileRepo, userRepo)
	authorizationService := service.NewAuthorizationService(authorizationRepo)
	projectService := service.NewProjectService(projectRepo)
	restAPIService := service.NewRestAPIService(restAPIRepo, projectRepo)
	graphqlAPIService := service.NewGraphQLAPIService(graphqlAPIRepo, projectRepo)
	grpcAPIService := service.NewGrpcAPIService(grpcAPIRepo, projectRepo)

	// Initialize controllers
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	profileController := controller.NewProfileController(profileService)
	permissionController := controller.NewPermissionController(nil)
	projectController := controller.NewProjectController(projectService)
	restAPIController := controller.NewRestAPIController(restAPIService)
	graphqlAPIController := controller.NewGraphQLAPIController(graphqlAPIService)
	grpcAPIController := controller.NewGrpcAPIController(grpcAPIService)

	// Initialize middleware
	jwtMiddleware := appmiddleware.NewJWTMiddleware()
	authMiddleware := appmiddleware.NewAuthorizationMiddleware(authorizationService)

	// Create Chi router
	r := chi.NewRouter()

	// ========== GLOBAL MIDDLEWARE ==========
	// Per Specs: Fast middleware first, I/O bound (logger) after
	r.Use(middleware.RequestID) // 1st - Fast
	r.Use(middleware.RealIP)    // 2nd - Fast
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Compress(5)) // Before auth
	r.Use(middleware.Recoverer)   // Last
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// Logger (I/O bound) placed after fast middleware per specs
	r.Use(appmiddleware.RequestLogger)

	// ========== HEALTH CHECK ==========
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ========== AUTH ROUTES (PUBLIC) ==========
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authController.Login)
		r.Post("/register", userController.CreateUser)
		r.Post("/refresh", authController.RefreshToken)
		r.With(jwtMiddleware.Authenticate()).Post("/logout", authController.Logout)
	})

	// ========== API v1 ROUTES ==========
	r.Route("/api/v1", func(r chi.Router) {

		// ----- USERS -----
		r.Route("/users", func(r chi.Router) {
			r.With(jwtMiddleware.Authenticate()).Get("/", userController.GetAllUsers)
			r.Post("/", userController.CreateUser)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate()).Get("/", userController.GetUserByID)
				r.With(jwtMiddleware.Authenticate()).Put("/", userController.UpdateUser)
				r.With(jwtMiddleware.Authenticate()).Put("/password", userController.UpdatePassword)
			})
		})

		// ----- PROFILES -----
		r.Route("/profiles", func(r chi.Router) {
			r.With(jwtMiddleware.OptionalAuth()).Get("/", profileController.GetAllProfiles)
			r.With(jwtMiddleware.Authenticate()).Post("/", profileController.CreateProfile)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.OptionalAuth()).Get("/", profileController.GetProfileByID)
				r.With(jwtMiddleware.Authenticate()).Put("/", profileController.UpdateProfile)
				r.With(jwtMiddleware.Authenticate()).Delete("/", profileController.DeleteProfile)
			})
		})

		// ----- PROJECTS -----
		r.Route("/projects", func(r chi.Router) {
			r.With(jwtMiddleware.OptionalAuth()).Get("/", projectController.ListProjects)
			r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("projects", "write")).Post("/", projectController.CreateProject)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.OptionalAuth()).Get("/", projectController.GetProjectByID)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("projects", "write")).Put("/", projectController.UpdateProject)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("projects", "delete")).Delete("/", projectController.DeleteProject)
			})

			// REST APIs nested under projects
			r.Route("/{projectID}/rest-apis", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "create")).Post("/", restAPIController.CreateRestAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "read")).Get("/", restAPIController.ListRestAPIsByProject)
			})

			// GraphQL APIs nested under projects
			r.Route("/{projectID}/graphql-apis", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "create")).Post("/", graphqlAPIController.CreateGraphQLAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "read")).Get("/", graphqlAPIController.ListGraphQLAPIsByProject)
			})

			// gRPC APIs nested under projects
			r.Route("/{projectID}/grpc-apis", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "create")).Post("/", grpcAPIController.CreateGrpcAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "read")).Get("/", grpcAPIController.ListGrpcAPIsByProject)
			})
		})

		// ----- REST APIs -----
		r.Route("/rest-apis", func(r chi.Router) {
			r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "read")).Get("/", restAPIController.ListRestAPIs)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "read")).Get("/", restAPIController.GetRestAPIByID)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "update")).Put("/", restAPIController.UpdateRestAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_rest_api", "delete")).Delete("/", restAPIController.DeleteRestAPI)
			})
		})

		// ----- GraphQL APIs -----
		r.Route("/graphql-apis", func(r chi.Router) {
			r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "read")).Get("/", graphqlAPIController.ListGraphQLAPIs)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "read")).Get("/", graphqlAPIController.GetGraphQLAPIByID)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "update")).Put("/", graphqlAPIController.UpdateGraphQLAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_graphql_api", "delete")).Delete("/", graphqlAPIController.DeleteGraphQLAPI)
			})
		})

		// ----- gRPC APIs -----
		r.Route("/grpc-apis", func(r chi.Router) {
			r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "read")).Get("/", grpcAPIController.ListGrpcAPIs)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "read")).Get("/", grpcAPIController.GetGrpcAPIByID)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "update")).Put("/", grpcAPIController.UpdateGrpcAPI)
				r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("api_docs_grpc_api", "delete")).Delete("/", grpcAPIController.DeleteGrpcAPI)
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
