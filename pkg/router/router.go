// Package router
package router

import (
	"net/http"
	"selfier/internal/module/job"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(jobHandler job.JobHTTPHandler) *gin.Engine {
	r := gin.Default()

	// Tell Gin to load our HTML template
	r.LoadHTMLGlob("templates/*.html")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// // API v1 routes
	// apiV1 := r.Group("/api/v1")
	// {
	// 	jobs := apiV1.Group("/jobs")
	// 	{
	// 		jobs.POST("", jobHandler.CreateJob)
	// 		jobs.GET("", jobHandler.GetAllJobs)
	// 		jobs.GET("/:id", jobHandler.GetJobByID)
	// 		jobs.DELETE("/:id", jobHandler.DeleteJobByID)
	// 	}
	// }

	// --- Swagger and Scalar Setup ---

	// 1. Create a new route for our beautiful Scalar UI
	r.GET("/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "scalar.html", nil)
	})

	// 2. We still need gin-swagger to serve the generated swagger.json file.
	//    Scalar will fetch this file at /swagger/doc.json
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
