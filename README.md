# SecureVault

> **Proyecto portfolio de backend Go — Arquitectura hexagonal real con API REST + CLI profesional.**

No es un producto de ciberseguridad. Es un **backend completo y bien construido** que demuestra todo lo que se espera de un desarrollador Go senior: diseño limpio, tests, CI/CD, Docker, documentación profesional y Go avanzado. El dominio de "gestión de secretos" es solo el vehículo para mostrar estas habilidades.

---

## 🎯 Objetivo del proyecto

**¿Para qué sirve SecureVault?**

Para demostrar que sé construir un backend profesional desde cero:

| Habilidad | Cómo se demuestra aquí |
|-----------|----------------------|
| **Go avanzado** | Interfaces, context, `crypto/` stdlib (AES, Argon2id), graceful shutdown, middlewares |
| **Arquitectura hexagonal** | 5 capas separadas: `model/` → `repository/` → `service/` → `handler/` → `middleware/` |
| **API REST real** | Chi router, 16 endpoints, errores tipados, respuestas JSON consistentes |
| **Auth profesional** | JWT pair (access 15m + refresh 7d) con rotation, Argon2id, API keys con SHA-256 |
| **PostgreSQL real** | Migraciones, UUIDs, JSONB, índices, pgx v5 connection pool |
| **Cifrado real** | AES-256-GCM sin CGO, cada secreto con nonce único, clave maestra rotable |
| **CLI profesional** | Cobra + Viper, 8 comandos, persistencia de token |
| **Docker** | Multi-stage build, docker-compose con PostgreSQL |
| **CI/CD** | GitHub Actions: lint → test → build → release |
| **Documentación** | OpenAPI 3.1, README completo, SDD artifacts |
| **Metodología** | Spec-Driven Development (SDD) — spec primero, código después |

---

## Stack tecnológico

| Componente | Tecnología | Versión |
|-----------|------------|---------|
| Lenguaje | Go | 1.23+ |
| API Router | chi | v5 |
| CLI | Cobra + Viper | v1.10 + v1 |
| Base de datos | PostgreSQL | 16 |
| Driver DB | pgx | v5 |
| JWT | golang-jwt | v5 (HMAC) |
| Cifrado | AES-256-GCM | crypto/aes + crypto/cipher |
| Passwords | Argon2id | golang.org/x/crypto |
| API Keys | SHA-256 hash | crypto/sha256 |
| CI/CD | GitHub Actions | lint + test + build |
| Contenedores | Docker | multi-stage alpine |

---

## Arquitectura

```
           ┌─────────────────────────────────────┐
           │         CLI (Cobra)                 │
           │  vault get/set/list/delete          │
           └──────────────┬──────────────────────┘
                          │ HTTP
           ┌──────────────▼──────────────────────┐
           │      API (Chi Router)               │
           │  /api/v1/*                          │
           ├─────────────────────────────────────┤
           │      Middleware                     │
           │  JWT Auth · Rate Limiter · CORS    │
           └──────────────┬──────────────────────┘
                          │
           ┌──────────────▼──────────────────────┐
           │      Service Layer                  │
           │  Auth · Vaults · Secrets · API Keys │
           └──────────────┬──────────────────────┘
                          │
           ┌──────────────▼──────────────────────┐
           │      Repository Layer               │
           │  PostgreSQL (pgx v5)                │
           └──────────────┬──────────────────────┘
                          │
              ┌───────────▼───────────┐
              │   PostgreSQL 16       │
              │   (datos cifrados)    │
              └───────────────────────┘
```

---

## Metodología: Spec-Driven Development

Este proyecto se construyó siguiendo **SDD híbrido** (Spec-Driven Development). Los artefactos de especificación están versionados junto al código en `spec/`:

```
spec/
├── constitution/
│   ├── mission.md       — Propósito y principios
│   ├── tech-stack.md    — Stack tecnológico
│   └── roadmap.md       — Features priorizadas
└── features/
    └── 001-securevault-core/
        ├── spec.md       — Requisitos funcionales (EARS)
        ├── plan.md       — Arquitectura y decisiones técnicas
        └── tasks.md      — Tareas atómicas verificables
```

La especificación es el artefacto primario. El código es un artefacto derivado.

---

## Quick Start

### Requisitos

- Go 1.23+
- PostgreSQL 16
- Docker (opcional)

### Configuración

