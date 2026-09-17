package main

import (
	"log"

	"api-clientes-pg/models"
	"api-clientes-pg/routes"

	"github.com/joho/godotenv"

	_ "api-clientes-pg/docs" // Necesario para Swagger
)

// @title API de Clientes (PostgreSQL Arquitectura por Capas)
// @version 1.0
// @description API refactorizada usando MVC.
// @host localhost:8080
// @BasePath /
func main() {
	// 1. Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env")
	}

	// 2. Conectar a la Base de Datos
	models.ConnectDatabase()

	// 3. Configurar Rutas
	r := routes.SetupRouter()

	// 4. Iniciar el servidor
	r.Run(":8080")
}
