# api-clientes-pg
Implementación de APIs en GoLang e implementación de documentación con Swagger

# Guía de Estudio

Para adaptar este ejercicio a **PostgreSQL**, una de las grandes ventajas de usar un ORM como GORM es que la lógica de nuestros endpoints (handlers) se mantiene exactamente igual. Solo necesitamos cambiar el driver de la base de datos y la cadena de conexión (DSN).

**1.Preparación del Entorno:**Instalar el driver de Postgres.

En la terminal, inicializamos el proyecto y descargamos las librerías, esta vez incluyendo el driver de PostgreSQL en lugar del de SQLite.


```bash
mkdir api-clientes-pg
cd api-clientes-pg
go mod init api-clientes-pg

# Framework web y ORM core
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm

# Driver de PostgreSQL para GORM
go get -u gorm.io/driver/postgres

# Librerías para Swagger
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest
```

**2.Preparar la Base de Datos:**Levantar Postgres rápidamente.

Para que el Junior pueda probar sin instalar PostgreSQL localmente, la forma más rápida es usar Docker. En otra terminal, levanta un contenedor con este comando:

```bash
docker run --name postgres-dev -e POSTGRES_USER=usuario -e POSTGRES_PASSWORD=secreto -e POSTGRES_DB=clientes_db -p 5432:5432 -d postgres
```

_Si ya tiene PostgreSQL instalado, solo debe asegurarse de crear una base de datos llamada `clientes_db`._

**3.Configurar main.go para PostgreSQL:**Imports y Conexión.

Crea el archivo `main.go`. Aquí reemplazamos SQLite por Postgres. Observa cómo cambia la función `initDB` utilizando un DSN (Data Source Name).

```go
package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
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
    // Cadena de conexión a PostgreSQL
    dsn := "host=localhost user=usuario password=secreto dbname=clientes_db port=5432 sslmode=disable TimeZone=America/Caracas"
    
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Error al conectar a PostgreSQL: ", err)
    }
    
    // Crea la tabla 'clientes' automáticamente si no existe
    db.AutoMigrate(&Cliente{}) 
    log.Println("Base de datos PostgreSQL conectada y migrada")
}
```

**4.Agregar los Endpoints (Handlers):**La lógica se mantiene intacta.

Añade el resto de las funciones debajo de `initDB`. Gracias a GORM, el código CRUD es agnóstico a la base de datos.

```go
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
```

**5.Generar Swagger y Ejecutar:**Compilar y probar.

Exactamente igual que en el ejemplo anterior, ejecutamos los comandos para autogenerar la documentación y arrancar el servidor:

```bash
swag init
go mod tidy
go run main.go
```

Al visitar `http://localhost:8080/swagger/index.html`, el Junior podrá probar la API interactuando directamente con PostgreSQL.

# Implementar el env

Excelente paso. Escribir credenciales (como contraseñas o nombres de usuario) directamente en el código fuente es un riesgo grave de seguridad, ya que si subes el código a un repositorio, cualquier persona podría verlas.

Para solucionarlo, usaremos un archivo oculto llamado `.env` y la librería `godotenv` para leerlo. Aquí tienes cómo implementarlo paso a paso:

**1.Instalar godotenv:**Descargar la dependencia.

En tu terminal, dentro de la carpeta del proyecto (`api-clientes-pg`), instala la librería ejecutando:

```bash
go get github.com/joho/godotenv
```

**2.Crear el archivo .env:**Definir las variables.

Crea un archivo nuevo en la raíz de tu proyecto y nómbralo exactamente `.env` (sin nombre antes del punto).

Dentro de este archivo, coloca las credenciales de tu base de datos utilizando el formato `CLAVE=VALOR`:

```bash
DB_HOST=localhost
DB_USER=usuario
DB_PASSWORD=secreto
DB_NAME=clientes_db
DB_PORT=5432
```

**3.Modificar main.go:**Importar y leer las variables.

Abre tu archivo `main.go`. Vamos a agregar los nuevos paquetes (`os` y `godotenv`) en la sección de imports, y actualizaremos la función `initDB` para que construya el DSN dinámicamente.

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "os" // Permite leer variables del sistema operativo

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv" // Importamos godotenv
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    _ "api-clientes-pg/docs" 
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// ... (La estructura Cliente se mantiene igual)

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

// ... (El resto de handlers y la función main se mantienen exactamente iguales)
```

**4.Crear el archivo .gitignore:**Práctica obligatoria de seguridad.

El propósito del archivo `.env` es mantener las contraseñas fuera del control de versiones. Para asegurarnos de que Git nunca suba este archivo, crea un archivo llamado `.gitignore` en la raíz del proyecto y añade esta línea:

```plaintext
# Ignorar archivo de variables de entorno
.env
```

> **Consejo para el Junior:** Lo normal es subir un archivo llamado `.env.example` al repositorio, el cual contiene las claves pero con valores vacíos o falsos (ej. `DB_PASSWORD=tunuevapassword`), para que otros desarrolladores sepan qué variables necesitan configurar.

Al ejecutar de nuevo `go run main.go`, tu aplicación leerá las credenciales desde el archivo `.env` de forma segura.

# Separar la lógica en capas

Separar el código en múltiples archivos y carpetas es lo que llamamos aplicar una **Arquitectura de Software** (en este caso, un patrón MVC o por capas). Esto ayuda a que el proyecto sea escalable, fácil de leer y de mantener. 

Para este ejercicio, organizaremos el proyecto en tres capas principales:  

1. **Modelos (`models`):** Definen las estructuras de datos y la conexión a la base de datos.    
2. **Controladores (`controllers`):** Contienen la lógica de los endpoints (lo que pasa cuando se recibe una petición).
3. **Rutas (`routes`):** Asignan cada URL a su controlador correspondiente.

Aquí tienes la guía de refactorización paso a paso:

**1.Estructura de Carpetas:**Crear los directorios.

En la terminal, dentro de tu proyecto (`api-clientes-pg`), crea las siguientes carpetas:

```bash
mkdir models controllers routes
```

Tu proyecto ahora debería verse así:

```plaintext
api-clientes-pg/
├── .env
├── main.go
├── models/
├── controllers/
└── routes/
```

**2.Capa de Modelos (Base de Datos):**Archivos cliente.go y setup.go.

Vamos a mover la estructura del cliente y la conexión a la base de datos.

**1. Crea el archivo `models/cliente.go`:**

```go
package models

type Cliente struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Nombre   string `json:"nombre"`
    Email    string `gorm:"unique" json:"email"`
    Telefono string `json:"telefono"`
    Empresa  string `json:"empresa"`
}
```

**2. Crea el archivo `models/setup.go`:**

Aquí moveremos la función que conecta a PostgreSQL. Nota que cambiamos `db` (minúscula) a `DB` (mayúscula) para que pueda ser exportada y usada por los controladores.

```go
package models

