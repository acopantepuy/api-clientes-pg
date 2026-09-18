package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-clientes-pg/models"
	"api-clientes-pg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB configura una base de datos temporal en memoria
func setupTestDB() {
	// file::memory: crea la base de datos en la RAM (se borra al terminar la prueba)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Error conectando a la BD de pruebas")
	}

	db.AutoMigrate(&models.Cliente{})
	models.DB = db // Reemplazamos la conexión global por esta falsa
}

// TestGetClientes es la prueba unitaria (debe empezar con la palabra "Test")
func TestGetClientes(t *testing.T) {
	// 1. Configurar Gin en "modo test" (para evitar logs innecesarios en consola)
	gin.SetMode(gin.TestMode)

	// 2. Preparar nuestra base de datos de mentira y meterle 2 registros
	setupTestDB()
	models.DB.Create(&models.Cliente{
		Nombre:  "Junior Perez",
		Empresa: "Tech C.A",
		Email:   "junior@tech.test",
	})
	models.DB.Create(&models.Cliente{
		Nombre:  "Maria Gomez",
		Empresa: "Dev INC",
		Email:   "maria@dev.test",
	})

	// 3. Levantar un Router de Gin solo para esta prueba
	r := gin.Default()
	r.GET("/clientes", GetClientes)

	// 4. Crear la petición simulada (HTTP GET hacia /clientes)
	req, _ := http.NewRequest(http.MethodGet, "/clientes", nil)

	// 5. Crear el Recorder (esto grabará la respuesta del servidor como si fuera Postman)
	w := httptest.NewRecorder()

	// 6. Ejecutar la petición en nuestro router
	r.ServeHTTP(w, req)

	// --- INICIAN LAS VALIDACIONES (ASSERTS) ---

	// A. Validar que el código HTTP sea 200 (OK)
	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba código %d, pero se obtuvo %d", http.StatusOK, w.Code)
	}

	// B. Decodificar el JSON de respuesta a nuestra estructura
	var response utils.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error al decodificar el JSON: %v", err)
	}

	// C. Validar que la API realmente devuelva los 2 registros que insertamos
	if response.Total != 2 {
		t.Errorf("Se esperaba un Total de 2, pero devolvió %d", response.Total)
	}
}
