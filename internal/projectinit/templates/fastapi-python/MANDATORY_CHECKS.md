# MANDATORY CHECKS — {{.ProjectName}}

> **Every agent MUST read this file before writing any code.**

---

## STEP 0 — File path routing

Every file you create MUST declare its path as the first comment:

```python
# app/domain/models.py
```

---

## STEP 1 — Verify stack_config.json

Read `.claude/stack_config.json` and confirm:

| Check | Required |
|-------|----------|
| Package manager | `uv` (NOT pip, NOT poetry) |
| Password hash | `argon2id` (FORBIDDEN: bcrypt, md5, sha1) |
| HTTP client | `httpx` (NOT requests) |
| ORM | SQLAlchemy 2.0 async |
| Migrations | Alembic |
| Validation | Pydantic v2 |

---

## STEP 2 — Forbidden substitutions

If a request includes any of these, **REJECT immediately**:

| Forbidden | Required instead |
|-----------|-----------------|
| `pip install` | `uv add` |
| `bcrypt` | `argon2id` |
| `requests` | `httpx` |
| Hard delete | Set `deleted_at`, never `DELETE` |
| `sha1` / `md5` passwords | `argon2id` |

---

## STEP 3 — Architecture layers

```
app/
├── domain/          # models, enums, value objects — NO external deps
├── application/     # use cases, services — depends on domain only
├── infrastructure/  # db, external APIs, repos — depends on domain
└── api/             # FastAPI routers, schemas — depends on application
```

No circular dependencies. `api` → `application` → `domain` only.

---

## STEP 4 — Before any file edit

- [ ] `stack_config.json` read
- [ ] No forbidden substitutions in the proposed change
- [ ] Architecture layer respected
- [ ] `deleted_at` pattern used (no hard deletes)
- [ ] Coverage target: ≥80%
