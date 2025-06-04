package api

import "github.com/gin-gonic/gin"

func Run() {
	r := gin.Default()

	r.Group("/")
	{
		r.POST("/api/register", ApiRegister)
		r.POST("/api/auth", ApiAuth)
	}

	r.GET("/api/items", AuthMiddleware(), ApiGetItems)

	r.POST("/api/orders/create", AuthMiddleware(), ApiCreateOrder)
	r.GET("/api/orders/:id", AuthMiddleware(), ApiGetOrder)
	r.POST("/api/orders/cancel/:id", AuthMiddleware(), ApiCancelOrder)
	r.GET("api/orders/actual", AuthMiddleware(), ApiGetActiveOrders)

	r.GET("/api/user", AuthMiddleware(), ApiGetUser)

	r.GET("/api/photos/:item", ApiGetPhoto)

	r.Run()
}
