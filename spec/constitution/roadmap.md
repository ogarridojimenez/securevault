# Roadmap — SecureVault

> Estado y orden de las features. Cada feature tiene su carpeta en `spec/features/NNN-nombre/`.

## Hecho ✅

*(Proyecto en inicialización)*

## Siguiente 🔜

1. **001 · Proyecto base** — Estructura Go, módulos, configuración, `main.go` de servidor y CLI skeletons
2. **002 · Auth** — Register, login, JWT (access+refresh), Argon2id, middleware
3. **003 · Vaults CRUD** — Gestión de vaults (contenedores lógicos de secretos)
4. **004 · Secrets CRUD** — Secretos cifrados con AES-256-GCM dentro de vaults
5. **005 · CLI Cobra** — 8 comandos CLI completos (login, set, get, list, delete, export, import, api-key)
6. **006 · API Keys** — API keys para CI/CD, create/list/revoke
7. **007 · Rate Limiting** — Protección contra abuso por IP/usuario
8. **008 · CI/CD** — GitHub Actions + Docker multi-stage + OpenAPI spec
9. **009 · Tests y hardening** — Suite completa de tests, benchmarks, documentación

## Backlog 💡

- **RBAC** — Roles (admin/editor/viewer) con permisos granulares sobre vaults
- **Audit log** — Registro de todas las operaciones sobre secretos (quién, cuándo, qué)
- **Secret rotation** — Auto-rotación periódica de secretos con notificaciones
- **Web UI** — Panel de administración básico embebido
- **Métricas Prometheus** — `/metrics` endpoint con contadores de operaciones
