# 001 — SecureVault · API + CLI de gestión de secretos

**Estado:** ✅ especificado

## Objetivo

Sistema backend de gestión de secretos (API REST + CLI) que permite a equipos de desarrollo almacenar, cifrar, recuperar y auditar credenciales y configuraciones sensibles. Los datos se cifran en reposo con AES-256-GCM, las contraseñas se protegen con Argon2id, y el acceso se autentica mediante JWT (access+refresh) o API Keys.

## Actores

| Actor | Descripción |
|-------|-------------|
| **Usuario registrado** | Accede vía CLI o API con JWT. Gestiona vaults y secretos. |
| **Pipeline CI/CD** | Accede vía API Key (solo lectura). Lee secretos para despliegues. |
| **Sistema** | Cifra/descifra, valida tokens, audita operaciones. |

## Historias de usuario

- Como **usuario registrado**, quiero autenticarme con email+password para obtener un token JWT
- Como **usuario registrado**, quiero crear vaults para organizar mis secretos por proyecto/entorno
- Como **usuario registrado**, quiero guardar/leer/actualizar/eliminar secretos cifrados dentro de un vault
- Como **pipeline CI/CD**, quiero usar una API Key para leer secretos sin intervención humana
- Como **usuario registrado**, quiero usar la CLI para gestionar secretos sin escribir código

## Requisitos funcionales (EARS)

### Auth

| ID | Requisito |
|----|-----------|
| AUTH-01 | CUANDO un usuario envía `POST /auth/register` con email+password válidos, EL SISTEMA DEBE crear el usuario con Argon2id y devolver 201 |
| AUTH-02 | CUANDO un usuario envía `POST /auth/login` con credenciales válidas, EL SISTEMA DEBE devolver un access token (JWT, 15min) + refresh token (JWT, 7d) |
| AUTH-03 | CUANDO un usuario envía `POST /auth/refresh` con un refresh token válido, EL SISTEMA DEBE devolver un nuevo par access+refresh |
| AUTH-04 | CUANDO un usuario envía `POST /auth/logout` con un refresh token, EL SISTEMA DEBE invalidar ese refresh token |
| AUTH-05 | SI un token JWT ha expirado o es inválido, EL SISTEMA DEBE rechazar la petición con 401 |
| AUTH-06 | SI un email ya está registrado, EL SISTEMA DEBE rechazar el registro con 409 |

### Vaults

| ID | Requisito |
|----|-----------|
| VAULT-01 | CUANDO un usuario autenticado envía `POST /vaults`, EL SISTEMA DEBE crear un vault con nombre y descripción, devolviendo 201 |
| VAULT-02 | CUANDO un usuario autenticado envía `GET /vaults`, EL SISTEMA DEBE listar todos sus vaults (paginados) |
| VAULT-03 | CUANDO un usuario autenticado envía `GET /vaults/{id}`, EL SISTEMA DEBE devolver el vault si pertenece al usuario |
| VAULT-04 | CUANDO un usuario autenticado envía `PUT /vaults/{id}`, EL SISTEMA DEBE actualizar nombre/descripción del vault |
| VAULT-05 | CUANDO un usuario autenticado envía `DELETE /vaults/{id}`, EL SISTEMA DEBE eliminar el vault y todos sus secretos |
| VAULT-06 | SI un usuario solicita un vault que no existe o no le pertenece, EL SISTEMA DEBE devolver 404 |

### Secrets

| ID | Requisito |
|----|-----------|
| SEC-01 | CUANDO un usuario autenticado envía `POST /vaults/{id}/secrets`, EL SISTEMA DEBE cifrar el valor con AES-256-GCM y almacenarlo |
| SEC-02 | CUANDO un usuario autenticado envía `GET /vaults/{id}/secrets`, EL SISTEMA DEBE listar los metadatos de los secretos (NUNCA los valores cifrados) |
| SEC-03 | CUANDO un usuario autenticado envía `GET /vaults/{id}/secrets/{secretId}`, EL SISTEMA DEBE descifrar y devolver el valor del secreto |
| SEC-04 | CUANDO un usuario autenticado envía `PUT /vaults/{id}/secrets/{secretId}`, EL SISTEMA DEBE actualizar el valor (recién cifrado) |
| SEC-05 | CUANDO un usuario autenticado envía `DELETE /vaults/{id}/secrets/{secretId}`, EL SISTEMA DEBE eliminar el secreto |
| SEC-06 | CUANDO el sistema cifra un secreto, DEBE generar un nuevo nonce de 12 bytes aleatorios para cada operación |