import (
    "fmt"
    "log"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// DB es la variable global que usaremos en los controladores
var DB *gorm.DB

func ConnectDatabase() {
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Caracas", 
        host, user, password, dbname, port)
    
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Error al conectar a PostgreSQL: ", err)
    }
    
    DB.AutoMigrate(&Cliente{}) 
    log.Println("Base de datos conectada exitosamente")
}
```

**3.Capa de Controladores (Lógica):**Archivo cliente_controller.go.

Crea el archivo **`controllers/cliente_controller.go`**. Moveremos todas las funciones handler aquí.


> _Nota para el Junior:_ Observa que ahora importamos nuestro propio paquete `models` y usamos `models.DB` en lugar de `db`. También, en los comentarios de Swagger, usamos `models.Cliente`.


```go
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
```

**4.Capa de Rutas:**Archivo routes.go.

Crea el archivo **`routes/routes.go`**. Aquí agruparemos todas las URLs. Esto limpia drásticamente nuestro archivo principal.

```go
package routes

import (
    "api-clientes-pg/controllers"
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Agrupamos las rutas de clientes
    clientesGroup := r.Group("/clientes")
    {
        clientesGroup.GET("", controllers.GetClientes)
        clientesGroup.POST("", controllers.CreateCliente)
        clientesGroup.POST("/masivo", controllers.CreateClientesMasivo)
        clientesGroup.PUT("/:id", controllers.UpdateCliente)
        clientesGroup.DELETE("/:id", controllers.DeleteCliente)
    }

    // Ruta de Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    return r
}
```

**5.El nuevo main.go:**Limpieza total.

Ahora reemplaza todo el contenido de tu **`main.go`** en la raíz del proyecto por esto. ¡Mira lo limpio y profesional que queda!

```go
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
```

**6.Actualizar Swagger y Ejecutar:**Comando obligatorio.

Como movimos los comentarios de Swagger (las anotaciones `@`) hacia la carpeta `controllers`, **debemos decirle a Swagger que vuelva a leer todo el proyecto** antes de ejecutar.

En la terminal, ejecuta:

```bash
swag init
go run main.go
```

Al aplicar esta estructura, el desarrollador Junior aprenderá a separar responsabilidades (Separation of Concerns). Si en el futuro necesita agregar una tabla de `Productos`, simplemente creará `models/producto.go`, `controllers/producto_controller.go` y añadirá las rutas correspondientes, sin tocar el código de los clientes.

# Implementando Middleware de seguridad SENCILLO

Un **Middleware** es una función que intercepta la petición HTTP _antes_ de que llegue a tu controlador. Es ideal para tareas como autenticación (verificar un token o API Key), registrar logs, o comprimir respuestas. 

Vamos a implementar un middleware muy sencillo que exija un header llamado `X-API-Key` para poder crear, actualizar o eliminar clientes, mientras que las consultas (GET) seguirán siendo públicas.

Aquí tienes los pasos para implementarlo en tu arquitectura:


**1.Configurar el API Key:**Variable en el .env.

Para no poner la contraseña directamente en el código, agrega una nueva línea a tu archivo `.env` en la raíz del proyecto.

```bash
DB_HOST=localhost
DB_USER=usuario
DB_PASSWORD=secreto
DB_NAME=clientes_db
DB_PORT=5432
# Nueva línea:
API_KEY=secreto123
```

**2.Crear el Middleware:**Nueva carpeta.

En la terminal, crea una nueva carpeta llamada `middlewares` en la raíz de tu proyecto:

```bash
mkdir middlewares
```

Dentro de esa carpeta, crea el archivo **`middlewares/auth.go`**:

```go
package middlewares

import (
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

// APIKeyMiddleware verifica que la petición contenga el header X-API-Key correcto
func APIKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Obtenemos el valor del header que envió el cliente (Postman/Swagger)
        clientKey := c.GetHeader("X-API-Key")
        
        // 2. Obtenemos el valor real desde nuestro archivo .env
        serverKey := os.Getenv("API_KEY")

        // 3. Comparamos. Si el servidor no tiene key configurada o la enviada es incorrecta:
        if serverKey == "" || clientKey != serverKey {
            // Abortamos la petición y respondemos 401 Unauthorized
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Acceso denegado. API Key inválida o ausente.",
            })
            return
        }

        // 4. Si todo es correcto, dejamos que la petición continúe hacia el controlador
        c.Next()
    }
}
```

**3.Aplicar el Middleware a las Rutas Protegidas:**Modificar routes.go.

Abre tu archivo **`routes/routes.go`**. Vamos a importar nuestra nueva carpeta `middlewares` y aplicar la protección solo a los métodos de escritura (POST, PUT, DELETE).

```go
package routes

import (
    "api-clientes-pg/controllers"
    "api-clientes-pg/middlewares" // Importamos nuestro middleware
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    clientesGroup := r.Group("/clientes")
    {
        // Ruta PÚBLICA: Cualquiera puede ver los clientes
        clientesGroup.GET("", controllers.GetClientes)

        // Creamos un sub-grupo para las rutas PROTEGIDAS
        // .Use() aplica el middleware a todo lo que esté dentro de este bloque
        protegidas := clientesGroup.Group("")
        protegidas.Use(middlewares.APIKeyMiddleware())
        {
            protegidas.POST("", controllers.CreateCliente)
            protegidas.POST("/masivo", controllers.CreateClientesMasivo)
            protegidas.PUT("/:id", controllers.UpdateCliente)
            protegidas.DELETE("/:id", controllers.DeleteCliente)
        }
    }

    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    return r
}
```

**4.Avisar a Swagger sobre el API Key:**Actualizar comentarios.

Para que el Junior pueda seguir probando la API desde Swagger, necesitamos que la interfaz gráfica tenga un botón de "Authorize" donde él pueda escribir el API Key.

**1. Modifica `main.go`** y agrega estas dos líneas en la cabecera (antes de `func main()`):

```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
```

**2. Modifica `controllers/cliente_controller.go`** y agrega el tag `@Security` solo a las funciones POST, PUT y DELETE. Por ejemplo, en `CreateCliente` quedaría así:

```go
// CreateCliente godoc
// @Summary Agregar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Security ApiKeyAuth  <-- NUEVA LÍNEA AÑADIDA
// @Param cliente body models.Cliente true "Datos del cliente"
// @Success 201 {object} models.Cliente
// @Router /clientes [post]
func CreateCliente(c *gin.Context) {
    // ... (el código queda igual)
}
```

_Haz lo mismo agregando `@Security ApiKeyAuth` en `CreateClientesMasivo`, `UpdateCliente` y `DeleteCliente`._

**5.Ejecutar los cambios:**Compilar y probar.

Como hicimos cambios en los comentarios, vuelve a generar la documentación y arranca el servidor:

```bash
swag init
go run main.go
```

Al abrir `http://localhost:8080/swagger/index.html`, verás un candado de "Authorize" en la parte superior. Si intentas ejecutar un POST o un DELETE sin poner allí el valor `secreto123` (configurado en el header `X-API-Key`), el servidor rechazará la petición con un error 401.

# Manejo de respuestas de Error

