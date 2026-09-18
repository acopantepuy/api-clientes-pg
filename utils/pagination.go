package utils

// PaginatedResponse estandariza las respuestas que devuelven listas de datos
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total" example:"50"`
    Page       int         `json:"page" example:"1"`
    Limit      int         `json:"limit" example:"10"`
    TotalPages int         `json:"total_pages" example:"5"`
}