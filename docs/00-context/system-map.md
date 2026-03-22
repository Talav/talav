# System Map

Navigation index. One-liner per module + links to READMEs. For architecture rationale see [vision.md](vision.md); for open decisions see [assumptions.md](assumptions.md).

---

## Layer diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ Application Layer  (main.go, pkg/module/*)                      │
│   Bootstrap application, compose FX modules, define commands    │
└─────────────────────────────────────────────────────────────────┘
                            ↓ imports
┌─────────────────────────────────────────────────────────────────┐
│ FX Layer  (pkg/fx/*)                                            │
│   Wire components via Uber FX, register commands, load config   │
└─────────────────────────────────────────────────────────────────┘
                            ↓ imports
┌─────────────────────────────────────────────────────────────────┐
│ Component Layer  (pkg/component/*)                              │
│   Pure business logic — zero framework or FX imports            │
└─────────────────────────────────────────────────────────────────┘
```

Import direction is enforced by `golangci-lint` (`depguard`). Violations fail the pre-commit hook.

---

## Components (`pkg/component/*`)

| Package | Purpose | README |
|---|---|---|
| `blob` | Cloud storage abstraction (gocloud.dev/blob) — local, S3, GCS, Azure | [README](../../pkg/component/blob/README.md) |
| `clock` | Testable time source — SystemClock, FixedClock, OffsetClock | [README](../../pkg/component/clock/README.md) |
| `config` | YAML + env config loading (koanf); `config:` struct tags; duration decoding | [README](../../pkg/component/config/README.md) |
| `email` | Email delivery abstraction | [README](../../pkg/component/email/README.md) |
| `httpserver` | Chi HTTP server with structured logging and Zorya integration | [README](../../pkg/component/httpserver/README.md) |
| `logger` | slog-based structured logging | [README](../../pkg/component/logger/README.md) |
| `mapstructure` | mapstructure decode hook utilities | [README](../../pkg/component/mapstructure/README.md) |
| `media` | Media upload, CDN routing, image presets *(prototype)* | [README](../../pkg/component/media/README.md) |
| `metrics` | Isolated Prometheus registry factory | [README](../../pkg/component/metrics/README.md) |
| `negotiation` | HTTP content negotiation | [README](../../pkg/component/negotiation/README.md) |
| `orm` | GORM + golang-migrate for PostgreSQL | [README](../../pkg/component/orm/README.md) |
| `schema` | HTTP request parameter decoding; struct tags `schema`, `body`, `header` | [README](../../pkg/component/schema/README.md) |
| `security` | Auth enforcement — roles, permissions, Casbin adapter | [README](../../pkg/component/security/README.md) |
| `seeder` | Database seeding for development and test fixtures | [README](../../pkg/component/seeder/README.md) |
| `tagparser` | Struct tag parsing utilities | [README](../../pkg/component/tagparser/README.md) |
| `user` | User + role management, CQRS handlers, password reset | [README](../../pkg/component/user/README.md) |
| `validator` | go-playground/validator wrapper with custom validators and i18n | [README](../../pkg/component/validator/README.md) |
| `zorya` | Type-safe HTTP API framework — generics, OpenAPI generation, RFC 9457 errors | [README](../../pkg/component/zorya/README.md) |

---

## FX modules (`pkg/fx/*`)

| Package | Wires | README |
|---|---|---|
| `fxblob` | `*blob.FilesystemRegistry` from config; `AsFilesystem` helper | [README](../../pkg/fx/fxblob/README.md) |
| `fxclock` | `clock.Clock` (system / fixed / offset) from config | [README](../../pkg/fx/fxclock/README.md) |
| `fxconfig` | `*config.Config`; `AsConfigSource`, `AsConfig[T]`, `AsConfigWithDefaults` | [README](../../pkg/fx/fxconfig/README.md) |
| `fxcore` | Command group wiring; `AsRootCommand`, `AsNamedCommand` | [README](../../pkg/fx/fxcore/README.md) |
| `fxemail` | Email sender from config | [README](../../pkg/fx/fxemail/README.md) |
| `fxhttpserver` | Chi server from config; `AsMiddleware`, `AsRoute` | [README](../../pkg/fx/fxhttpserver/README.md) |
| `fxlogger` | `*slog.Logger` from config | [README](../../pkg/fx/fxlogger/README.md) |
| `fxmedia` | Media command/query handlers, CDN + provider + resizer wiring | [README](../../pkg/fx/fxmedia/README.md) |
| `fxmetrics` | `*prometheus.Registry`; `AsMetricsCollector` | [README](../../pkg/fx/fxmetrics/README.md) |
| `fxorm` | `*gorm.DB`, migration runner, migrate CLI commands; `AsRepository[T]` | [README](../../pkg/fx/fxorm/README.md) |
| `fxsecurity` | Security enforcer + Casbin adapter from config | [README](../../pkg/fx/fxsecurity/README.md) |
| `fxseeder` | Seeder runner + `seed` CLI command | [README](../../pkg/fx/fxseeder/README.md) |
| `fxuser` | User + role repos, CQRS handlers, password validator | [README](../../pkg/fx/fxuser/README.md) |
| `fxvalidator` | `*validator.Validate`; `AsValidatorConstructor`, `AsTranslationConstructor` | [README](../../pkg/fx/fxvalidator/README.md) |

---