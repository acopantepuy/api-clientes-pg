package routes

import (
	"api-clientes-pg/controllers"
	"api-clientes-pg/middlewares"

	"time" // Necesario para configurar el tiempo de caché del preflight

	"github.com/gin-contrib/cors" // Importamos el middleware CORS
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
 
    // 1. Configuración de CORS

	r.Use(cors.New(cors.Config{
		// Orígenes permitidos (Ej: React en el 3000, Angular en el 4200)
		// Para permitir TODO en desarrollo puedes usar: AllowAllOrigins: true,
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:4200"},

		// Métodos HTTP permitidos
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// Cabeceras permitidas (IMPORTANTE: incluir nuestra X-API-Key)
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "X-API-Key"},

		// Cabeceras que el frontend puede leer de la respuesta
		ExposeHeaders: []string{"Content-Length"},

		// Permite el envío de cookies o tokens de autenticación
		AllowCredentials: true,

		// Tiempo que el navegador cacheará la petición preflight (OPTIONS)
		MaxAge: 12 * time.Hour,
	}))

	// 2. Definición de Rutas (igual que antes)
	clientesGroup := r.Group("/clientes")
	{
		clientesGroup.GET("", controllers.GetClientes)

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
