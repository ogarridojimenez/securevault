# 001 · SecureVault — Tareas

| # | Tarea | Depende de | Criterio de aceptación |
|---|-------|-----------|------------------------|
| 1 | Inicializar módulo Go y estructura de directorios | — | `go mod init`, `go build ./...` pasa |
| 2 | Crear migraciones SQL (users, vaults, secrets, api_keys, refresh_tokens) | 1 | SQL ejecutable contra PostgreSQL, índices creados |
| 3 | Implementar `internal/repository/` (pgx pool, queries parametrizadas) | 2 | Cada tabla tiene repo con CRUD básico + tests |
| 4 | Implementar `internal/model/` — structs compartidos | 1 | Todos los tipos definidos y exportables |
| 5 | Implementar `internal/service/auth.go` — register, login, refresh, logout | 3, 4 | Argon2id, JWT pair, refresh rotation |
| 6 | Implementar `internal/middleware/jwt.go` — JWT auth middleware | 5 | Extrae user_id del token, rechaza 401 si inválido |
| 7 | Implementar `internal/handler/auth.go` — endpoints /auth/* | 5, 6 | 4 endpoints funcionales con chi |
| 8 | Implementar `internal/service/vault.go` — CRUD vaults | 3, 4 | Operaciones con ownership check |
| 9 | Implementar `internal/handler/vault.go` — endpoints /vaults/* | 8 | 5 endpoints funcionales |
| 10 | Implementar `internal/service/secret.go` — CRUD secrets + AES-256-GCM | 3, 4 | Valores cifrados en DB, descifrados al leer |
| 11 | Implementar `internal/handler/secret.go` — endpoints /vaults/{id}/secrets/* | 10 | 5 endpoints funcionales |
| 12 | Implementar `internal/service/apikey.go` — generate, list, revoke | 3, 4 | Prefijo + hash, devolver full key una vez |
| 13 | Implementar `internal/handler/apikey.go` — endpoints /api-keys/* | 12 | 3 endpoints funcionales |
| 14 | Implementar `internal/handler/health.go` — health + readiness | 1 | Ping DB, devolver estado |
| 15 | Implementar `cmd/server/main.go` — servidor HTTP completo | 7, 9, 11, 13, 14 | chi con todos los routers montados, graceful shutdown |
| 16 | Implementar `internal/middleware/ratelimit.go` — rate limiting | 6 | Sliding window IP + user, 429 si excede |
| 17 | Implementar CLI Cobra (cmd/vault/) — login, set, get, list, delete, export, import, api-key | 15 | 8 comandos funcionales contra API |
| 18 | Crear OpenAPI 3.1 spec (openapi.yaml) | 15 | 16 endpoints documentados |
| 19 | Crear Dockerfile multi-stage + docker-compose.yml | 15 | Build reproduce imagen <20MB |
| 20 | Crear GitHub Actions CI (lint → test → build → release) | 19 | Pipeline pasa en push |
| 21 | Tests: unitarios + integración + httptest | 17 | `go test ./...` pasa, cobertura >60% |
| 22 | README.md + AGENTS.md + documentación final | 21 | Documentación completa del proyecto |

### Dependencias visuales

```
1 → 2 → 3 → 4 ──────────────────────────────────────────┐
                 ↓                                       │
                 5 → 6 → 7 (auth)                       │
                 8 → 9 (vaults)                         │
                 10 → 11 (secrets)                      │
                 12 → 13 (api-keys)                     │
                 14 (health)                            │
                 ↓                                      │
                 15 (server main) ←─────────────────────┘
                 ↓
                 16 (rate limit)
                 17 (CLI)
                 18 (OpenAPI)
                 19 (Docker)
                 20 (CI/CD)
                 21 (Tests)
                 22 (Docs)
```

### Paralelizables

- Tareas 5, 8, 10, 12, 14 pueden implementarse en paralelo (dependen de 3 y 4)
- Tareas 7, 9, 11, 13 (handlers) pueden implementarse en paralelo después de sus servicios respectivos
- Tareas 18, 19 pueden ir en paralelo con 20, 21