Estandarizar las respuestas de error es una de las mejores prácticas en el desarrollo de APIs. Si el frontend (o el cliente que consume tu API) recibe siempre la misma estructura cuando ocurre un fallo, puede programar interceptores automáticos para mostrar alertas sin tener que adivinar qué formato de error devolvió cada endpoint.

Vamos a implementar un formato de error que devuelva siempre un **Código HTTP**, un **Mensaje amigable** (para el usuario final) y un **Detalle técnico** (útil para el desarrollador).

**1.Crear la estructura del Error (DTO):**Carpeta utils.

Crea una nueva carpeta llamada `utils` en la raíz de tu proyecto y dentro crea el archivo **`utils/errors.go`**.

Esta estructura (Data Transfer Object o DTO) será el molde estricto que usarán todos nuestros errores.

```go
package utils

// APIError define la estructura estándar para todas las respuestas fallidas de la API
type APIError struct {
    Status  int    `json:"status" example:"400"`
    Message string `json:"message" example:"Datos de entrada inválidos"`
    Detail  string `json:"detail,omitempty" example:"El campo 'email' es obligatorio"`
}
```

> **Nota:** La etiqueta `omitempty` en `detail` le indica a Go que, si el campo está vacío, no lo incluya en el JSON final, manteniendo la respuesta más limpia.
> 
>   

**2.Refactorizar el Middleware:**middlewares/auth.go.

Abre tu middleware y cambia la respuesta antigua `gin.H{...}` por nuestra nueva estructura estandarizada.


```go
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
```

**3.Refactorizar los Controladores:**controllers/cliente_controller.go.

Ahora vamos a estandarizar los errores dentro de los endpoints. Por ejemplo, en el endpoint `CreateCliente`, interceptamos cuando el JSON viene mal formado.

Además, agregaremos las etiquetas de Swagger (`@Failure`) para que la documentación muestre esta nueva estructura. Modifica la función así:


```go
package controllers

import (
    "net/http"

    "api-clientes-pg/models"
    "api-clientes-pg/utils" // Importamos utils
    "github.com/gin-gonic/gin"
)

// CreateCliente godoc
// @Summary Agregar un cliente
// @Tags clientes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param cliente body models.Cliente true "Datos del cliente"
// @Success 201 {object} models.Cliente
// @Failure 400 {object} utils.APIError  <-- Swagger ahora sabe que el error 400 usa esta estructura
// @Router /clientes [post]
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
```

_Puedes replicar esta misma lógica (sustituir `gin.H` por `utils.APIError` y agregar el `@Failure`) en `UpdateCliente` (para el error 404 de no encontrado y 400 de bind) y en `DeleteCliente` (para el error 500)._

**4.Actualizar Swagger:**Generar y visualizar.

Como agregamos la anotación `@Failure 400 {object} utils.APIError` y creamos un paquete nuevo, debemos volver a inicializar la documentación.


```bash
swag init
go run main.go
```

Al visitar tu interfaz de Swagger y revisar el endpoint POST `/clientes`, el Junior podrá ver que, debajo de las respuestas exitosas, ahora aparece documentado el error 400 con su esquema detallado (`status`, `message`, `detail`), sirviendo como contrato estricto para quien consuma la API.

# Implementar paginación

La paginación es fundamental cuando una base de datos crece, ya que devolver miles de registros en una sola petición sobrecarga el servidor y hace que la aplicación cliente (el frontend) sea muy lenta.


En GORM, la paginación se logra combinando dos métodos:

- `Limit(n)`: Indica cuántos registros traer como máximo.
- `Offset(n)`: Indica cuántos registros saltarse antes de empezar a contar.

Para hacerlo como un profesional, no solo devolveremos los datos, sino también la "metadata" (total de páginas, página actual, etc.) para que el frontend pueda pintar los botones de "Siguiente" o "Anterior".

**1.Crear el DTO de Paginación:**Carpeta utils.

Primero, vamos a definir una estructura estándar para responder. En tu carpeta `utils`, crea un archivo llamado **`utils/pagination.go`**:

```go
package utils

// PaginatedResponse estandariza las respuestas que devuelven listas de datos
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total" example:"50"`
    Page       int         `json:"page" example:"1"`
    Limit      int         `json:"limit" example:"10"`
    TotalPages int         `json:"total_pages" example:"5"`
}
```

> **Nota para el Junior:** Usamos `interface{}` en `Data` para que esta estructura sea reutilizable. Mañana podrás usarla tanto para paginar `Clientes` como `Productos` o `Usuarios`.
> 
>   

**2.Implementar la Lógica en el GET:**controllers/cliente_controller.go.

Abre tu controlador de clientes y modifica la función `GetClientes`.

Añadiremos la importación de `math` (para redondear el total de páginas) y `strconv` (para convertir los parámetros de la URL, que vienen como texto, a números enteros).

```go
package controllers

import (
    "math"
    "net/http"
    "strconv" // Para convertir string a int

    "api-clientes-pg/models"
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
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }

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
```

> **El error clásico del Junior:** Intentar hacer el `.Count(&total)` _después_ de aplicar `.Limit()`. Si lo haces después, el conteo siempre te dará como máximo el número del límite, perdiendo el total real de tu tabla.
> 
>   

**3.Generar Documentación Swagger:**Actualizar y probar.

Como hemos añadido nuevos parámetros `@Param` para `page` y `limit`, y hemos cambiado la respuesta esperada a `@Success 200 {object} utils.PaginatedResponse`, debemos actualizar Swagger.

En tu terminal, ejecuta:

```bash
swag init
go run main.go
```

Ahora, si el Junior entra a `http://localhost:8080/swagger/index.html` y prueba el endpoint GET, verá campos para introducir `page` y `limit`.

Si añade 25 clientes a la base de datos y pide la página 1 con un límite de 10, recibirá una respuesta muy limpia indicándole que está en la página 1 de 3, y le mostrará solo los primeros 10 registros.


# Implementar CORS en el frontend

CORS (Cross-Origin Resource Sharing) es un mecanismo de seguridad integrado en los navegadores web. Por defecto, un navegador bloquea las peticiones que un frontend (ej. `http://localhost:3000` en React) hace a un backend que está en otro puerto o dominio (ej. `http://localhost:8080`), a menos que el backend autorice explícitamente ese origen.

Para solucionar esto en Gin, utilizaremos el paquete oficial de la comunidad. Aquí tienes la implementación paso a paso:

**1.Instalar el Middleware CORS:**Descargar el paquete oficial.

En tu terminal, dentro de la raíz del proyecto, ejecuta el siguiente comando para descargar el middleware de CORS mantenido por el equipo de Gin:

```bash
go get github.com/gin-contrib/cors
```

**2.Configurar e Inyectar CORS:**Modificar routes/routes.go.

Abre tu archivo **`routes/routes.go`**. Vamos a importar el paquete nuevo y a configurarlo justo después de inicializar el router (`r := gin.Default()`), pero **antes** de definir nuestras rutas.

