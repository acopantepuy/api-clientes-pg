package middlewares

import (
	"net/http"
	"os"

	"api-clientes-pg/utils" // Importamos la nueva estructura

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := c.GetHeader("X-API-Key")
		serverKey := os.Getenv("API_KEY")

		if serverKey == "" || clientKey != serverKey {
			// Usamos utils.APIError en lugar de gin.H
			errorResponse := utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: "Acceso denegado a la operación",
				Detail:  "API Key inválida o no proporcionada en el header X-API-Key",
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse)
			return
		}

		c.Next()
	}
}
