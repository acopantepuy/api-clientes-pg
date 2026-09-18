package controllers

import (
	"math"
	"net/http"
	"strconv"

	"api-clientes-pg/models" // Importamos nuestros modelos
	"api-clientes-pg/utils"

	"github.com/gin-gonic/gin"
)

// GetClientes godoc
// @Summary Obtener clientes
// @Tags clientes
// @Produce json
// @Param empresa query string false "Filtrar por empresa"
// @Param page query int false "Número de página (default: 1)"
// @Param limit query int false "Cantidad por página (default: 10)"
// @Success 200 {object} utils.PaginatedResponse
// @Router /clientes [get]
func GetClientes(c *gin.Context) {
	var clientes []models.Cliente
	empresa := c.Query("empresa")

	// 1. Obtener parámetros de paginación de la URL (con valores por defecto)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Evitar números negativos o ceros que rompan la base de datos
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// 2. Calcular el Offset matemático
	// Ejemplo: Si estoy en la página 3 y el límite es 10, me salto (3-1)*10 = 20 registros
	offset := (page - 1) * limit

	// 3. Preparar la consulta base
	query := models.DB.Model(&models.Cliente{})
	if empresa != "" {
		query = query.Where("empresa = ?", empresa)
	}

	// 4. Contar el total de registros (ANTES de aplicar límite y offset)
	var total int64
	query.Count(&total)

	// 5. Ejecutar la búsqueda final con Limit y Offset
	query.Limit(limit).Offset(offset).Find(&clientes)

	// 6. Calcular el total de páginas usando math.Ceil (redondeo hacia arriba)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 7. Retornar la respuesta estructurada
	c.JSON(http.StatusOK, utils.PaginatedResponse{
		Data:       clientes,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// CreateCliente godoc
// @Summary Agregar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Param cliente body models.Cliente true "Datos del cliente"
// @Success 201 {object} models.Cliente
// @Failure 400 {object} utils.APIError  <-- Swagger ahora sabe que el error 400 usa esta estructura
// @Router /clientes [post]
// @Security ApiKeyAuth
func CreateCliente(c *gin.Context) {
	var cliente models.Cliente
	if err := c.ShouldBindJSON(&cliente); err != nil {
		// Implementación del error estándar
		c.JSON(http.StatusBadRequest, utils.APIError{
			Status:  http.StatusBadRequest,
			Message: "No se pudo procesar la solicitud",
			Detail:  err.Error(), // Aquí pasamos el error técnico real
		})
		return
	}
	models.DB.Create(&cliente)
	c.JSON(http.StatusCreated, cliente)
}

// CreateClientesMasivo godoc
// @Summary Agregar masivo
// @Tags clientes
// @Accept json
// @Produce json
// @Param clientes body []models.Cliente true "Array de clientes"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.APIError  <-- Swagger ahora sabe que el error 400 usa esta estructura
// @Router /clientes/masivo [post]
// @Security ApiKeyAuth
func CreateClientesMasivo(c *gin.Context) {
	var clientes []models.Cliente
	if err := c.ShouldBindJSON(&clientes); err != nil {
		// Implementación del error estándar
		c.JSON(http.StatusBadRequest, utils.APIError{
			Status:  http.StatusBadRequest,
			Message: "No se pudo procesar la solicitud",
			Detail:  err.Error(), // Aquí pasamos el error técnico real
		})
		return
	}
	models.DB.Create(&clientes)
	c.JSON(http.StatusCreated, gin.H{"mensaje": "Creados", "cantidad": len(clientes)})
}

// UpdateCliente godoc
// @Summary Actualizar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param cliente body models.Cliente true "Datos"
// @Success 200 {object} models.Cliente
// @Failure 400 {object} utils.APIError  <-- Swagger ahora sabe que el error 400 usa esta estructura
// @Router /clientes/{id} [put]
// @Security ApiKeyAuth
func UpdateCliente(c *gin.Context) {
	id := c.Param("id")
	var cliente models.Cliente

	if err := models.DB.First(&cliente, id).Error; err != nil {
		// Implementación del error estándar
		c.JSON(http.StatusBadRequest, utils.APIError{
			Status:  http.StatusBadRequest,
			Message: "No se pudo procesar la solicitud",
			Detail:  err.Error(), // Aquí pasamos el error técnico real
		})
		return
	}
	if err := c.ShouldBindJSON(&cliente); err != nil {
		// Implementación del error estándar
		c.JSON(http.StatusBadRequest, utils.APIError{
			Status:  http.StatusBadRequest,
			Message: "No se pudo procesar la solicitud",
			Detail:  err.Error(), // Aquí pasamos el error técnico real
		})
		return
	}
	models.DB.Save(&cliente)
	c.JSON(http.StatusOK, cliente)
}

// DeleteCliente godoc
// @Summary Eliminar un cliente
// @Tags clientes
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.APIError  <-- Swagger ahora sabe que el error 400 usa esta estructura
// @Router /clientes/{id} [delete]
// @Security ApiKeyAuth
func DeleteCliente(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.Cliente{}, id).Error; err != nil {
		// Implementación del error estándar
		c.JSON(http.StatusBadRequest, utils.APIError{
			Status:  http.StatusBadRequest,
			Message: "Error al eliminar el cliente",
			Detail:  err.Error(), // Aquí pasamos el error técnico real
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensaje": "Eliminado"})
}
