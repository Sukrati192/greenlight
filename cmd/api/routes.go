package main

import (
	"github.com/gin-gonic/contrib/expvar"
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.NoMethod(app.methodNotAllowedResponse)
	router.NoRoute(app.notFoundResponse)
	router.Use(app.metrics())
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS())
	router.Use(app.rateLimit())
	router.Use(app.authenticate())
	router.GET("/v1/healthcheck", app.healthcheckHandler)
	router.POST("/v1/users", app.registerUserHandler)
	router.PUT("/v1/users/activated", app.activateUserHandler)
	router.POST("/v1/tokens/authentication", app.createAuthentication)
	router.GET("/debug/vars", expvar.Handler())

	readMovies := router.Group("/v1/movies")
	readMovies.Use(app.requirePermission("movies:read"))
	readMovies.GET("", app.listMoviesHandler)
	readMovies.GET("/:id", app.showMoviesHandler)

	writeMovies := router.Group("/v1/movies")
	writeMovies.Use(app.requirePermission("movies:write"))
	writeMovies.POST("", app.createMoviesHandler)
	writeMovies.PATCH("/:id", app.updateMoviesHandler)
	writeMovies.DELETE("/:id", app.deleteMoviesHandler)
	return router
}
