package utils

// APIError define la estructura estándar para todas las respuestas fallidas de la API
type APIError struct {
    Status  int    `json:"status" example:"400"`
    Message string `json:"message" example:"Datos de entrada inválidos"`
    Detail  string `json:"detail,omitempty" example:"El campo 'email' es obligatorio"`
}