```go
package routes

import (
    "time" // Necesario para configurar el tiempo de caché del preflight

    "api-clientes-pg/controllers"
    "api-clientes-pg/middlewares"
    
    "github.com/gin-contrib/cors" // Importamos el middleware CORS
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // 1. Configuración de CORS
    r.Use(cors.New(cors.Config{
        // Orígenes permitidos (Ej: React en el 3000, Angular en el 4200)
        // Para permitir TODO en desarrollo puedes usar: AllowAllOrigins: true,
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:4200"},
        
        // Métodos HTTP permitidos
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        
        // Cabeceras permitidas (IMPORTANTE: incluir nuestra X-API-Key)
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-API-Key"},
        
        // Cabeceras que el frontend puede leer de la respuesta
        ExposeHeaders:    []string{"Content-Length"},
        
        // Permite el envío de cookies o tokens de autenticación
        AllowCredentials: true,
        
        // Tiempo que el navegador cacheará la petición preflight (OPTIONS)
        MaxAge: 12 * time.Hour,
    }))

    // 2. Definición de Rutas (igual que antes)
    clientesGroup := r.Group("/clientes")
    {
        clientesGroup.GET("", controllers.GetClientes)

        protegidas := clientesGroup.Group("")
        protegidas.Use(middlewares.APIKeyMiddleware())
        {
            protegidas.POST("", controllers.CreateCliente)
            protegidas.POST("/masivo", controllers.CreateClientesMasivo)
            protegidas.PUT("/:id", controllers.UpdateCliente)
            protegidas.DELETE("/:id", controllers.DeleteCliente)
        }
    }

    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    return r
}
```

**3.Entender los conceptos clave:**Teoría para el Junior.

Es vital que el Junior entienda tres detalles de esta configuración:

1. **El método `OPTIONS` (Preflight):** Antes de enviar un POST o un DELETE, el navegador envía una petición oculta llamada `OPTIONS` para preguntar "¿Tengo permiso para hacer esto?". Si no permites el método `OPTIONS` en `AllowMethods`, la petición real fallará.    
2. **La cabecera `X-API-Key`:** Como creamos un middleware de seguridad personalizado en el paso anterior, debemos decirle explícitamente a CORS que acepte la cabecera `X-API-Key`. Si no lo hacemos, el navegador bloqueará la petición del frontend.
3. **El orden importa:** El middleware `r.Use(cors...)` debe colocarse **siempre arriba** de las rutas. Si lo colocas al final del archivo, las rutas se procesarán sin la protección CORS.


Con esto, tu API está completamente lista para ser consumida de forma segura por cualquier Single Page Application (SPA) en desarrollo.

# Pruebas unitarias

Las pruebas unitarias (Unit Tests) son esenciales para asegurar que tu código funciona antes de enviarlo a producción. En Go, la librería estándar incluye el paquete `net/http/httptest`, que actúa como un "navegador falso" capaz de hacer peticiones a tu API sin necesidad de levantar el servidor web en un puerto real.

Para este ejercicio, aplicaremos un truco profesional muy común: usaremos **SQLite en memoria** exclusivamente para la prueba. Así evitamos conectarnos a PostgreSQL, garantizando que la prueba sea rápida y no ensucie nuestra base de datos real.

Aquí tienes el paso a paso:

**1.Instalar SQLite:**Dependencia exclusiva para testing.

En tu terminal, descarga el driver de SQLite de GORM. Solo lo usaremos en nuestro archivo de pruebas.

```bash
go get -u gorm.io/driver/sqlite
```

**2.Crear el archivo de pruebas:**controllers/cliente_controller_test.go.

En Go, los archivos de prueba deben terminar siempre en `_test.go` y ubicarse en la misma carpeta del código que van a probar.

Crea el archivo **`controllers/cliente_controller_test.go`** y añade este código:

```go
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
    models.DB.Create(&models.Cliente{Nombre: "Junior Perez", Empresa: "Tech C.A"})
    models.DB.Create(&models.Cliente{Nombre: "Maria Gomez", Empresa: "Dev INC"})

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
```

**3.Correr las pruebas unitarias:**Ejecutar desde consola.

Abre tu terminal y ejecuta el comando de pruebas de Go indicándole que busque en la carpeta `controllers`:

```bash
go test ./controllers -v
```

El flag `-v` (verbose) te mostrará un reporte detallado. Deberías ver una salida exitosa indicando `PASS` y el tiempo de ejecución (generalmente unos pocos milisegundos).

Con esto, el Junior puede automatizar la validación de su código. Cada vez que haga un cambio, solo corre `go test ./...` y sabrá de inmediato si rompió algo que antes funcionaba.


# Implementando POST protegido

Para probar un endpoint POST protegido, necesitamos simular dos cosas adicionales en nuestro "navegador falso" (`httptest`): enviar un cuerpo (body) en formato JSON e inyectar cabeceras HTTP, específicamente nuestro `X-API-Key`.

Además, como buena práctica, siempre debemos probar dos escenarios: el **camino feliz** (cuando enviamos la clave correcta) y el **camino de error** (cuando el middleware nos debe rechazar).

Abre tu archivo **`controllers/cliente_controller_test.go`** y añade el siguiente código debajo de la prueba que hicimos anteriormente.

**1.Actualizar los imports:**Importaciones adicionales.

Asegúrate de que tus importaciones en la parte superior del archivo se vean así, añadiendo `bytes` (para enviar el JSON), `os` (para variables de entorno) y tu carpeta de `middlewares`.

```go
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
```

**2.Prueba POST Exitosa:**El camino feliz (Status 201).

Agrega esta función para simular la creación de un cliente con el API Key correcto.

```go
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
```

**3.Prueba POST Fallida (Sin API Key):**Comprobando la seguridad (Status 401).

Para garantizar que nuestro middleware realmente funciona, haremos una prueba donde enviamos los datos correctamente, pero "olvidamos" enviar la cabecera `X-API-Key`.

```go
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
```

Al ejecutar en tu terminal el comando `go test -v ./controllers`, el Junior podrá ver que sus tres pruebas (`TestGetClientes`, `TestCreateCliente_Exito`, `TestCreateCliente_FalloAutenticacion`) pasan exitosamente, dándole la tranquilidad de que tanto la lógica de creación como la de seguridad funcionan a la perfección.

# Automatizar las pruebas

Automatizar las pruebas es lo que separa a un proyecto aficionado de uno profesional. Esto se conoce como **Integración Continua (CI)**. Con GitHub Actions, le diremos a GitHub que actúe como un robot que descarga nuestro código, instala Go y ejecuta las pruebas cada vez que alguien hace un `push`.

Aquí viene la mayor ventaja de la arquitectura que implementamos antes: **como usamos SQLite en memoria para las pruebas, no necesitamos configurar servidores de bases de datos complejos en GitHub**. El robot podrá correr las pruebas en segundos de forma aislada.

Aquí tienes el paso a paso para el desarrollador Junior:

**1.Crear la carpeta oculta de GitHub:**La estructura estricta.

GitHub busca automáticamente las instrucciones de automatización en una carpeta muy específica.

