package routes

import (
	"net/http"
	"os"

	"golang-basic/api/internal/config"
	"golang-basic/api/internal/controller"
	appmiddleware "golang-basic/api/internal/middleware"
	"golang-basic/api/internal/repository"
	"golang-basic/api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SetupRoutes() *chi.Mux {
	var minioController *controller.MinioController
	if config.MinioClient != nil {
		minioService := service.NewMinioService(config.MinioClient, getEnv("MINIO_BUCKET_NAME", "golang-basic"))
		minioController = controller.NewMinioController(minioService)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(config.DB)
	profileRepo := repository.NewProfileRepository(config.DB)
	authorizationRepo := repository.NewAuthorizationRepository(config.DB)
	permissionRepo := repository.NewPermissionRepository(config.DB)
	policyRepo := repository.NewPolicyRepository(config.DB)
	userPolicyRepo := repository.NewUserPolicyRepository(config.DB)
	projectRepo := repository.NewProjectRepository(config.DB)
	restAPIRepo := repository.NewRestAPIRepository(config.DB)
	graphqlAPIRepo := repository.NewGraphQLAPIRepository(config.DB)
	grpcAPIRepo := repository.NewGrpcAPIRepository(config.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	profileService := service.NewProfileService(profileRepo, userRepo)
	authorizationService := service.NewAuthorizationService(authorizationRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	policyService := service.NewPolicyService(policyRepo)
	userPolicyService := service.NewUserPolicyService(userPolicyRepo, authorizationRepo)
	projectService := service.NewProjectService(projectRepo)
	restAPIService := service.NewRestAPIService(restAPIRepo, projectRepo)
	graphqlAPIService := service.NewGraphQLAPIService(graphqlAPIRepo, projectRepo)
	grpcAPIService := service.NewGrpcAPIService(grpcAPIRepo, projectRepo)

	// Initialize controllers
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	profileController := controller.NewProfileController(profileService)
	permissionController := controller.NewPermissionController(permissionService)
	policyController := controller.NewPolicyController(policyService)
	userPolicyController := controller.NewUserPolicyController(userPolicyService)
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
		AllowedOrigins:   []string{"http://localhost:5173", "https://localhost:5173", "http://localhost:5174", "https://localhost:5174", "http://localhost:3000", "https://localhost:3000"},
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

	// ========== PROFILING ENDPOINTS ==========
	// Mount pprof profiler at /debug for Go profiling
	// Access at: http://localhost:3003/debug/
	r.Mount("/debug", appmiddleware.Profiler())

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
			r.With(jwtMiddleware.Authenticate()).Get("/me", profileController.GetProfileByUserID)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.OptionalAuth()).Get("/", profileController.GetProfileByID)
				r.With(jwtMiddleware.Authenticate()).Put("/", profileController.UpdateProfile)
				r.With(jwtMiddleware.Authenticate()).Delete("/", profileController.DeleteProfile)
			})
		})

		// ----- PROJECTS -----
		r.Route("/projects", func(r chi.Router) {
			r.With(jwtMiddleware.OptionalAuth()).Get("/", projectController.ListProjects)
			// Optimized endpoint with API stats in single query (use this for better performance)
			r.With(jwtMiddleware.OptionalAuth()).Get("/with-stats", projectController.ListProjectsWithStats)
			r.With(jwtMiddleware.Authenticate(), authMiddleware.RequirePermission("projects", "write")).Post("/", projectController.CreateProject)

			r.Route("/{id}", func(r chi.Router) {
				r.With(jwtMiddleware.OptionalAuth()).Get("/", projectController.GetProjectByID)
				r.With(jwtMiddleware.OptionalAuth()).Get("/api-stats", projectController.GetProjectAPIStats)
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

		// ----- MINIO (Object Storage) -----
		if minioController != nil {
			r.Route("/minio", func(r chi.Router) {
				r.With(jwtMiddleware.Authenticate()).Get("/upload/{object}", minioController.GetUploadURL)
				r.With(jwtMiddleware.Authenticate()).Get("/download/{object}", minioController.GetDownloadURL)
			})
		}

		// ----- PERMISSIONS (Nested Routes) -----
		r.Route("/permissions", func(r chi.Router) {
			// Attributes
			r.Route("/attributes", func(r chi.Router) {
				r.Get("/", permissionController.ListAttributes)
				r.Post("/", permissionController.CreateAttribute)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetAttributeByID)
					r.Put("/", permissionController.UpdateAttribute)
					r.Delete("/", permissionController.DeleteAttribute)
				})
			})

			// Resources
			r.Route("/resources", func(r chi.Router) {
				r.Get("/", permissionController.ListResources)
				r.Post("/", permissionController.CreateResource)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetResourceByID)
					r.Put("/", permissionController.UpdateResource)
					r.Delete("/", permissionController.DeleteResource)
				})
			})

			// Permissions
			r.Route("/permissions", func(r chi.Router) {
				r.Get("/", permissionController.ListPermissions)
				r.Post("/", permissionController.CreatePermission)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", permissionController.GetPermissionByID)
					r.Put("/", permissionController.UpdatePermission)
					r.Delete("/", permissionController.DeletePermission)
				})
			})

			// Policies (Full CRUD)
			r.Route("/policies", func(r chi.Router) {
				r.Get("/", policyController.ListPolicies)
				r.With(jwtMiddleware.Authenticate()).Post("/", policyController.CreatePolicy)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", policyController.GetPolicyByID)
					r.With(jwtMiddleware.Authenticate()).Put("/", policyController.UpdatePolicy)
					r.With(jwtMiddleware.Authenticate()).Delete("/", policyController.DeletePolicy)
				})
			})

			// User Policies (Priority-based policy overrides)
			r.Route("/user-policies", func(r chi.Router) {
				r.Get("/", userPolicyController.ListUserPolicies)
				r.Get("/details", userPolicyController.ListUserPolicyDetails)
				r.Post("/", userPolicyController.CreateUserPolicy)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", userPolicyController.GetUserPolicyByID)
					r.Put("/", userPolicyController.UpdateUserPolicy)
					r.Delete("/", userPolicyController.DeleteUserPolicy)
				})

				// Get policies by user ID
				r.Get("/user/{userId}", userPolicyController.GetUserPoliciesByUserID)
			})
		})
	})

	return r
}
