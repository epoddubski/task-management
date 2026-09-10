# Task Management

REST API for managing tasks with user authentication.

## Rules

### Task Lifecycle & Assignment
- **Task Statuses**: Every task strictly operates within one of four valid states: `created`, `in_progress`, `completed`, or `overdue`.
- **Default Assignee**: When a new task is created without specifying an assignee (`assignee_id`), the system automatically binds the task to its **creator**.
- **Overdue Guard (NEW)**: If a task is created or updated with a deadline that is already in the past, the system automatically forces its status to `overdue`.
- **Deadline Monitoring**: An automatic background process continuously tracks task deadlines, shifting incomplete tasks into the `overdue` status when their timeline expires

### Authorization & Permissions
- **Visibility**: Any authenticated user has the permission to work with tasks (create, view specific tasks, or list all tasks with filters)
- **Resource Guarding**: Only the **creator** of a task has the authority to update its fields or permanently delete it
- **Access Restraints**: Any modification or deletion attempt by a non-creator is strictly blocked

## Running it

### With Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

This starts Postgres, waits for its healthcheck, and then triggers a dedicated migrations container (`task_management_migrations`) to apply any outstanding schema scripts. The main Go application container (`task_management`) automatically waits for the migrations to complete successfully before booting up.
The API is then available at `http://localhost:8080`.

### Locally

```bash
cp .env.example .env
go run ./cmd/app
```

## Configuration

All settings are loaded from environment variables (`internal/config`):

| Variable | Required | Default | Description |
|---|---|---|---|
| `DB_USER` | yes | `postgres` | Database user |
| `DB_PASSWORD` | yes | `postgres` | Database password |
| `DB_NAME` | yes | `tasks` | Database name |
| `DB_PORT` | no | `5432` | Database port |
| `APP_PORT` | no | `8080` | HTTP listen port |
| `JWT_SECRET` | yes | — | HMAC secret for signing/verifying JWTs |
| `TOKEN_TTL` | no | `15m` | JWT validity period |
| `DEADLINE_WORKER_INTERVAL` | no | `10m` | How often the overdue-task sweeper runs |

See `.env.example` for a ready-to-copy template.

## API

Authenticated routes require:
`Authorization: Bearer <token>`.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/register` | — | `{"email", "password"}` -> `201` with `{"id", "email"}` |
| POST | `/api/auth/login` | — | `{"email", "password"}` -> `200` with `{"token"}` |

### Tasks

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/tasks` | required | Create a task |
| GET | `/api/tasks` | required | List tasks, filtered + paginated |
| GET | `/api/tasks/{id}` | required | Get one task |
| PATCH | `/api/tasks/{id}` | required | Partial update — **creator only** |
| DELETE | `/api/tasks/{id}` | required | Delete — **creator only** |

## Testing

You can import the `swagger.json` file directly into **Postman** or **Swagger UI** for testing
