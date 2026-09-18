package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"api-clientes-pg/middlewares" // Importamos el middleware
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
	// Limpia los registros previos entre ejecuciones de tests
	db.Exec("DELETE FROM clientes")

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

func TestCreateCliente_Exito(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB()

	// 1. Configuramos la variable de entorno solo para esta prueba
	os.Setenv("API_KEY", "clave_secreta_test")
	defer os.Unsetenv("API_KEY") // Asegura que se borre al terminar la prueba

	// 2. Levantamos el router INCLUYENDO el middleware
	r := gin.Default()
	r.Use(middlewares.APIKeyMiddleware())
	r.POST("/clientes", CreateCliente)

	// 3. Preparamos los datos del nuevo cliente en JSON
	payload := []byte(`{
        "nombre": "Carlos Dev",
        "email": "carlos@test.com",
        "empresa": "Tech Solutions"
    }`)

	// 4. Creamos la petición enviando el payload (usando bytes.NewBuffer)
	req, _ := http.NewRequest(http.MethodPost, "/clientes", bytes.NewBuffer(payload))

	// 5. ¡LA MAGIA! Inyectamos los Headers
	req.Header.Set("Content-Type", "application/json") // Avisamos que enviamos un JSON
	req.Header.Set("X-API-Key", "clave_secreta_test")  // Pasamos el middleware

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// --- VALIDACIONES ---

	// A. Validar que la respuesta sea 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("Se esperaba 201, se obtuvo %d", w.Code)
	}

	// B. Validar que realmente se guardó en la base de datos de pruebas
	var count int64
	models.DB.Model(&models.Cliente{}).Count(&count)
	if count != 1 {
		t.Errorf("Se esperaba 1 cliente en la BD, pero hay %d", count)
	}
}

func TestCreateCliente_FalloAutenticacion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB()

	os.Setenv("API_KEY", "clave_secreta_test")
	defer os.Unsetenv("API_KEY")

	r := gin.Default()
	r.Use(middlewares.APIKeyMiddleware())
	r.POST("/clientes", CreateCliente)

	payload := []byte(`{"nombre": "Hacker", "email": "hacker@test.com"}`)
	req, _ := http.NewRequest(http.MethodPost, "/clientes", bytes.NewBuffer(payload))

	// Enviamos el Content-Type, pero NO enviamos el X-API-Key
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// --- VALIDACIONES ---

	// A. Validar que el servidor nos rechace con 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Se esperaba que el servidor rechazara con 401, pero devolvió %d", w.Code)
	}

	// B. Validar que NO se guardó nada en la base de datos
	var count int64
	models.DB.Model(&models.Cliente{}).Count(&count)
	if count != 0 {
		t.Errorf("El cliente se guardó en BD saltándose la seguridad. Registros: %d", count)
	}
}
