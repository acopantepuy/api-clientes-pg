package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"api-clientes-pg/models"
	"api-clientes-pg/routes"

	"github.com/joho/godotenv"

	docs "api-clientes-pg/docs"
)

// @title API de Clientes (Arquitectura por Capas)
// @version 1.0
// @description API usando MVC.
// @host localhost:8443
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

func main() {
	// 1. Cargamos variables de entorno desde el archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env")
	}

	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("SSL_PORT")

	if appPort == "" {
		appPort = "8080"
	}

	// 2. Configurar Host y Esquema dinámicamente para Swagger
	if appHost != "" {
		docs.SwaggerInfo.Host = appHost + ":" + appPort
	} else {
		docs.SwaggerInfo.Host = "localhost:" + appPort
	}

	if os.Getenv("SSL_ENABLED") == "true" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	appName := os.Getenv("APP_HOST")

	// 3. Conectamos a la base de datos y ejecutamos migraciones automáticas
	models.ConnectDatabase()

	// 4. Configuramos las rutas
	r := routes.SetupRouter()

	if os.Getenv("SSL_ENABLED") == "true" {

		certFile := os.Getenv("SSL_CERT_PATH")
		if certFile == "" { // Verificamos que la variable de entorno esté definida
			log.Fatal("Error: SSL_CERT_PATH no está definido en .env")
		}

		keyFile := os.Getenv("SSL_KEY_PATH")
		if keyFile == "" { // verificamos que la variable de entorno esté definida
			log.Fatal("Error: SSL_KEY_PATH no está definido en .env")
		}

		// Hardening TLS: Mínimo TLS 1.2 o TLS 1.3
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
		}

		srv := &http.Server{
			Addr:         ":" + appPort,
			Handler:      r,
			TLSConfig:    tlsConfig,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		log.Println("Servidor seguro escuchando en https://" + appName + ":" + appPort)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error crítico al iniciar HTTPS: %v", err)
		}
	} else {
		r.Run(":" + appPort)
		log.Println("Servidor escuchando en http://" + appName + ":" + appPort)
	}

}
