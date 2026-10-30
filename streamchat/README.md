# StreamChat

StreamChat is a modular-monolith real-time chat application built in Go/Gin and React. Development follows the phased requirements in `../ANGENTS.md`.

## Implementation status

Implemented:

- Go/Gin API, PostgreSQL migrations, authentication, conversations, messages, and WebSocket updates.
- React/Vite chat UI with virtualized message history and scroll anchoring.
- Local attachment storage and access authorization.
- Media job persistence, Redis queue, SSRF checks, and isolated yt-dlp worker.
- Unit tests for backend packages and frontend components.

Media pipeline limitations:

- Worker storage currently uses the shared local volume, not Cloudflare R2.
- The egress proxy is mandatory for media downloads. Configure `MEDIA_EGRESS_PROXY` to a controlled proxy that enforces hostname and IP egress restrictions; API DNS validation alone does not constrain extractor follow-up requests.
- Media download limits are configured with `MEDIA_MAX_SIZE_MB`, `MEDIA_MAX_DURATION_SECONDS`, `MEDIA_MAX_HEIGHT`, `MEDIA_DOWNLOAD_TIMEOUT_SECONDS`, and `MEDIA_WORKER_CONCURRENCY`.
- A standalone HTTPS video URL from a supported provider creates a video message and queued media job. The chat polls job status and renders a protected inline video player after download completes.
- Video output remains in the shared local volume, not Cloudflare R2. FFmpeg normalization, thumbnails, and push-based WebSocket media status events are not implemented.
- The media content endpoint supports byte ranges for the current local storage backend. Other storage adapters must provide a seekable reader for range requests.

## Run locally

```sh
cp .env.example .env
cd backend && go mod tidy && go test ./... && cd ..
npm --prefix frontend install
make build
make dev
```

Open `http://localhost:5173`. Compose builds the Goose migration CLI with the backend image and runs versioned SQL migrations before starting API and worker. To apply new migrations to an existing database, run `docker compose run --build --rm migrate`. Do not remove the PostgreSQL volume unless you intentionally want to delete all local database data.

## Run with Podman Compose

Podman uses an external Compose provider. On Windows, start the Podman machine first and run these commands from the project root (the directory containing `docker-compose.yml`):

```powershell
podman machine start
podman compose version
podman compose up --build
```

The foreground command starts PostgreSQL, Redis, runs migrations, and then starts the API, worker, and frontend. Open `http://localhost:5173`. For detached operation, use `podman compose up -d --build`; stop the services with `podman compose down`. To apply migrations separately, run `podman compose run --build --rm migrate`.

Where Make is available, invoke targets with `make COMPOSE='podman compose' dev`; replace `dev` with `up`, `migrate`, or `down` as needed. On Windows without Make, use the direct Podman commands above. Configure `MEDIA_EGRESS_PROXY` to a controlled proxy before using media downloads. Do not remove the PostgreSQL volume unless you intentionally want to delete all local database data.

## Local media worker

The worker is read-only except for its dedicated temporary and media volumes, runs as UID 10001, and has Compose CPU, memory, and process limits. Local downloads are disabled unless `MEDIA_EGRESS_PROXY` points to a controlled egress proxy. Do not expose a generic unrestricted proxy to the worker.

The media worker stores outputs on the shared local volume. Video URL messages are playable through the authenticated media content endpoint after the worker completes. Downloads require `MEDIA_EGRESS_PROXY` to point to a controlled egress proxy; configure the same shared `LOCAL_STORAGE_ROOT` for the API and worker.

## API

- `GET /health/live`
- `GET /health/ready`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

The development cookie is not Secure so it works over localhost HTTP. Set `COOKIE_SECURE=true` behind HTTPS in production and replace `JWT_SECRET` with a random secret.