En la raíz de tu proyecto, crea una carpeta llamada `.github` y dentro de ella otra llamada `workflows`.

```bash
mkdir -p .github/workflows
```

**2.Crear el archivo YAML:**El archivo de instrucciones.

Dentro de la carpeta `workflows`, crea un archivo llamado `tests.yml`. El lenguaje YAML es muy sensible a los espacios (indentación), así que asegúrate de copiar el formato exacto.


```YAML
name: Pruebas Unitarias Go

# 1. ¿Cuándo se va a ejecutar este robot?
on:
  push:
    branches: [ "main", "master" ]
  pull_request: # También se ejecuta cuando alguien propone un cambio (PR)
    branches: [ "main", "master" ]

# 2. ¿Qué trabajos va a realizar?
jobs:
  test:
    name: Ejecutar Tests
    runs-on: ubuntu-latest # Una máquina virtual gratuita de Linux

    # 3. Paso a paso de lo que hará la máquina
    steps:
    - name: Clonar el repositorio
      uses: actions/checkout@v4

    - name: Configurar Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22' # Pon aquí la versión de Go que tengas en tu go.mod

    - name: Descargar dependencias
      run: go mod download

    - name: Ejecutar Pruebas Unitarias
      # ./... significa "busca pruebas en todas las carpetas del proyecto"
      run: go test -v ./...
```

**3.Hacer Push y ver la magia:**Subir los cambios.

Ahora, solo debes guardar este archivo, hacer un commit y subirlo a tu repositorio en GitHub.

```bash
git add .
git commit -m "Agrega pipeline de CI con GitHub Actions"
git push origin main
```

**4.Revisar en la plataforma:**Validación visual.

Una vez que hagas el push, entra a la página de tu repositorio en GitHub.com.

1. Haz clic en la pestaña superior que dice **"Actions"**.
2. Verás que hay un proceso corriendo (con un ícono amarillo de carga) llamado "Agrega pipeline de CI con GitHub Actions".
3. Si le das clic, podrás ver en tiempo real cómo la máquina de GitHub ejecuta los comandos.
4. Cuando termine, se pondrá en **verde (Success)** indicando que las pruebas pasaron, o en **rojo (Failed)** si alguna prueba falló o el código no compila.

> **La lección de oro para el Junior:** A partir de ahora, si alguien del equipo modifica un controlador y accidentalmente rompe la lógica de la API, GitHub pondrá una enorme X roja en su código antes de que pueda mezclarse con la rama principal, salvando al proyecto de un error en producción.


# Desplegar en Nube

Empaquetar la aplicación en un contenedor Docker es el paso definitivo para llevar el código a producción. La gran ventaja de Go es que compila en un único archivo binario, lo que nos permite usar una técnica avanzada pero muy fácil de entender llamada **Multi-stage Build** (Construcción en múltiples etapas).

Esta técnica usa una imagen "pesada" con las herramientas de Go para compilar el código, y luego copia _únicamente el binario resultante_ a una imagen "ultraligera" (Alpine) para ejecutarlo. Así logramos que nuestra API pase de pesar ~800MB a solo unos ~15MB, ideal para cualquier nube.

Aquí tienes el paso a paso:

**1.Crear el archivo .dockerignore:**Evitar basura en el contenedor.

Al igual que con `.gitignore`, no queremos que Docker copie nuestro archivo `.env` (las contraseñas se pasan diferente en la nube) ni carpetas innecesarias al momento de compilar.

En la raíz de tu proyecto, crea un archivo llamado `.dockerignore`:

```plaintext
.env
.git
.github
**/*_test.go
clientes.db
```

**2.Crear el Dockerfile:**La receta del contenedor.

En la misma carpeta raíz, crea un archivo llamado exactamente `Dockerfile` (sin extensión). Copia esta estructura explicada para el Junior:

```Dockerfile
# ==========================================
# ETAPA 1: Constructor (Builder)
# ==========================================
# Usamos la imagen oficial de Go basada en Alpine (ligera)
FROM golang:1.26.3-alpine AS builder

# Creamos y nos movemos a la carpeta /app dentro del contenedor
WORKDIR /app

# Copiamos primero los archivos de dependencias
COPY go.mod go.sum ./

# Descargamos las dependencias (Docker guardará esto en caché si no cambian)
RUN go mod download

# Copiamos el resto del código (incluyendo la carpeta docs de Swagger)
COPY . .

# Compilamos el binario. 
# CGO_ENABLED=0 asegura que sea un binario estático independiente.
RUN CGO_ENABLED=0 GOOS=linux go build -o api-main .

# ==========================================
# ETAPA 2: Producción (Imagen final)
# ==========================================
# Partimos de una imagen Alpine limpia (pesa apenas 5MB)
FROM alpine:latest

# Añadimos certificados de seguridad (por si hacemos peticiones HTTPS) 
# y tzdata (vital porque nuestro DSN de Postgres usa TimeZone=America/Caracas)
RUN apk --no-cache add ca-certificates tzdata

# Nos movemos a la carpeta raíz del usuario
WORKDIR /root/

# Copiamos ÚNICAMENTE el binario desde la "Etapa 1"
COPY --from=builder /app/api-main .

# Documentamos que el contenedor usará el puerto 8080
EXPOSE 8080

# Comando que se ejecutará al iniciar el contenedor
CMD ["./api-main"]
```

**3.Construir la Imagen:**Transformar código en imagen.

Abre tu terminal en la raíz del proyecto y ejecuta el comando de construcción. El punto `.` al final es crucial, ya que le indica a Docker que busque el `Dockerfile` en la carpeta actual.

```bash
docker build -t api-clientes-go .
```

_`api-clientes-go` es el nombre (etiqueta o tag) que le estamos dando a nuestra imagen._

**4.Ejecutar la API:**Probar el contenedor.

Una vez termine de construir, ya tenemos nuestra API empaquetada. Para ejecutarla simulando un entorno de producción, le pasaremos nuestro archivo `.env` local para que pueda conectarse a PostgreSQL.

```bash
docker run -p 8080:8080 --env-file .env api-clientes-go
```

> **Nota para el Junior:** En la nube (AWS, Google Cloud, Azure), no se sube el archivo `.env`. Las plataformas tienen gestores de "Secretos" o "Variables de Entorno" donde escribes estas claves desde una interfaz web, y el contenedor las lee automáticamente al arrancar gracias a que usamos `os.Getenv()`.
> 
>   

Con esto, el archivo binario compilado de Go se ejecuta de forma aislada. Si el Junior abre `http://localhost:8080/swagger/index.html`, verá que la API responde perfectamente, pero ahora está corriendo completamente encapsulada.

# Crear Docker-compose

Docker Compose es la herramienta definitiva para orquestar múltiples contenedores. En la vida real, una API rara vez vive sola; siempre necesita una base de datos, una caché (como Redis) o un servidor web. 

Esta es la sección de la guía donde el Junior aprenderá cómo conectar ambos mundos (Go y PostgreSQL) usando una red interna, de modo que con un solo comando se levante toda la infraestructura.

