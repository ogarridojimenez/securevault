# Misión — SecureVault

> API + CLI profesional de gestión de secretos construida en Go. Backend real de dominio de seguridad, demostrando arquitectura hexagonal, Go avanzado y buenas prácticas de ingeniería.

## Qué construimos

SecureVault es un sistema de gestión de secretos (API REST + CLI) que permite a equipos de desarrollo almacenar, cifrar, rotar y auditar credenciales y configuraciones sensibles (passwords, API keys, tokens, certificados) de forma centralizada. Los secretos se cifran en reposo con AES-256-GCM y las contraseñas se protegen con Argon2id.

## Para quién

- **Usuario principal:** Desarrollador/a DevOps o backend que necesita gestionar secretos de aplicación desde CLI o integrarlos vía API.
- **Usuario secundario:** Pipeline CI/CD que necesita leer secretos mediante API Keys en tiempo de deploy.
- **Reclutador técnico:** Proyecto portfolio que demuestra dominio de Go, arquitectura hexagonal, REST APIs profesionales, PostgreSQL, cifrado, JWT y CLI tooling.

## Principios

- **Seguridad por diseño** — Los secretos nunca viajan ni se almacenan en texto plano. Cifrado AES-256-GCM en reposo, Argon2id para passwords, JWT con refresh rotation.
- **Arquitectura hexagonal** — Core/service/handler/repository estrictamente separados. Sin dependencias circulares. Cada capa es testeable de forma aislada.
- **API first** — Toda funcionalidad está disponible vía API REST antes que vía CLI. El CLI es un cliente de la API.
- **Código profesional** — Tests (unit + integration + httptest), OpenAPI spec, CI/CD, Docker multi-stage, logging estructurado.
- **Simplicidad sobre abstracción** — Sin patternitis. Interfaces donde aportan testeabilidad, no donde sobran.

## Qué NO es

- **No es un gestor de contraseñas personal** — Es un backend para equipos, no un reemplazo de 1Password/Bitwarden para uso individual.
- **No es HashiCorp Vault** — No tiene dynamic secrets, leasing, ni auto-rotación. Es un sustituto ligero para equipos pequeños.
- **No es un proyecto de ciberseguridad** — Es un backend de ingeniería cuyo dominio es la gestión de secretos. El foco es la calidad del código Go, no la novedad del algoritmo criptográfico.
- **No tiene UI web** — Solo API REST + CLI. La interfaz primaria es programática.
