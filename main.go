package main

import (
    "fmt"
	"log"
    "net/http"
	"os" //Permite leer variables del sistema operativo

    "github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Importamos godotenv
    "gorm.io/driver/postgres" // <-- Nuevo driver
    "gorm.io/gorm"

    _ "api-clientes-pg/docs" 
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

type Cliente struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Nombre   string `json:"nombre"`
	Email    string `gorm:"unique" json:"email"`
	Telefono string `json:"telefono"`
	Empresa  string `json:"empresa"`
}

var db *gorm.DB

func initDB() {
    // 1. Cargar el archivo .env
    if err := godotenv.Load(); err != nil {
        log.Println("Advertencia: No se encontró el archivo .env, usando variables de entorno del sistema")
    }

    // 2. Leer las variables
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")

    // 3. Construir el DSN dinámicamente
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Caracas", 
        host, user, password, dbname, port)
    
    // 4. Conectar a PostgreSQL
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Error al conectar a PostgreSQL: ", err)
    }
    
    db.AutoMigrate(&Cliente{}) 
    log.Println("Base de datos conectada exitosamente")
}

// GetClientes godoc
// @Summary Obtener clientes
// @Description Obtiene todos los clientes o filtra por empresa
// @Tags clientes
// @Accept json
// @Produce json
// @Param empresa query string false "Filtrar por empresa"
// @Success 200 {array} Cliente
// @Router /clientes [get]
func GetClientes(c *gin.Context) {
    var clientes []Cliente
    empresa := c.Query("empresa")

    query := db
    if empresa != "" {
        query = query.Where("empresa = ?", empresa)
    }
    
    query.Find(&clientes)
    c.JSON(http.StatusOK, clientes)
}

// CreateCliente godoc
// @Summary Agregar un cliente
// @Description Crea un nuevo cliente en Postgres
// @Tags clientes
// @Accept json
// @Produce json
// @Param cliente body Cliente true "Datos del cliente"
// @Success 201 {object} Cliente
// @Router /clientes [post]
func CreateCliente(c *gin.Context) {
    var cliente Cliente
    if err := c.ShouldBindJSON(&cliente); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&cliente)
    c.JSON(http.StatusCreated, cliente)
}

// CreateClientesMasivo godoc
// @Summary Agregar clientes masivamente
// @Description Inserta múltiples clientes en una sola transacción
// @Tags clientes
// @Accept json
// @Produce json
// @Param clientes body []Cliente true "Array de clientes"
// @Success 201 {object} map[string]interface{}
// @Router /clientes/masivo [post]
func CreateClientesMasivo(c *gin.Context) {
    var clientes []Cliente
    if err := c.ShouldBindJSON(&clientes); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&clientes) 
    c.JSON(http.StatusCreated, gin.H{"mensaje": "Clientes creados masivamente", "cantidad": len(clientes)})
}

// UpdateCliente godoc
// @Summary Actualizar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Param id path int true "ID del cliente"
// @Param cliente body Cliente true "Nuevos datos"
// @Success 200 {object} Cliente
// @Router /clientes/{id} [put]
func UpdateCliente(c *gin.Context) {
    id := c.Param("id")
    var cliente Cliente
    
    if err := db.First(&cliente, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Cliente no encontrado"})
        return
    }
    if err := c.ShouldBindJSON(&cliente); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Save(&cliente)
    c.JSON(http.StatusOK, cliente)
}

// DeleteCliente godoc
// @Summary Eliminar un cliente
// @Tags clientes
// @Produce json
// @Param id path int true "ID del cliente"
// @Success 200 {object} map[string]interface{}
// @Router /clientes/{id} [delete]
func DeleteCliente(c *gin.Context) {
    id := c.Param("id")
    if err := db.Delete(&Cliente{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"mensaje": "Cliente eliminado"})
}

// @title API de Clientes (PostgreSQL)
// @version 1.0
// @description API de estudio conectada a PostgreSQL.
// @host localhost:8080
// @BasePath /
func main() {
    initDB()
    r := gin.Default()

    r.GET("/clientes", GetClientes)
    r.POST("/clientes", CreateCliente)
    r.POST("/clientes/masivo", CreateClientesMasivo)
    r.PUT("/clientes/:id", UpdateCliente)
    r.DELETE("/clientes/:id", DeleteCliente)

    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.Run(":8080")
}
