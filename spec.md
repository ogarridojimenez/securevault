# SecureVault — Especificación del Proyecto

## 🎯 Propósito
API REST + CLI para gestión de secretos y credenciales. Un backend real, con todo lo que espera un reclutador de un desarrollador Go senior: arquitectura limpia, tests, CI/CD, Docker, y documentación profesional.

**No es un proyecto de ciberseguridad** — es un proyecto de INGENIERÍA DE SOFTWARE cuyo dominio resulta ser la gestión de secretos. Como construir un sistema bancario o un SaaS.

---

## 🏗️ Arquitectura (Hexagonal / Clean Architecture)

```
securevault/
├── cmd/
│   ├── api/main.go              # Servidor HTTP (REST API)
│   └── vault/main.go            # CLI tool
├── internal/
│   ├── core/                     # Capa de dominio (entidades, puertos)
│   │   ├── model/
│   │   │   ├── secret.go         # Entidad Secret
│   │   │   ├── vault.go          # Entidad Vault (contenedor de secrets)
│   │   │   ├── user.go           # Entidad User
│   │   │   └── apikey.go         # API Key
│   │   └── port/                 # Interfaces (puertos)
│   │       ├── repository.go     # Port: almacenamiento
│   │       ├── crypto.go         # Port: cifrado/descifrado
│   │       └── auth.go           # Port: autenticación
│   ├── service/                  # Casos de uso (orquestación)
│   │   ├── vault_service.go      # CRUD de vaults
│   │   ├── secret_service.go     # CRUD de secrets (cifrados)
│   │   ├── user_service.go       # Registro, login, perfil
│   │   └── apikey_service.go     # Gestión de API keys
│   ├── handler/                  # Handlers HTTP (adaptadores entrada)
│   │   ├── vault_handler.go      # /api/v1/vaults
│   │   ├── secret_handler.go     # /api/v1/vaults/{id}/secrets
│   │   ├── auth_handler.go       # /api/v1/auth (login, register)
│   │   └── health_handler.go     # /health, /ready
│   ├── repository/               # Implementaciones DB (adaptadores salida)
│   │   ├── postgres_vault.go     # PostgreSQL: vaults
│   │   ├── postgres_secret.go    # PostgreSQL: secrets (cifrados)
│   │   ├── postgres_user.go      # PostgreSQL: users
│   │   └── migrations/           # Migraciones SQL
│   │       ├── 001_users.sql
│   │       ├── 002_vaults.sql
│   │       └── 003_secrets.sql
│   ├── crypto/                   # Implementación cifrado
│   │   ├── aesgcm.go             # AES-256-GCM para secrets
│   │   ├── argon2.go             # Argon2id para hashing passwords
│   │   └── jwt.go                # JWT access + refresh tokens
│   ├── middleware/               # Middlewares HTTP
│   │   ├── auth.go               # JWT validation
│   │   ├── rbac.go               # Role-Based Access Control
│   │   ├── ratelimit.go          # Rate limiting por API key
│   │   ├── logging.go            # Request logging estructurado
│   │   └── cors.go               # CORS configurable
│   └── config/                   # Configuración
│       └── config.go             # Env vars + defaults
├── pkg/
│   └── cli/                      # CLI commands (Cobra)
│       ├── root.go               # Comando raíz
│       ├── login.go              # vault login
│       ├── get.go                # vault get <path>
│       ├── set.go                # vault set <path> <value>
│       ├── list.go               # vault list <path>
│       ├── delete.go             # vault delete <path>
│       └── export.go             # vault export (formato portátil)
├── api/
│   └── openapi.yaml              # Especificación OpenAPI 3.1
├── Dockerfile                    # Multi-stage (api + cli)
├── docker-compose.yml            # API + PostgreSQL + Adminer
├── Makefile                      # Comandos comunes
├── .github/workflows/
│   ├── ci.yml                    # Tests + lint + build
│   └── release.yml               # Release automatizado
├── docs/
│   ├── architecture.md           # Decisiones técnicas
│   └── examples.md               # Ejemplos de uso
└── README.md                     # Badges, quickstart, demo
```

---

## ⚙️ API REST (16 endpoints)

### Autenticación
| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Registrar usuario |
| POST | `/api/v1/auth/login` | Login → JWT pair |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/logout` | Invalidar refresh token |

### Vaults (contenedores)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/vaults` | Listar vaults del usuario |
| POST | `/api/v1/vaults` | Crear vault |
| GET | `/api/v1/vaults/{id}` | Obtener vault |
| PUT | `/api/v1/vaults/{id}` | Actualizar vault |
| DELETE | `/api/v1/vaults/{id}` | Eliminar vault |

