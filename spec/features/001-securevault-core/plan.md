# 001 · SecureVault — Plan Técnico

## Arquitectura

```
┌─────────────────────────────────────────────────────┐
│                    CLI (Cobra)                       │
│  cmd/vault/main.go → internal/cli/                  │
└─────────────────────┬───────────────────────────────┘
                      │ HTTP (REST) + JWT
┌─────────────────────▼───────────────────────────────┐
│               API Server (chi v5)                    │
│  cmd/server/main.go → internal/handler/              │
│  Middleware: JWT Auth · Rate Limit · Logger · Recover│
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                   Service Layer                      │
│  internal/service/ (Auth, Vault, Secret, APIKey)    │
│  → Orquesta reglas de negocio, cifrado/descifrado    │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                 Repository Layer                     │
│  internal/repository/ (pgx v5 → PostgreSQL)          │
│  → SQL parametrizado, migraciones en /migrations     │
└─────────────────────────────────────────────────────┘
```

### Capas

| Capa | Responsabilidad | No contiene |
|------|----------------|-------------|
| `cmd/` | Entry points (servidor, CLI) | Lógica de negocio |
| `internal/handler/` | Transporte HTTP (chi), parseo, serialización | Reglas de negocio, SQL |
| `internal/service/` | Casos de uso, coordinación, cifrado/descifrado | HTTP, SQL directo |
| `internal/repository/` | Persistencia PostgreSQL (pgx), migraciones | Reglas de negocio |
| `internal/middleware/` | JWT auth, rate limiting, logging | Lógica de negocio |
| `internal/model/` | Structs de dominio compartidos | Cualquier lógica |

## Modelo de datos

### Tablas PostgreSQL

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,          -- Argon2id
    created_at   TIMESTAMPTZ DEFAULT now(),
    updated_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL,           -- SHA-256 del refresh token
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked      BOOLEAN DEFAULT false,
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE vaults (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    description  TEXT DEFAULT '',
    created_at   TIMESTAMPTZ DEFAULT now(),
    updated_at   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id, name)
);

CREATE TABLE secrets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id     UUID REFERENCES vaults(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,   -- ej: "production/db/pwd"
    ciphertext   BYTEA NOT NULL,          -- AES-256-GCM cifrado
    nonce        BYTEA NOT NULL,          -- 12 bytes nonce
    metadata     JSONB DEFAULT '{}',      -- Etiquetas, versión, etc.
    version      INTEGER DEFAULT 1,
    created_at   TIMESTAMPTZ DEFAULT now(),
    updated_at   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(vault_id, name)
);

CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    key_prefix   VARCHAR(10) NOT NULL,    -- ej: "sv_" + primeros 8 chars del hash
    key_hash     TEXT NOT NULL,           -- SHA-256 del valor completo
    last_used_at TIMESTAMPTZ,
    revoked      BOOLEAN DEFAULT false,
    created_at   TIMESTAMPTZ DEFAULT now()
);

-- Índices
CREATE INDEX idx_vaults_user ON vaults(user_id);
CREATE INDEX idx_secrets_vault ON secrets(vault_id);
CREATE INDEX idx_secrets_name ON secrets(vault_id, name);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
```

## Contratos API

| Método | Ruta | Auth | Body | Respuesta |
|--------|------|------|------|-----------|
| POST | /auth/register | No | `{email, password}` | `{user}` 201 |
| POST | /auth/login | No | `{email, password}` | `{access_token, refresh_token}` 200 |
| POST | /auth/refresh | No | `{refresh_token}` | `{access_token, refresh_token}` 200 |
| POST | /auth/logout | No | `{refresh_token}` | 204 |
| GET | /health | No | — | `{status}` 200 |
| GET | /health/readiness | No | — | `{status, db}` 200/503 |
| POST | /vaults | JWT | `{name, description}` | `{vault}` 201 |
| GET | /vaults | JWT | — | `{vaults, total}` 200 |
| GET | /vaults/{id} | JWT | — | `{vault}` 200 |
| PUT | /vaults/{id} | JWT | `{name, description}` | `{vault}` 200 |
| DELETE | /vaults/{id} | JWT | — | 204 |
| POST | /vaults/{id}/secrets | JWT/APIKey | `{name, value, metadata?}` | `{secret}` 201 |
| GET | /vaults/{id}/secrets | JWT/APIKey | — | `{secrets}` 200 |
| GET | /vaults/{id}/secrets/{sid} | JWT/APIKey | — | `{secret, value}` 200 |
| PUT | /vaults/{id}/secrets/{sid} | JWT/APIKey | `{value, metadata?}` | `{secret}` 200 |
| DELETE | /vaults/{id}/secrets/{sid} | JWT | — | 204 |
| POST | /api-keys | JWT | `{name}` | `{api_key_full, api_key}` 201 |
| GET | /api-keys | JWT | — | `{api_keys}` 200 |
| DELETE | /api-keys/{id} | JWT | — | 204 |

## Decisiones técnicas

| Decisión | Opción elegida | Alternativa descartada |
|----------|---------------|----------------------|
| **Cifrado de secretos** | AES-256-GCM con clave derivada del master key vía HKDF | Vault transit engine (demasiado peso) |
| **Hash de passwords** | Argon2id (stdlib `crypto/argon2`) | bcrypt (menos resistente a GPU) |
| **Persistencia** | pgx v5 directo, SQL plano | GORM (pérdida de control SQL), sqlx |
| **Refresh tokens** | SHA-256 hash en DB, rotación en cada refresh | JWT refresh sin store (no revocable) |
| **API Keys** | Prefijo legible + hash SHA-256 del secreto en DB | Almacenar en texto plano (inseguro) |
| **Router** | chi v5 | gorilla/mux (mantenimiento incierto), stdlib (sin middleware nativo) |
| **Migraciones** | Archivos .sql planos con contador numérico | goose/golang-migrate (dependencia extra) |

## Riesgos y mitigaciones

| Riesgo | Mitigación |
|--------|-----------|
| Pérdida de master key → imposible descifrar | Documentar backup procedure, separar key de la DB |
| Nonce reutilizado → ruptura AES-GCM | `crypto/rand` siempre, assert if nonce repetido |
| Token JWT robado | Refresh tokens rotativos + corta duración access token (15min) |
| SQL injection | Consultas parametrizadas siempre (`$1`, `$2`), nunca concatenación |
| Rate limit saturation | Sliding window en memoria por IP y por user ID |