**1.Crear el archivo docker-compose.yml:**Definir la infraestructura como código.

En la raíz de tu proyecto (donde está el `Dockerfile`), crea un archivo llamado `docker-compose.yml`.

Copia esta estructura. Presta especial atención a los comentarios, ya que explican la magia detrás de la conexión:

```YAML
version: '3.8'

services:
  # 1. SERVICIO DE BASE DE DATOS
  db:
    image: postgres:15-alpine
    container_name: postgres-clientes
    environment:
      POSTGRES_USER: usuario
      POSTGRES_PASSWORD: secreto
      POSTGRES_DB: clientes_db
    ports:
      - "5432:5432"
    volumes:
      # Guarda los datos en un volumen persistente para no perderlos si se apaga el contenedor
      - pg_data:/var/lib/postgresql/data
    restart: unless-stopped

  # 2. SERVICIO DE LA API EN GO
  api:
    build: . # Le dice que busque el Dockerfile en esta misma carpeta y lo compile
    container_name: api-clientes-go
    ports:
      - "8080:8080"
    environment:
      # ¡ATENCIÓN JUNIOR! Aquí ocurre la magia de Docker.
      # Ya no usamos "localhost". Usamos el nombre del servicio ("db")
      # Docker tiene un DNS interno que resuelve "db" a la IP del contenedor de Postgres.
      DB_HOST: db
      DB_USER: usuario
      DB_PASSWORD: secreto
      DB_NAME: clientes_db
      DB_PORT: 5432
      API_KEY: secreto123
    depends_on:
      - db # Espera a que el contenedor 'db' se inicie antes de levantar la API
    restart: unless-stopped

# Declaramos el volumen que usamos en el servicio de la base de datos
volumes:
  pg_data:
```

**2.Entender el cambio de localhost a db:**El error más común del Junior.

Es vital explicarle al Junior por qué cambiamos `DB_HOST=localhost` por `DB_HOST=db`.

Cuando un programa corre dentro de un contenedor, su `localhost` es el interior de ese contenedor. Si la API busca la base de datos en su propio `localhost`, no encontrará nada y fallará. Al usar `docker-compose`, todos los servicios se unen a una misma red interna invisible, y el nombre del servicio (`db`) se convierte en su dirección IP.

**3.Levantar el entorno completo:**Un solo comando.

Abre tu terminal en la raíz del proyecto. Asegúrate de tener detenido tu servidor de Go local y cualquier contenedor de PostgreSQL que hayas levantado antes (para que los puertos 8080 y 5432 estén libres).

Ejecuta:

```bash
docker compose up --build -d
```

_Explicación de los flags:_

- `up`: Levanta los contenedores.
- `--build`: Fuerza a Docker a leer el `Dockerfile` y compilar el código de Go de nuevo (útil si hiciste cambios en el código).
- `-d` (Detached): Corre los contenedores en segundo plano, dejándote la terminal libre.

**4.Comandos útiles para el día a día:**Gestión diaria.

El Junior debe conocer estos comandos básicos para gestionar su nueva infraestructura:

```bash
# Ver si los contenedores están corriendo y en qué puertos
docker compose ps

# Ver los logs en tiempo real de la API (por si hay algún error)
docker compose logs -f api

# Apagar todos los contenedores sin borrar los datos de la base de datos
docker compose down

# Apagar todo y ELIMINAR la base de datos por completo (reseteo total)
docker compose down -v
```

Al ejecutar esto, el Junior tendrá un entorno idéntico al de producción ejecutándose en su máquina, eliminando para siempre el famoso problema de "en mi máquina sí funciona".

# Subir a Docker Hub

Subir tu imagen a **Docker Hub** es exactamente igual que subir tu código a GitHub, pero en lugar de subir el código fuente, subes la aplicación ya compilada y empaquetada. De esta forma, cualquier servidor de producción (como un VPS en AWS, DigitalOcean o Azure) solo tiene que descargarla y ejecutarla sin necesidad de instalar Go ni compilar nada.

Aquí tienes el paso a paso para explicarle al Junior cómo publicar su imagen:

**1.Iniciar sesión en Docker Hub:**Autenticación.

