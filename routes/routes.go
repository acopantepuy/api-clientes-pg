package routes

import (
    "api-clientes-pg/controllers"
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Agrupamos las rutas de clientes
    clientesGroup := r.Group("/clientes")
    {
        clientesGroup.GET("", controllers.GetClientes)
        clientesGroup.POST("", controllers.CreateCliente)
        clientesGroup.POST("/masivo", controllers.CreateClientesMasivo)
        clientesGroup.PUT("/:id", controllers.UpdateCliente)
        clientesGroup.DELETE("/:id", controllers.DeleteCliente)
    }

    // Ruta de Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    return r
}