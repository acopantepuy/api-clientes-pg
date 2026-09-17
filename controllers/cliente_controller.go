package controllers

import (
    "net/http"

    "api-clientes-pg/models" // Importamos nuestros modelos
    "github.com/gin-gonic/gin"
)

// GetClientes godoc
// @Summary Obtener clientes
// @Tags clientes
// @Produce json
// @Param empresa query string false "Filtrar por empresa"
// @Success 200 {array} models.Cliente
// @Router /clientes [get]
func GetClientes(c *gin.Context) {
    var clientes []models.Cliente
    empresa := c.Query("empresa")

    query := models.DB
    if empresa != "" {
        query = query.Where("empresa = ?", empresa)
    }
    
    query.Find(&clientes)
    c.JSON(http.StatusOK, clientes)
}

// CreateCliente godoc
// @Summary Agregar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Param cliente body models.Cliente true "Datos del cliente"
// @Success 201 {object} models.Cliente
// @Router /clientes [post]
func CreateCliente(c *gin.Context) {
    var cliente models.Cliente
    if err := c.ShouldBindJSON(&cliente); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Router /clientes/masivo [post]
func CreateClientesMasivo(c *gin.Context) {
    var clientes []models.Cliente
    if err := c.ShouldBindJSON(&clientes); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Router /clientes/{id} [put]
func UpdateCliente(c *gin.Context) {
    id := c.Param("id")
    var cliente models.Cliente
    
    if err := models.DB.First(&cliente, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "No encontrado"})
        return
    }
    if err := c.ShouldBindJSON(&cliente); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Router /clientes/{id} [delete]
func DeleteCliente(c *gin.Context) {
    id := c.Param("id")
    if err := models.DB.Delete(&models.Cliente{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"mensaje": "Eliminado"})
}