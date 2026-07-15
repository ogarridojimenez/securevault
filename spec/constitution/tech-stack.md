# Tech Stack — SecureVault

> Stack tecnológico y convenciones del proyecto.

## Tecnologías

| Componente | Tecnología | Propósito |
|-----------|-----------|-----------|
| Lenguaje | **Go 1.23+** (1.26.5 real) | Rendimiento, concurrencia, despliegue simple |
| Router HTTP | **chi v5** (`github.com/go-chi/chi/v5`) | Ligero, composable, middleware nativo |
| CLI | **Cobra v1.9.1** + **Viper v1.20.1** | Comandos + configuración |
| Base de datos | **PostgreSQL 16** + `pgx v5` | Persistencia, JSONB, UUIDs |
| Auth | **JWT** (`golang-jwt/jwt/v5`) + **Argon2id** (stdlib `crypto/argon2`) | Autenticación |
| Cifrado | **AES-256-GCM** (stdlib `crypto/aes` + `crypto/cipher`) | Cifrado de secretos |
| Tests | **testify** + `httptest` + `testcontainers-go` | Unit, integración, e2e |
| CI/CD | **GitHub Actions** | Lint → Test → Build → Release |
| Docker | **Multi-stage** (golang:1.23-alpine → alpine:3.20) | Imagen final ~15MB |
| Documentación | **OpenAPI 3.1** (YAML) | Documentación de API |
| Logging | `log/slog` (stdlib) | Logging estructurado |

## Comandos

```bash
go run ./cmd/server          # Arranca servidor API
go run ./cmd/vault           # CLI
go test ./...                # Tests
go vet ./...                 # Análisis estático
go build -o vault.exe ./cmd/vault  # Build CLI
go build -o server.exe ./cmd/server # Build servidor
golangci-lint run            # Linting completo
```

## Convenciones

- **Naming:** `camelCase` en Go (público → `UpperCase`, privado → `lowerCase`)
- **Estructura hexagonal:** `cmd/` (entrypoints), `internal/core/` (dominio), `internal/service/` (casos de uso), `internal/handler/` (transport HTTP), `internal/repository/` (persistencia)
- **Manejo de errores:** Errores envueltos con `fmt.Errorf("contexto: %w", err)`. Nunca `panic` en handlers.
- **Idioma:** Código en inglés (variables, funciones, comentarios). Documentación SDD en español.
- **Validación:** Input validation siempre en handler/service boundary. Nunca confiar en datos del cliente.
- **SQL:** Migraciones con archivos `.sql` planos (sin ORM). Consultas parametrizadas siempre (`$1`, `$2`).

## Límites duros

| Regla | Explicación |
|-------|-------------|
| No añadir dependencias externas sin aprobación | Evaluar primero si stdlib lo cubre |
| No usar `panic/recover` en lógica de negocio | Solo en `main()` como safety net |
| No subir `.env`, `*.key`, `*.pem` al repo | Secrets en variables de entorno o `.env.example` |
| No mutar `context.Background()` | Pasar context siempre; propagar cancelación |
| No hardcodear secrets en código | Config vía entorno o archivo `.env` |
| No usar `database/sql` directamente | Preferir `pgx` que es tipo-seguro y más rápido |