```bash
# 1. Clonar
git clone https://github.com/ogarridojimenez/securevault.git
cd securevault

# 2. Clave de cifrado (32 bytes hex)
python3 -c "import secrets; print(secrets.token_hex(32))"

# 3. Copiar y configurar
cp config.example.yaml config.yaml
# Editar config.yaml:
#   db.url          → postgres://user:pass@localhost:5432/securevault
#   jwt.secret      → string de 32+ caracteres
#   encryption.key  → 64 caracteres hex (de paso 2)
```

### Base de datos

```bash
createdb securevault
psql -d securevault -f migrations/001_initial_schema.sql
```

### Ejecutar

```bash
# Servidor (API REST)
go run ./cmd/server/

# CLI
go run ./cmd/vault/ --help
```

### O con Docker

```bash
docker compose up -d
```

---

## API Endpoints (16)

### Auth
| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Registro de usuario |
| POST | `/api/v1/auth/login` | Inicio de sesión |
| POST | `/api/v1/auth/refresh` | Renovar access token |
| POST | `/api/v1/auth/logout` | Cerrar sesión |

### Vaults
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/vaults` | Listar vaults |
| POST | `/api/v1/vaults` | Crear vault |
| GET | `/api/v1/vaults/{id}` | Obtener vault |
| PUT | `/api/v1/vaults/{id}` | Actualizar vault |
| DELETE | `/api/v1/vaults/{id}` | Eliminar vault |

### Secrets (cifrados AES-256-GCM)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/vaults/{id}/secrets` | Listar secrets (sin valores) |
| POST | `/api/v1/vaults/{id}/secrets` | Crear/actualizar secret |
| GET | `/api/v1/vaults/{id}/secrets/{sid}` | Obtener valor descifrado |
| DELETE | `/api/v1/vaults/{id}/secrets/{sid}` | Eliminar secret |

### API Keys
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/api-keys` | Listar API keys |
| POST | `/api/v1/api-keys` | Crear API key |
| PUT | `/api/v1/api-keys/{id}/revoke` | Revocar API key |

### Health
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/readiness` | Readiness check |

---

## CLI (8 comandos)

```bash
vault login                           # Autenticarse (email + password)
vault register                        # Crear cuenta
vault set <vault/name> <value>        # Guardar secreto
vault get <vault/name>                # Leer secreto
vault list <vault>                    # Listar secrets de un vault
vault delete <vault/name>             # Eliminar secreto
vault export                          # Exportar todos los secrets (JSON)
vault import [file]                   # Importar secrets (JSON)
vault api-key create <name>           # Crear API key
vault api-key list                    # Listar API keys
vault api-key revoke <id>             # Revocar API key
```

---

## Modelo de datos (5 tablas)

| Tabla | Propósito | Campos clave |
|-------|-----------|-------------|
| `users` | Usuarios del sistema | email (UNIQUE), password_hash (Argon2id) |
| `vaults` | Contenedores de secrets | user_id (FK), name (UNIQUE por user) |
| `secrets` | Secrets cifrados | ciphertext + nonce (AES-256-GCM), metadata (JSONB), version |
| `refresh_tokens` | JWT refresh tokens | token_hash, expires_at, revoked |
| `api_keys` | API keys programáticas | key_prefix, key_hash (SHA-256), revoked |

---

## Seguridad

- **Passwords**: Argon2id (memoria 64KB, 3 iteraciones, 4 hilos)
- **Secrets**: AES-256-GCM con nonce único por secreto (12 bytes aleatorios)
- **Clave maestra**: Configurable, rotable en caliente
- **JWT**: Access token 15 minutos + refresh token 7 días con rotation
- **API Keys**: Solo se muestran una vez; almacenadas como SHA-256 hash
- **Rate limiting**: Por dirección IP

---

## Estructura del proyecto

```
securevault/
├── cmd/
│   ├── server/          ← Entry point del servidor HTTP
│   └── vault/           ← Entry point del CLI
├── internal/
│   ├── cli/             ← Comandos Cobra (5 archivos)
│   ├── handler/         ← Handlers HTTP + test
│   ├── middleware/       ← JWT auth, rate limiter, logger
│   ├── model/           ← Estructuras del dominio
│   ├── repository/      ← Persistencia PostgreSQL (pgx)
│   └── service/         ← Lógica de negocio (cifrado incluido)
├── migrations/          ← SQL scheme de base de datos
├── spec/                ← Artefactos SDD
├── .github/workflows/   ← CI/CD pipeline
├── Dockerfile           ← Multi-stage build
├── docker-compose.yaml  ← Dev environment
├── openapi.yaml         ← Documentación de la API
└── config.example.yaml  ← Configuración de ejemplo
```

---

## Licencia

MIT
