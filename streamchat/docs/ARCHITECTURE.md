# Architecture

The project is a modular monolith. HTTP handlers call domain services and repositories; PostgreSQL is the durable source of truth and Redis is reserved for queue/pub-sub work. The API and media worker are separate processes with independent graceful shutdown.

Phase 1 establishes the API/config/auth boundary. Later phases add conversations/messages, the WebSocket hub, storage, and media jobs without moving business logic into Gin handlers.
