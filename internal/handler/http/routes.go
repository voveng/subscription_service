package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "subscriptions-service/docs"
)

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API
	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", h.Create)
			subscriptions.GET("", h.List)
			subscriptions.GET("/total_cost", h.GetTotalCost)
			subscriptions.GET("/:id", h.GetByID)
			subscriptions.PUT("/:id", h.Update)
			subscriptions.DELETE("/:id", h.Delete)
		}
	}

	return router
}
