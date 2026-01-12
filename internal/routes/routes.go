package routes

import (
	"golang-basic/internal/config"
	"golang-basic/internal/controller"
	"golang-basic/internal/repository"
	"golang-basic/internal/service"
	"net/http"
)

func SetupRoutes() {
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	profileRepo := repository.NewProfileRepository(config.DB)
	profileService := service.NewProfileService(profileRepo, userRepo)
	profileController := controller.NewProfileController(profileService)

	http.HandleFunc("/users", userController.HandleUserRoutes)
	http.HandleFunc("/users/", userController.HandleUserRoutes)

	http.HandleFunc("/profiles", profileController.HandleProfileRoutes)
	http.HandleFunc("/profiles/", profileController.HandleProfileRoutes)
}
