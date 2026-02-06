FROM debian:bookworm-slim

# --- basics ---
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    bash \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# --- copy runtime artifacts ---
# These are produced by Makefile, NOT built here
COPY bin/server /app/server
COPY web /app/web
COPY migrations /app/migrations

# --- runtime ---
EXPOSE 8080

ENTRYPOINT ["/app/server"]