### Secrets (dentro de un vault)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/vaults/{id}/secrets` | Listar secrets (solo nombres y metadatos, valores cifrados) |
| POST | `/api/v1/vaults/{id}/secrets` | Crear secret (el valor se cifra antes de guardar) |
| GET | `/api/v1/vaults/{id}/secrets/{secretId}` | Obtener secret (descifrado en servidor) |
| PUT | `/api/v1/vaults/{id}/secrets/{secretId}` | Actualizar secret |
| DELETE | `/api/v1/vaults/{id}/secrets/{secretId}` | Eliminar secret |

### API Keys (para integraciones)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/api-keys` | Listar API keys |
| POST | `/api/v1/api-keys` | Crear API key |
| DELETE | `/api/v1/api-keys/{id}` | Revocar API key |

### Health
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | Health check simple |
| GET | `/ready` | Readiness (DB conectada) |

---

## 💻 CLI (Cobra)

```bash
# Autenticación
vault login                      # Login interactivo
vault login --token <token>      # Login con token existente

# Gestión de secrets
vault set production/db/password "SuperSecret123!"
vault get production/db/password
vault list production/
vault delete production/db/password

# Export/Import
vault export --format json > backup.json
vault import backup.json

# Administración
vault api-key create "CI/CD Pipeline"
vault api-key revoke <key-id>

# Utilidades
vault version
vault help
```

---

## 🛠️ Stack Tecnológico

| Componente | Tecnología | Por qué |
|------------|-----------|---------|
| **Lenguaje** | Go 1.23+ | Rendimiento, tipado fuerte, concurrencia |
| **Router HTTP** | chi (go-chi/chi) | Ligero, middlewares, compatible net/http |
| **CLI** | Cobra + Viper | Estándar en Go, flags + config file |
| **DB** | PostgreSQL 16 | Migraciones, JSONB, tipos robustos |
| **Migrations** | golang-migrate | Versionadas, idempotentes |
| **Cifrado** | crypto/aes + crypto/argon2 | AES-256-GCM para datos, Argon2id para passwords |
| **Auth** | golang-jwt/jwt/v5 | JWT access (15min) + refresh (7d) |
| **Testing** | testing + testify + httptest | Tests unitarios, integración, contratos |
| **API Spec** | OpenAPI 3.1 + swagger-ui | Documentación endpoints |
| **CI/CD** | GitHub Actions | Lint → Test → Build → Release |
| **Container** | Docker multi-stage + Compose | Desarrollo + producción |
| **Logging** | slog (stdlib) + zerolog | Log estructurado, niveles |
| **Validation** | go-playground/validator | Validación de inputs |

### go.mod (dependencias clave)
```go
require (
    github.com/go-chi/chi/v5 v5.2.x
    github.com/spf13/cobra v1.9.x
    github.com/spf13/viper v1.19.x
    github.com/golang-jwt/jwt/v5 v5.2.x
    github.com/golang-migrate/migrate/v4 v4.18.x
    github.com/go-playground/validator/v10 v10.23.x
    github.com/lib/pq v1.10.x
    github.com/stretchr/testify v1.10.x
    github.com/rs/zerolog v1.33.x
)
```

---

## 🗄️ Modelo de Datos (PostgreSQL)

```sql
-- Tabla: users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,     -- Argon2id
    display_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- user | admin
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tabla: vaults (contenedores de secrets)
CREATE TABLE vaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50) DEFAULT 'lock',          -- Para UI si se agrega después
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, name)
);

-- Tabla: secrets (valores cifrados)
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
    path VARCHAR(500) NOT NULL,               -- ej: "production/db/password"
    encrypted_value BYTEA NOT NULL,           -- AES-256-GCM cifrado
    metadata JSONB DEFAULT '{}',              -- metadatos descriptivos
    version INT NOT NULL DEFAULT 1,           -- versionado de secrets
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(vault_id, path)
);

-- Tabla: api_keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(64) NOT NULL,            -- SHA-256 de la API key
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE
);

-- Tabla: refresh_tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,           -- SHA-256 del refresh token
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE
);

-- Índices
CREATE INDEX idx_secrets_vault ON secrets(vault_id);
CREATE INDEX idx_secrets_path ON secrets(vault_id, path);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

---

## 🔐 Flujo de Cifrado

```
          SET                         GET
     ┌──────────┐              ┌──────────┐
     │  Usuario  │              │  Usuario  │
     └────┬─────┘              └────▲──────┘
          │ "pass123"               │ "pass123" (descifrado)
          ▼                         │
     ┌──────────┐              ┌────┴──────┐
     │API Server│              │API Server │
     └────┬─────┘              └────▲──────┘
          │ AES-256-GCM             │ AES-256-GCM
          │ con master key          │ con master key
          ▼                         │
     ┌──────────┐              ┌────┴──────┐
     │PostgreSQL│              │PostgreSQL │
     │ [BYTEAS] │              │ [BYTEAS]  │
     └──────────┘              └───────────┘
