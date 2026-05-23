# MondaiPhi (問題)

**Question Bank Service for JLPT Exam Microservices**

MondaiPhi is the question bank microservice for the Philia-Space ecosystem. It stores, manages, and serves JLPT (Japanese Language Proficiency Test) practice questions, passages, audio/image assets, and package templates.

## Overview

| Attribute | Value |
|-----------|-------|
| **Name** | MondaiPhi (問題 = "question/problem") |
| **Port** | 8087 |
| **Database** | PostgreSQL |
| **Storage** | BiznetGio S3 (Neo Cloud) |
| **Stack** | Go + standard lib HTTP + `phi-core` DDD |

## Features

- **Question Management**: CRUD for JLPT N1–N5 questions with options
- **Passage Support**: Reading/listening passages that group related questions
- **Asset Management**: Audio/image files stored in S3 with presigned URL access
- **Package Templates**: Configurable exam blueprints (e.g., `balanced_75`: 75 questions across sections)
- **Atomic Units**: Questions sharing a passage or source group are always returned together
- **Admin API**: Secure CRUD endpoints for content management
- **Answer Security**: Correct answers never leak to public endpoints

## Architecture

```
MondaiPhi/
├── main.go                          # Entry point
├── config/                          # Environment configuration
├── handlers/                        # HTTP handlers
│   ├── question.go                  # Public read-only routes
│   └── admin.go                     # Admin CRUD routes
├── internal/                        # Domain + Application layers
│   ├── domain/                      # Aggregates, entities, repositories
│   │   ├── question.go              # Question aggregate root
│   │   ├── passage.go               # Passage entity
│   │   ├── asset.go                 # Asset entity
│   │   ├── template.go              # PackageTemplate entity
│   │   └── repository.go            # Repository interfaces
│   └── application/                 # Command handlers (use cases)
│       └── create_question.go
├── repositories/
│   ├── postgres/                    # PostgreSQL implementation
│   │   ├── client.go                # DB connection
│   │   └── repository.go            # Full repository with all queries
│   └── memory/                      # In-memory fake for tests
└── migrations/
    └── 001_schema.sql               # PostgreSQL DDL
```

## API Endpoints

### Public (Read-Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/questions?level=N3&section=grammar&limit=15` | List questions (sanitized) |
| `GET` | `/questions/:id` | Single question with options (no answer) |
| `GET` | `/passages/:id` | Passage with nested questions |
| `GET` | `/templates` | List package templates |
| `GET` | `/assets/:id` | Redirect to presigned S3 URL |

### Admin (Requires `admin` or `superadmin` role)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/questions` | Create question + options |
| `PUT` | `/admin/questions/:id` | Update question |
| `DELETE` | `/admin/questions/:id` | Soft delete |
| `POST` | `/admin/passages` | Create passage |
| `PUT` | `/admin/passages/:id` | Update passage |
| `POST` | `/admin/assets` | Upload file → S3 |
| `DELETE` | `/admin/assets/:id` | Delete from S3 + metadata |
| `POST` | `/admin/templates` | Create template |

## Data Model

### Sections (3 unified sections)
- `grammar` — Grammar, vocabulary, kanji (merged)
- `reading` — Reading comprehension
- `listening` — Listening comprehension

### ID Prefixes
- `qst_` — Question
- `psg_` — Passage
- `ast_` — Asset
- `tpl_` — Template

## Configuration

```env
MONDAIPHI_PORT=8087
MONDAIPHI_ENVIRONMENT=development
MONDAIPHI_DATABASE_URL=postgres://user:pass@localhost:5433/mondaiphi?sslmode=disable
MONDAIPHI_AUTH_JWKS_URL=http://localhost:8080/.well-known/jwks.json
MONDAIPHI_STORAGE_ENDPOINT=https://s3.biznetgio.net
MONDAIPHI_STORAGE_REGION=ap-southeast-1
MONDAIPHI_STORAGE_BUCKET=philia-space
MONDAIPHI_STORAGE_ACCESS_KEY=...
MONDAIPHI_STORAGE_SECRET_KEY=...
MONDAIPHI_STORAGE_PRESIGN_TTL=3600
MONDAIPHI_ADMIN_USER_IDS=discord_id_1,discord_id_2
```

## Getting Started

```bash
# Install dependencies
cd services/MondaiPhi
go mod download

# Run migrations
psql -U phi -d mondaiphi -f migrations/001_schema.sql

# Start server
make run
# or
go run main.go
```

## Related Services

- **ShikenPhi** (試験) — Exam session service (consumes this API)
- **AuthPhi** — Authentication & JWT provider
- **LyraPhi** — Frontend exam application

## License

ISC
