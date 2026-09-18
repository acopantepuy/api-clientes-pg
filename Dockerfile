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