```

- **Master Key:** Generada por vault, cifrada con la password del usuario
- **Cada secret** se cifra con AES-256-GCM (autenticado)
- **Nonce único** por cada cifrado
- **Passwords** hasheadas con Argon2id (memoria ~64MB, iteraciones ~3)
- **Nunca se almacenan** valores en texto plano en DB

---

## 📋 Plan de Implementación (4 semanas)

| Semana | Módulo | Entregable |
|--------|--------|------------|
| **1** | Proyecto + DB + Auth | go mod init, chi server, migrations, register/login, JWT, tests |
| **2** | Vaults + Secrets API | CRUD vaults, CRUD secrets (cifrados), middleware auth, validator |
| **3** | CLI + API Keys | Cobra commands (set/get/list/delete), API keys, RBAC, rate limiting |
| **4** | CI/CD + Docker + Docs | GitHub Actions, multi-stage Docker, OpenAPI spec, README, benchmarks |

---

## ✅ Lo que aporta a tu perfil de DESARROLLADOR

| Habilidad | Cómo la demuestra SecureVault |
|-----------|------------------------------|
| **Arquitectura Hexagonal** | core/service/handler/repository separado por capas |
| **Go avanzado** | Interfaces, genéricos, context, middlewares |
| **APIs REST** | 16 endpoints, OpenAPI spec, versionado |
| **Autenticación** | JWT pair (access + refresh), Argon2id, API keys |
| **Base de datos** | PostgreSQL, migraciones, índices, JSONB |
| **Cifrado** | AES-256-GCM, sin CGO, nonce management |
| **CLI profesional** | Cobra + Viper, flags, subcomandos |
| **Testing** | Unit + integration + httptest |
| **CI/CD** | GitHub Actions, lint, test, build, release |
| **Docker** | Multi-stage, docker-compose dev |
| **Documentación** | OpenAPI, architecture docs, README |

---

## 📂 Mock para desarrollo

```go
// Test server simula PostgreSQL en memoria (para tests)
POST /api/v1/auth/register  {"email":"test@test.com","password":"secure123"}
     → 201 {"user_id":"...", "access_token":"...", "refresh_token":"..."}

POST /api/v1/auth/login     {"email":"test@test.com","password":"secure123"}
     → 200 {"access_token":"...", "refresh_token":"..."}

POST /api/v1/vaults         {"name":"production","description":"Prod secrets","icon":"server"}
     → 201 {"id":"...", "name":"production", ...}

POST /api/v1/vaults/{id}/secrets {"path":"db/password","value":"SuperSecret123!","metadata":{"env":"prod"}}
     → 201 {"id":"...", "path":"db/password", "version":1, "created_at":"..."}

GET  /api/v1/vaults/{id}/secrets/{secretId}
     → 200 {"id":"...", "path":"db/password", "value":"SuperSecret123!", "version":1}
```

---

## 🚀 Ejemplo completo de uso

```bash
# 1. Iniciar servicios
docker-compose up -d

# 2. Registrar usuario
curl -X POST localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Str0ng!Pass"}' | jq .

# 3. Login (guardar token)
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Str0ng!Pass"}' | jq -r '.access_token')

# 4. Crear vault
VAULT_ID=$(curl -s -X POST localhost:8080/api/v1/vaults \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"production","description":"Servidores producción"}' | jq -r '.id')

# 5. Guardar secret
curl -s -X POST "localhost:8080/api/v1/vaults/$VAULT_ID/secrets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"db/password","value":"p@ssw0rd_pr0d"}' | jq .

# 6. Leer secret (descifrado automático)
curl -s "localhost:8080/api/v1/vaults/$VAULT_ID/secrets" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 7. O usando CLI (después de compilar)
vault login --token "$TOKEN"
vault set production/db/password "p@ssw0rd_pr0d"
vault get production/db/password
```
