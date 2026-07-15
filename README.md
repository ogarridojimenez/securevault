# SecureVault

> API REST + CLI profesional de gestión de secretos en Go.
> Arquitectura hexagonal · Go 1.23+ · Chi · PostgreSQL · AES-256-GCM · JWT · Argon2id

## Stack

| Componente | Tecnología |
|-----------|------------|
| Lenguaje | Go 1.23+ |
| API Router | chi v5 |
| CLI | Cobra + Viper |
| Base de datos | PostgreSQL 16 |
| Driver DB | pgx v5 |
| JWT | golang-jwt v5 (HMAC) |
| Cifrado | AES-256-GCM |
| Passwords | Argon2id |
| API Keys | SHA-256 hash |
| CI/CD | GitHub Actions |
| Contenedores | Docker multi-stage |

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
           └──────────────┬──────────────────────┘
                          │
           ┌──────────────▼──────────────────────┐
           │      Service Layer                  │
           │  Auth · Vaults · Secrets · API Keys │
           └──────────────┬──────────────────────┘
                          │
           ┌──────────────▼──────────────────────┐
           │      Repository Layer               │
           │  PostgreSQL (pgx)                   │
           └──────────────┬──────────────────────┘
                          │
              ┌───────────▼───────────┐
              │   PostgreSQL 16       │
              │   (datos cifrados)    │
              └───────────────────────┘
```

## Quick Start

### 1. Requisitos

- Go 1.23+
- PostgreSQL 16
- Docker (opcional)

### 2. Configuración

```bash
# Generar clave de cifrado (32 bytes hex)
python3 -c "import secrets; print(secrets.token_hex(32))"

# Copiar config
cp config.example.yaml config.yaml
# Editar config.yaml con:
#   - db.url: conexión PostgreSQL
#   - jwt.secret: secreto JWT
#   - encryption.key: clave AES-256-GCM (64 hex chars)
```

### 3. Base de datos

```bash
createdb securevault
psql -d securevault -f migrations/001_initial_schema.sql
```

### 4. Ejecutar

```bash
# Servidor
go run ./cmd/server/

# CLI
go run ./cmd/vault/ --help
```

### O con Docker

```bash
docker compose up -d
```

## API Endpoints (16)

### Auth
| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | /api/v1/auth/register | Registro |
| POST | /api/v1/auth/login | Login |
| POST | /api/v1/auth/refresh | Refresh token |
| POST | /api/v1/auth/logout | Logout |

### Vaults
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /api/v1/vaults | Listar |
| POST | /api/v1/vaults | Crear |
| GET | /api/v1/vaults/{id} | Obtener |
| PUT | /api/v1/vaults/{id} | Actualizar |
| DELETE | /api/v1/vaults/{id} | Eliminar |

### Secrets (cifrados AES-256-GCM)
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /api/v1/vaults/{id}/secrets | Listar |
| POST | /api/v1/vaults/{id}/secrets | Crear |
| GET | /api/v1/vaults/{id}/secrets/{sid} | Obtener valor |
| DELETE | /api/v1/vaults/{id}/secrets/{sid} | Eliminar |

### API Keys
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /api/v1/api-keys | Listar |
| POST | /api/v1/api-keys | Crear |
| PUT | /api/v1/api-keys/{id}/revoke | Revocar |

### Health
| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /api/v1/health | Health check |
| GET | /api/v1/readiness | Readiness check |

## CLI (8 comandos)

```bash
vault login                          # Autenticarse
vault set production/db/pwd "X"      # Guardar secreto
vault get production/db/pwd          # Leer secreto
vault list production/               # Listar secrets
vault delete production/db/pwd       # Eliminar secreto
vault export > backup.json           # Exportar todo
vault import backup.json             # Importar
vault api-key create "CI/CD"         # Crear API key
vault api-key list                   # Listar API keys
vault api-key revoke <id>            # Revocar API key
```

## Modelo de datos

- **users**: id, email, password_hash (Argon2id), timestamps
- **vaults**: id, user_id, name, description, timestamps (UNIQUE user_id+name)
- **secrets**: id, vault_id, name, ciphertext (AES-256-GCM), nonce, metadata (JSONB), version, timestamps
- **refresh_tokens**: id, user_id, token_hash, expires_at, revoked, timestamps
- **api_keys**: id, user_id, name, key_prefix, key_hash, last_used_at, revoked, timestamps

## Seguridad

- Passwords hasheadas con **Argon2id** (m=64KB, t=3, p=4)
- Secrets cifrados con **AES-256-GCM** (nonce único por secreto)
- Clave maestra de cifrado configurable, rotable
- JWT access token (15min) + refresh token (7d) con rotation
- API Keys hasheadas con SHA-256 (solo se muestra una vez)
- Rate limiting por IP

## Licencia

MIT