### API Keys

| ID | Requisito |
|----|-----------|
| KEY-01 | CUANDO un usuario autenticado envía `POST /api-keys`, EL SISTEMA DEBE generar una API Key con prefijo identificable y devolverla una sola vez |
| KEY-02 | CUANDO un usuario autenticado envía `GET /api-keys`, EL SISTEMA DEBE listar sus API Keys (sin el valor completo) |
| KEY-03 | CUANDO un usuario autenticado envía `DELETE /api-keys/{id}`, EL SISTEMA DEBE revocar la API Key |
| KEY-04 | CUANDO una petición llega con una API Key válida en `X-API-Key`, EL SISTEMA DEBE autenticarla y permitir acceso de solo lectura a secretos |

### Health

| ID | Requisito |
|----|-----------|
| HLTH-01 | CUANDO un cliente envía `GET /health`, EL SISTEMA DEBE devolver 200 con estado del servidor |
| HLTH-02 | CUANDO un cliente envía `GET /health/readiness`, EL SISTEMA DEBE verificar conexión a PostgreSQL y devolver 200/503 |

## CLI (Cobra)

| ID | Requisito |
|----|-----------|
| CLI-01 | CUANDO el usuario ejecuta `vault login`, EL SISTEMA DEBE autenticar y almacenar el token localmente |
| CLI-02 | CUANDO el usuario ejecuta `vault set <path> <value>`, EL SISTEMA DEBE crear/actualizar un secreto en la ruta especificada |
| CLI-03 | CUANDO el usuario ejecuta `vault get <path>`, EL SISTEMA DEBE mostrar el valor del secreto |
| CLI-04 | CUANDO el usuario ejecuta `vault list <prefix>`, EL SISTEMA DEBE listar secretos bajo ese prefijo |
| CLI-05 | CUANDO el usuario ejecuta `vault delete <path>`, EL SISTEMA DEBE eliminar el secreto |
| CLI-06 | CUANDO el usuario ejecuta `vault export`, EL SISTEMA DEBE exportar todos los secretos (cifrados) a JSON |
| CLI-07 | CUANDO el usuario ejecuta `vault import <file>`, EL SISTEMA DEBE importar secretos desde un JSON |
| CLI-08 | CUANDO el usuario ejecuta `vault api-key create <name>`, EL SISTEMA DEBE generar y mostrar una API Key |

## Requisitos no funcionales

| ID | Requisito |
|----|-----------|
| NFR-01 | EL SISTEMA DEBE responder en <200ms (p95) para operaciones de lectura de secretos |
| NFR-02 | EL SISTEMA DEBE usar AES-256-GCM con nonce único de 12 bytes por cifrado |
| NFR-03 | EL SISTEMA DEBE hashear contraseñas con Argon2id (salt 16 bytes, time=3, memory=64MB, threads=4) |
| NFR-04 | EL SISTEMA DEBE emitir logs estructurados (JSON) via `log/slog` |
| NFR-05 | EL SISTEMA DEBE soportar rate limiting por IP (100 req/min) y por usuario (300 req/min) |

## Non-goals (fuera de alcance)

- No hay UI web ni frontend
- No hay dynamic secrets (generación automática de credenciales)
- No hay integración con proveedores cloud (AWS Secrets Manager, etc.)
- No hay HA/clustering — es mononodo con PostgreSQL

## Criterios de aceptación

- [ ] `go build ./...` pasa sin errores
- [ ] `go test ./...` pasa con cobertura >60%
- [ ] Registro + login + refresh + logout funcionan end-to-end
- [ ] Creación de vault, creación de secreto, lectura del valor descifrado
- [ ] API Key autentica correctamente en endpoints de solo lectura
- [ ] CLI login + set + get + list + delete funcionan contra servidor real
- [ ] Docker build multi-stage produce imagen funcional
- [ ] OpenAPI spec describe correctamente los 16 endpoints