Primero, el Junior debe asegurarse de haber creado una cuenta gratuita en [hub.docker.com](https://hub.docker.com/).

Luego, en la terminal de su computadora, debe autenticarse ejecutando:

```bash
docker login
```

_La terminal le pedirá su nombre de usuario (Username) y su contraseña o token de acceso. Si el login es exitoso, verá el mensaje `Login Succeeded`._

**2.Etiquetar (Tag) la imagen:**La regla de oro del nombre.

Para que Docker sepa a qué cuenta subir la imagen, el nombre de la imagen **debe** comenzar estrictamente con el nombre de usuario de Docker Hub.

Nuestra imagen local actual se llama `api-clientes-go`. Supongamos que el usuario del Junior en Docker Hub es `juniordev`. Vamos a crear una "etiqueta" que apunte a nuestra imagen local.

Ejecuta el siguiente comando (cambiando `juniordev` por tu usuario real):

```bash
# Estructura: docker tag <imagen-local> <tu-usuario>/<nombre-repo>:<version>
docker tag api-clientes-go juniordev/api-clientes-go:v1.0
```

> **Nota para el Junior:** También es buena práctica subir una versión llamada `latest` (la más reciente). Puedes etiquetar la misma imagen dos veces:
> 
> `docker tag api-clientes-go juniordev/api-clientes-go:latest`
> 
>   

**3.Subir (Push) la imagen:**El despliegue a la nube.

Ahora que la imagen tiene el prefijo correcto de la cuenta, le ordenamos a Docker que la suba a los servidores públicos de Docker Hub.

```bash
docker push juniordev/api-clientes-go:v1.0
```

_(Y si creaste la etiqueta latest, súbela también con `docker push juniordev/api-clientes-go:latest`)_.

  Verás unas barras de carga mientras Docker sube las "capas" (layers) de tu imagen. Gracias a que usamos Alpine en nuestro archivo Dockerfile, esto tomará apenas unos segundos porque la imagen es muy liviana.

**4.Descargar en el Servidor (Pull):**Consumir la imagen en producción.

¡Listo! La imagen ya está en la nube.

Ahora, si el Junior entra a su servidor de producción (por ejemplo, por SSH a una máquina de Amazon EC2 o un Droplet), **ya no necesita el código fuente**. Solo necesita instalar Docker y correr:


```bash
docker pull juniordev/api-clientes-go:latest
```

Y para levantarla con su base de datos, simplemente usaría un archivo `docker-compose.yml` muy parecido al que creamos, pero cambiando la línea `build: .` por `image: juniordev/api-clientes-go:latest`.

> **Consejo de Arquitectura:** En entornos empresariales, las imágenes no se suben manualmente desde la computadora del desarrollador. Se configura el robot (GitHub Actions) para que él mismo compile y suba la imagen a Docker Hub automáticamente cada vez que el código principal es aprobado.


# Automatizar la subida a Docker Hub

Llevar tu automatización un paso más allá para que publique la imagen se conoce como **Entrega Continua (Continuous Delivery - CD)**.

Para lograr esto, le enseñaremos a nuestro "robot" de GitHub Actions a iniciar sesión en tu cuenta de Docker Hub. Como el código puede ser visto por otros, **jamás** escribiremos tu contraseña directamente en el archivo YAML. Usaremos los "Secretos" de GitHub.

**1.Configurar los Secretos en GitHub:**Seguridad de credenciales.

Antes de tocar el código, debes guardar tus credenciales de forma segura en la plataforma de GitHub.

1. Ve a la página de tu repositorio en GitHub.com.
2. Haz clic en la pestaña **Settings** (Configuración).
3. En el menú lateral izquierdo, busca la sección **Secrets and variables** y selecciona **Actions**.
4. Haz clic en el botón verde **New repository secret**.
5. Crea un primer secreto con el nombre `DOCKER_USERNAME` y en el valor pon tu usuario de Docker Hub (ej: `juniordev`).
6. Crea un segundo secreto con el nombre `DOCKER_PASSWORD` y en el valor pega tu contraseña de Docker Hub.

**2.Añadir el trabajo de Despliegue:**Modificar tests.yml.

Abre tu archivo `.github/workflows/tests.yml`. Vamos a añadir un segundo "job" (trabajo) debajo del que ya teníamos.

La palabra clave aquí es `needs: test`. Esto le dice a GitHub: _"Ni se te ocurra construir la imagen de Docker si las pruebas unitarias fallan"_.

Reemplaza todo el contenido del archivo con este nuevo código:

```YAML
name: Pruebas Unitarias y Despliegue

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  # ==========================================
  # TRABAJO 1: PRUEBAS (Integración Continua)
  # ==========================================
  test:
    name: Ejecutar Tests
    runs-on: ubuntu-latest
    steps:
    - name: Clonar el repositorio
      uses: actions/checkout@v4
      
    - name: Configurar Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'
        
    - name: Descargar dependencias
      run: go mod download
      
    - name: Ejecutar Pruebas Unitarias
      run: go test -v ./...

  # ==========================================
  # TRABAJO 2: DOCKER (Entrega Continua)
  # ==========================================
  build-and-push:
    name: Construir y Subir a Docker Hub
    needs: test # ¡LA MAGIA! Solo arranca si "test" terminó en verde
    runs-on: ubuntu-latest
    
    # Opcional pero recomendado: Solo publicar si el cambio ocurre en la rama 'main'
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Clonar el repositorio
      uses: actions/checkout@v4

    - name: Iniciar sesión en Docker Hub
      uses: docker/login-action@v3
      with:
        # Aquí GitHub inyecta los secretos que guardaste en el Paso 1
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Construir y Subir imagen
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        # Etiquetamos la imagen dinámicamente con tu usuario
        tags: ${{ secrets.DOCKER_USERNAME }}/api-clientes-go:latest
```

**3.Ver la automatización en acción:**Confirmar y empujar.

Guarda los cambios, haz un nuevo commit y empújalo a tu rama principal:

```bash
git add .
git commit -m "Agrega publicación automática de Docker Hub"
git push origin main
```

Si vas a la pestaña **Actions** en GitHub, verás un diagrama de flujo profesional: primero corre la caja de `test`, y cuando esta termina exitosamente, se activa la caja de `build-and-push`. ¡Si alguien sube código que rompe una prueba, la imagen defectuosa nunca llegará a Docker Hub!

# Mejora seguridad Docker Hub

Excelente pregunta. En el ámbito de la gobernanza de seguridad de la información, usar la contraseña maestra en flujos automatizados rompe uno de los pilares fundamentales del control de accesos: el **Principio de Menor Privilegio (PoLP)**.

Para que el desarrollador Junior entienda por qué las arquitecturas empresariales y las normativas de cumplimiento (como PCI DSS o ISO 27001) prohíben esta práctica, explícale estos dos riesgos críticos:

1. **Superficie de ataque y control total:** La contraseña real otorga privilegios administrativos absolutos. Si un atacante logra extraer ese secreto de GitHub, no solo podría subir imágenes maliciosas (envenenando la cadena de suministro), sino que podría borrar repositorios enteros, alterar configuraciones de seguridad o secuestrar la cuenta cambiando el correo de recuperación.
2. **Trazabilidad y revocación:** Si sospechas que una contraseña ha sido expuesta, debes cambiarla inmediatamente. Esto rompe instantáneamente cualquier otro servidor, script o automatización que dependiera de ella. Un _Access Token_, por el contrario, es desechable: si se filtra el token de GitHub Actions, simplemente lo revocas en Docker Hub y creas uno nuevo, sin afectar tu acceso personal ni otros sistemas.

Aquí tienes el paso a paso para implementar esta mejora de seguridad:

**1.Generar el Access Token en Docker Hub:**Creación controlada.

Primero, vamos a emitir una credencial que solo sirva para interactuar con las imágenes, sin permisos administrativos.

1. Entra a tu cuenta en [Docker Hub](https://hub.docker.com/).
2. Haz clic en tu foto de perfil (arriba a la derecha) y selecciona **Account settings**.
3. En el menú lateral izquierdo, haz clic en **Security** y luego en el botón azul **New Access Token**.
4. **Access Token Description:** Ponle un nombre descriptivo para la auditoría, por ejemplo: `GitHub Actions CI-CD`.
5. **Access permissions:** Aquí aplicamos el Principio de Menor Privilegio. Cambia el permiso a **Read & Write**. (No le des permisos de "Delete" a menos que tu pipeline lo requiera estrictamente).
6. Haz clic en **Generate**.

**2.Copiar el Token:**Respaldo seguro.

La pantalla te mostrará una cadena larga de texto (que suele empezar por `dckr_pat_`).

**Cópiala inmediatamente**. Por seguridad, Docker Hub jamás te volverá a mostrar este token completo una vez que cierres esa ventana. Si lo pierdes, tendrás que revocarlo y crear uno nuevo.

**3.Actualizar los Secretos en GitHub:**Rotación de credenciales.

Ahora debemos reemplazar la contraseña vulnerable por nuestra nueva credencial segura.

1. Ve a tu repositorio en GitHub y entra a **Settings**.
2. Navega a **Secrets and variables** > **Actions**.  
3. Busca el secreto que creamos anteriormente llamado `DOCKER_PASSWORD` y haz clic en el ícono del lápiz (Editar).
4. Borra tu contraseña real, pega el nuevo **Access Token** de Docker Hub y guarda los cambios.

El archivo `.yml` de GitHub Actions no necesita ninguna modificación. La acción `docker/login-action` es lo suficientemente inteligente como para aceptar un Access Token en el campo `password`.

A partir de ahora, el pipeline del Junior opera bajo un modelo de confianza cero mucho más robusto y alineado con los estándares de la industria.

# Protección Ramas en Github

Esta es una de las configuraciones más críticas. En el mundo del gobierno de la seguridad de la información, bloquear los "commits directos" a la rama de producción implementa un control vital: la **Segregación de Funciones (Segregation of Duties)**. De hecho, implementar este flujo de revisión cruzada y pruebas automatizadas es un control de gestión de cambios obligatorio si el proyecto alguna vez necesita cumplir con marcos como PCI DSS o ISO 27001.

Para el Junior, esto significa que ya no podrá hacer `git push origin main`. Tendrá que crear una rama paralela (ej. `feature/nueva-ruta`), subir sus cambios ahí y solicitar permiso para unirlos mediante un **Pull Request (PR)**.

Aquí tienes el paso a paso para configurar esta barrera en GitHub:

**1.Navegar a las Reglas de Rama:**Acceder como administrador.

1. Ve a la página principal de tu repositorio en GitHub.
2. Haz clic en la pestaña superior derecha **Settings** (Configuración).
3. En el menú lateral izquierdo, bajo la sección "Code and automation", selecciona **Branches** (Ramas).
4. Haz clic en el botón verde **Add branch protection rule** (Añadir regla de protección de rama).

**2.Definir el objetivo:**Branch name pattern.

En el primer campo llamado **Branch name pattern**, escribe exactamente el nombre de la rama que quieres proteger.

Escribe: `main` (o `master` si tu repositorio usa el nombre antiguo). Esto le dice a GitHub que las siguientes reglas solo aplican a la rama de producción.

**3.Exigir un Pull Request:**El principio de los 4 ojos.

Busca la sección "Protect matching branches" y marca la primera casilla:

**"Require a pull request before merging"** (Requerir un pull request antes de fusionar).

Al marcarla, se desplegarán sub-opciones. Asegúrate de configurar esto:

- **Require approvals:** Actívalo y pon el número en **1**. Esto significa que al menos un desarrollador distinto al autor del código (en este caso, tú) debe revisar y aprobar los cambios antes de que el botón de _Merge_ se habilite.

**4.Exigir que las pruebas pasen (Status Checks):**Integración Continua obligatoria.

Más abajo, marca la casilla:

**"Require status checks to pass before merging"** (Requerir que las comprobaciones de estado pasen antes de fusionar).

Esta es la conexión con nuestro archivo de GitHub Actions.

1. Marca también la sub-opción **"Require branches to be up to date before merging"**.
2. En el buscador que aparece justo debajo, escribe el nombre del _job_ que definimos en el archivo YAML (si usaste el ejemplo anterior, busca y selecciona `test`).

_Resultado:_ GitHub bloqueará el botón de _Merge_ hasta que la caja de pruebas de nuestro robot se ponga en verde.

**5.Bloquear a los administradores (Opcional):**Disciplina estricta.

Casi al final de la página, hay una casilla llamada:

**"Do not allow bypassing the above settings"** (No permitir saltarse las configuraciones anteriores).

Si marcas esto, ni siquiera tú como administrador o dueño del repositorio podrás hacer un push directo a `main` para meter un "hotfix" rápido. Tendrás que seguir el mismo proceso de Pull Request que el Junior. Es una excelente práctica de disciplina técnica.

Finalmente, haz clic en el botón verde **Create** o **Save changes** al final de la página.
### El nuevo flujo de trabajo del Junior

Con esto configurado, el flujo diario para enseñar al Junior será el siguiente:

1. **Desarrollo:** Hace sus cambios en una rama local (`git checkout -b feature/nuevo-endpoint`).
2. **Push:** Sube su rama a GitHub (`git push origin feature/nuevo-endpoint`).
3. **Solicitud:** Entra a GitHub y abre un **Pull Request** apuntando hacia `main`.
4. **Validación Automática:** El robot de GitHub Actions corre las pruebas de Go (Status Checks).
5. **Revisión Humana:** Tú revisas el código, dejas comentarios si es necesario, y apruebas.
6. **Merge:** Solo con el visto bueno del robot y el tuyo, los cambios entran a `main` y se dispara el despliegue a Docker Hub.

# Plantilla de revisión

Implementar una plantilla de Pull Request es una excelente estrategia metodológica para enseñar disciplina y estandarizar la calidad del código. Funciona como un recordatorio visual que evita la omisión de pasos críticos antes de la revisión, reduciendo la fricción durante el proceso de _Code Review_.

Aquí tienes el paso a paso para configurarlo:

**1.Crear el archivo en la carpeta .github:**Ubicación estándar.

En la raíz de tu proyecto ya tenemos la carpeta oculta `.github` (donde ubicamos los workflows de Actions).

Dentro de esa misma carpeta, crea un archivo llamado exactamente `PULL_REQUEST_TEMPLATE.md`.

**2.Redactar la plantilla de revisión:**Estructura Markdown.

Abre el archivo y pega el siguiente contenido. GitHub soporta la sintaxis `- [ ]` para crear casillas de verificación interactivas. Esta plantilla está adaptada específicamente al stack tecnológico que hemos construido (Go, Gin, Swagger, Docker):

```Markdown
## 📝 Descripción del Cambio
<!-- Describe brevemente qué hace este Pull Request y por qué es necesario. -->

## 🔗 Ticket o Issue relacionado
<!-- Si aplica, coloca el número del issue. Ej: Closes #12 -->

## ✅ Checklist de Calidad para el Junior
<!-- Marca con una 'x' los pasos que ya completaste: [x] -->

- [ ] **Compilación:** El código compila localmente sin errores.
- [ ] **Pruebas:** Escribí pruebas unitarias para la nueva lógica y ejecuté `go test ./...`.
- [ ] **Swagger:** Actualicé las anotaciones (`@Summary`, `@Param`, etc.) y ejecuté `swag init`.
- [ ] **Seguridad:** No dejé contraseñas, tokens ni API Keys en el código fuente.
- [ ] **Entorno:** Si agregué nuevas variables, las documenté en `.env.example`.
- [ ] **Docker:** Confirmé que el proyecto sigue levantando correctamente con `docker compose up`.

## 📸 Evidencia (Opcional)
<!-- Adjunta capturas de pantalla de la terminal con las pruebas en verde o de la interfaz de Swagger. -->
```

**3.Subir la plantilla a main:**Despliegue de la regla.

Guarda el archivo y súbelo a tu repositorio. Como esta es una configuración administrativa del repositorio, debes subirla directamente a la rama principal:

```bash
git add .github/PULL_REQUEST_TEMPLATE.md
git commit -m "Agrega plantilla de checklist obligatoria para Pull Requests"
git push origin main
```

### La experiencia del Junior

A partir de este momento, cada vez que el desarrollador Junior suba una nueva rama (por ejemplo, `feature/nuevo-endpoint`) y haga clic en el botón verde **Compare & pull request** en GitHub, el cuadro de descripción ya no estará vacío.  

Aparecerá pre-llenado con tu formato. Para enviar la solicitud, él deberá completar la descripción y marcar las casillas cambiando los corchetes vacíos `[ ]` por corchetes con una equis `[x]`. Si durante tu revisión notas que marcó "Actualicé Swagger" pero los archivos generados no están en el commit, tienes una base objetiva para rechazar el PR y pedir correcciones.