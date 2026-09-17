package routes

import (
    "api-clientes-pg/controllers"
    "api-clientes-pg/middlewares" // Importamos nuestro middleware
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    clientesGroup := r.Group("/clientes")
    {
        // Ruta PÚBLICA: Cualquiera puede ver los clientes
        clientesGroup.GET("", controllers.GetClientes)

        // Creamos un sub-grupo para las rutas PROTEGIDAS
        // .Use() aplica el middleware a todo lo que esté dentro de este bloque
        protegidas := clientesGroup.Group("")
        protegidas.Use(middlewares.APIKeyMiddleware())
        {
            protegidas.POST("", controllers.CreateCliente)
            protegidas.POST("/masivo", controllers.CreateClientesMasivo)
            protegidas.PUT("/:id", controllers.UpdateCliente)
            protegidas.DELETE("/:id", controllers.DeleteCliente)
        }
    }

    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    return r
}