package http

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "testovoe_again/docs"
)

func (h *Handler) Routing(e *echo.Echo) {
	group := e.Group("/api/v1")

	// роутинг эндпоинтов
	subs := group.Group("/subscriptions")
	{
		subs.POST("", h.Create)
		subs.GET("/:id", h.GetByID)
		subs.GET("/list/:user_id", h.List)
		subs.PUT("/:id", h.Update)
		subs.DELETE("/:id", h.Delete)
	}

	// статистика из второго пункта
	group.POST("/stats", h.GetSum)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	group.GET("/healthcheck", h.Health)
}
