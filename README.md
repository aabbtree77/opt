# Introduction

**initialsdb** is a public bulletin board (message store) implemented as a Go backend serving a React SPA, together with PostgreSQL and Docker Compose infrastructure. It uses proof of work and rate limiting to fight bots.

This repository includes application code and infrastructure so that one can host more such web apps on the same VPS, not just initialsdb. Makefile, Dockerfile, docker-compose.yml files and this readme.md show how to replicate the whole system end-to-end.

The big picture:

- Browser → Caddy: HTTPS (443).

- Caddy → app: HTTP over Docker network (external).

- App → Postgres: TCP inside Docker network (internal).

- Volumes: Postgres data and Caddy TLS state.

- Containers use debian:bookworm-slim (Alpine lacks debugging tools).

- No building inside containers. Cross-compilation to ARM64 happens on dev.

- Go compiles to binary (unlike Js/Ts/Node), needs 10x less RAM (than Js/Ts/Node).

- React solves accessibility/EU compliance via shadcn/ui and Radix UI.

- PostgreSQL is No. 1: [SO, 2025.](https://survey.stackoverflow.co/2025/technology#1-databases)

Each application is fully self-contained using Docker Compose, including its own database. Infrastructure services such as TLS termination (Caddy) are shared per host.

This is a "12-factor app" in the sense of dev and prod being mildly isomorphic and there is a bit of extra-robustness against possible shenanigans with environment variables and variable expansions inside docker-compose.yml. Postgres DNS is assembled in Go for this very reason.

# Part I: Infrastructure

## 1. VPS Setup (Hetzner)

Choose Ubuntu 22.04, create user `deploy` with passwordless sudo, ensure login with ssh keys and passphrase, disable password logins. Optionally set up the ufw rules, see ufw.sh.

Create `/opt` (`root` owned):

VPS:

```bash
sudo mkdir -p /opt
sudo chown root:root /opt
sudo chmod 755 /opt
```

and then `caddy` and the app folder (`initialsdb`), all owned by `deploy`:

```bash
sudo mkdir -p /opt/caddy
sudo mkdir -p /opt/initialsdb
sudo chown -R deploy:deploy /opt/caddy
sudo chown -R deploy:deploy /opt/initialsdb
```

## 2. DNS (Porkbun)

Set up DNS records with www and wildcards @ and \*.

Create Caddyfile with your domain name and point to the app container:

```Caddyfile
initials.dev, www.initials.dev {
    reverse_proxy initialsdb-app:8080
    encode gzip
}
```

## 3. Run Everything Locally First

Clone this repo to dev. On dev:

```bash
cd ~/opt/initialsdb/src
make build
```

This compiles and distributes binaries and their assets (Js) to `bin` and `web`, resp., on `dev` and `prod`.

```bash
cd ~/opt/initialsdb/dev
make up
```

This will create and run all the containers on dev:

```bash
[+] up 5/5
 [+] up 5/5
 ✔ Image initialsdb-dev
 ✔ Network initialsdb_dev_app_net
 ✔ Volume initialsdb_postgres_dev
 ✔ Container initialsdb-dev-db
 ✔ Container initialsdb-dev-app
```

Go to `http://localhost:8080/`, the app should work now.

To simply stop/restart all the containers (data intact), use

```bash
make down
make up
```

This is enough for code updates: make down, rebuild, make up.

If Dockerfile changes (rarely):

```bash
make soft-reset
make up
```

If you are done testing and do not care about any Postgres data (!), nuke it all:

```bash
make hard-reset
```

## 4. Release (VPS)

prod (VPS) is almost identical to dev, except that prod:

- adds a reverse-proxy (Caddy),
- must install and run Posgres backup.
- must be more careful about .secrets (though everything is the same routine).

### 4.1 VPS Preparation

On dev, inside `/prod`, add new passwords to .secrets with

```bash
openssl rand -base64 32
```

Adjust VPS if it already has Makefile and older instance running.

VPS:

```bash
cd /opt/initialsdb
make down
```

This will stop containers and also remove them. Intact: images (build time), volumes (DB data), networks (unless orphaned).

If prod Dockerfile got updated, remove the image, but keep the DB volume intact:

VPS:

```bash
cd /opt/initialsdb
make soft-reset
```

To nuke the whole old app running (including data!):

VPS:

```bash
cd /opt/initialsdb
make hard-reset
```

To nuke the Caddy container and the Docker container network edge_net:

```bash
cd /opt/caddy
make clean
make net-remove
```

### 4.2 Actual Deployment

On dev:

```bash
cd deploy
make copy
make copy-backup-script
make install-backup-cron
```

VPS (if never run before or after `make net-remove`):

```bash
cd /opt/caddy
make net
```

VPS (if Caddyfile updated, skip otherwise):

```bash
cd /opt/caddy
make restart
```

VPS:

```bash
cd /opt/initialsdb
make up
```

This should output:

```bash
[+] up 5/5
 ✔ Image initialsdb-prod
 ✔ Network initialsdb_app_net
 ✔ Volume initialsdb_postgres_prod
 ✔ Container initialsdb-db
 ✔ Container initialsdb-app
```

## 5. Dangerous Commands

They are:

```bash
make hard-reset
make nuke-db
docker volume rm initialsdb_postgres_prod
docker compose down --volumes
```

Use them only if you explicitly want to destroy all data and start everything from scratch!

For normal updates, use `make down` and `make up`, or `make soft-reset` if the prod Dockerfile is updated.

Also, I do not use binds to regular files outside the containers, but if for some reason one does that (see dev/docker-compose.yml.volume-bind), then removing the bind destroys data, i.e.

```bash
sudo rm -rf ./volumes/postgres
```

Initially, I had these binds in dev, but removed them as they turned the data volume (mounted to containers) to some kind of a pointer. Removing a volume would not destroy data, but removing the bind would do that. I did not need any of that.

The Postgres container `initialsdb-db` does not own data, it owns the Postgres process and its file system.

One can make the container gone:

VPS:

```bash
docker stop initialsdb-db
docker rm initialsdb-db
```

The data will remain intact. Postgres stores data in `/var/lib/postgresql/data`, Docker mounts this volume into the container `initialsdb-db`. The containers can be stopped, removed, rebuild with new images, the whole VPS can reboot, the data volume survives.

## 6. .secrets

Inside .gitignore put these lines:

```text
# --------------------------------------------------
# Never commit .secrets (.env fine as they are public)
# --------------------------------------------------
.secrets
.secrets.*
!.secrets.example
!.secrets.*.example
```

Rule order matters. Git applies ignores top to bottom:

- .secrets ignores .secrets.

- .secrets.\* ignores everything starting with .secrets.

- !.secrets.example punches a hole for the base example.

- !.secrets.\*.example punches holes for all variant examples.

So now git won't push .secrets, .secrets.local, .secrets.prod should there be any later. It will commit .secrets.example, .secrets.local.example.

## 7. Adding an Extra (Backup) SSH Key

Just in case, for extra safety, esp. if the dev machine gets ever busted, generate a backup ssh key, add it for use, and also stash it somewhere on non-dev.

All this is very optional, and a bit of a hassle, but it might be useful to know how to deal with multiple SSH keys.
Otherwise, one can also simply stash the main key. Hetzner also provides the VPS recovery with the account credentials.

On dev:

```bash
ssh-keygen -t ed25519 -a 100 \
  -f ~/.ssh/id_ed25519_vps_backup \
  -C "deploy@vps-backup"
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519_vps_backup
ssh-copy-id -f -i ~/.ssh/id_ed25519_vps_backup.pub deploy@vps
```

See my .ssh/config.example with the both keys, it is necessary to replace the old ./ssh/config manually.

VPS:

```bash
cat ~/.ssh/authorized_keys
```

must show two keys.

On dev:

```bash
ssh -i ~/.ssh/id_ed25519_vps vps
ssh -i ~/.ssh/id_ed25519_vps_backup vps
ssh vps
```

All three should succeed.

To see the default key in use, login to VPS with -v:

```bash
ssh -v vps
```

and look for "Offering public key:".

To set only the specific key to use, such as `d_ed25519_vps_backup`:

```bash
ssh-add -D          # remove all keys
ssh-add ~/.ssh/id_ed25519_vps_backup
ssh vps
```

## 8. Inspecting DB on the VPS

VPS:

```bash
cd /opt/initialsdb
docker exec -it initialsdb-db psql -U initialsdb -d initialsdb
```

Inside psql (initialsdb=#):

```sql
-- list tables
\dt

-- describe the listings table
\d listings

-- see some rows
SELECT id, created_at, body, is_hidden
FROM listings
ORDER BY created_at DESC
LIMIT 10;

-- count visible listings
SELECT COUNT(*) FROM listings WHERE is_hidden = false;
```

Exit psql with `\q`.

If the DB credentials change, to quickly get the DB name, user, and password:

VPS:

```
cd /opt/initialsdb
docker exec -it initialsdb-db env | grep POSTGRES
```

A quick direct inspection without getting into the psql prompt:

```bash
cd /opt/initialsdb
docker exec -it initialsdb-db psql -U initialsdb -d initialsdb -c \
"SELECT id, created_at, body FROM listings ORDER BY created_at DESC LIMIT 5;"
```

## 9. VPS Reboot

It changes nothing, tested! What actually happens on VPS reboot:

- Linux boots.

- systemd starts services.

- Docker daemon starts automatically.

- Docker looks at containers it knows about.

- Containers with a restart policy are handled.

- All the containers have `restart: unless-stopped` policy in their docker-compose.yml files.

## 10. Adding a New Application

On dev:

```bash
cd ~/opt
make clone-app SRC=initialsdb DST=yournewapp
```

It will do these:

- copy `initialsdb` while skipping mounts, binaries, .git, .secrets, but not .secrets.example.
- search and replace every occurrence of `initialsdb` with `yournewapp`.

This looks archaic, but it is simple and reliable.

Avoid variable interpolation, nonlocal ../ environments, aliases inside docker-compose.yml.

## 11. Remove Docker Compose?

Docker removes tight coupling with the host OS, but why Docker Compose?

Instead of prod/docker-compose.yml and prod/Makefile we could have one slightly more complicated prod/Makefile:

```Makefile
SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

APP_NAME := initialsdb
APP_IMAGE := initialsdb-prod
APP_CONTAINER := initialsdb-app
DB_CONTAINER := initialsdb-db

APP_NET := initialsdb_app_net
EDGE_NET := edge_net

DB_VOLUME := initialsdb_postgres_prod

ENV_FILES := --env-file .env --env-file .secrets

.PHONY: up down build app db networks volumes clean logs ps

# ---------------------------
# Top-level lifecycle
# ---------------------------

up: networks volumes build db app
	@echo "✅ initialsdb is up"

down:
	docker stop $(APP_CONTAINER) $(DB_CONTAINER) 2>/dev/null || true
	docker rm $(APP_CONTAINER) $(DB_CONTAINER) 2>/dev/null || true
	@echo "🛑 Containers stopped and removed"

clean: down
	docker image rm $(APP_IMAGE) 2>/dev/null || true
	@echo "🧹 Images cleaned"

logs:
	docker logs -f $(APP_CONTAINER)

ps:
	docker ps --filter name=$(APP_NAME)

# ---------------------------
# Infra primitives
# ---------------------------

networks:
	@docker network inspect $(APP_NET) >/dev/null 2>&1 || \
		docker network create --internal $(APP_NET)
	@docker network inspect $(EDGE_NET) >/dev/null 2>&1 || \
		docker network create $(EDGE_NET)
	@echo "🌐 Networks ready"

volumes:
	@docker volume inspect $(DB_VOLUME) >/dev/null 2>&1 || \
		docker volume create $(DB_VOLUME)
	@echo "💾 Volume ready"

# ---------------------------
# Build & run
# ---------------------------

build:
	docker build -t $(APP_IMAGE) .

db:
	docker run -d \
		--name $(DB_CONTAINER) \
		--restart unless-stopped \
		$(ENV_FILES) \
		-v $(DB_VOLUME):/var/lib/postgresql/data \
		--network $(APP_NET) \
		postgres:16-bookworm

	@echo "⏳ Waiting for Postgres to be ready..."
	@until docker exec $(DB_CONTAINER) pg_isready -U initialsdb -d initialsdb >/dev/null 2>&1; do sleep 1; done
	@echo "🐘 Postgres ready"

app:
	docker run -d \
		--name $(APP_CONTAINER) \
		--restart unless-stopped \
		$(ENV_FILES) \
		--network $(APP_NET) \
		--network $(EDGE_NET) \
		-p 8080:8080 \
		$(APP_IMAGE)
```

This is more verbose than docker-compose.yml, yet guaranteed fewer bugs with environment loading and variable expansion. One tool less as well. \*.yml files are tiny, but their debugging time is not.

# Part II: Application Code (initialsdb)

This git repo also includes a complete running application called `initialsdb` which is Go with sqlc and net/http (no frameworks). Go also serves a React SPA, which is a Js artifact as vite + React output in `web`.

The best way to understand the system is to extend it, e.g. add the global counter which will display the number of total messages stored on the landing page above the search bar.

## 1. SQL

To add the global counter, first add the SQL query to db/queries.sql:

```sql
-- name: CountVisibleListings :one
SELECT COUNT(*)::bigint
FROM listings
WHERE is_hidden = FALSE;
```

followed by

```bash
~/opt/initialsdb/src/backend
sqlc generate
```

It will create the code inside db/queries.sql.go:

```go
func (q *Queries) CountVisibleListings(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countVisibleListings)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}
```

This is the brilliance of the sqlc: it is a static generator. Ask AI to write SQL queries, save on tokens as the Go code will be generated by the sqlc. No ORMs, no SQL strings inside Go.

## 2. JSON API Endpoints

Add the endpoint listings/count.go:

```go
package listings

import (
	"context"
	"net/http"
	"time"

	"app.root/db"
	"app.root/httpjson"
)

type CountHandler struct {
	DB *sql.DB
}

func (h *CountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	q := db.New(h.DB)

	n, err := q.CountVisibleListings(ctx)
	if err != nil {
		http.Error(w, "count failed", http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]int64{
		"count": n,
	})
}
```

along with its corresponding setup and call inside routes/routes.go:

```go
mux.Handle("/api/listings/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	(&listings.CountHandler{
		DB: db,
	}).ServeHTTP(w, r)
}))
```

This is a bit of a hassle, but no magic, hence easy to debug.

## 3. Guards (Middleware)

Guards are manually applied per handler, no middleware patterns, no next(). A guard simply outputs true and false which is checked inside a handler's loop over the guards. They are opt-in.

A set of guards per handler/route is hard-coded in routes.go, but the guards can be disabled via their boolean flags inside .env.

The paradigm is:

**Guards protect scarce resources, not endpoints.**

CreateHandler: Handles POST, parses body, mutates DB. Therefore, it is maximally protected with a proof of work (PoW), rate limit, body size guards.

SearchHandler and CountHandler do not use PoW.

PoW is the computation inflicted on the browser, see this line

```ts
const nonce = await solvePoW(...)
```

inside App.tsx. At the moment, it blocks the UI, ignores AbortController, and it cannot be interrupted. If user navigates away, the computation continues until solved.

PoW has two parameters: the difficulty level and the TTL value. The latter cannot be too small as a slower device won't be able to complete the challenge. It can not be too big as the attacker can solve it quickly and then bombard the endpoint with a solved challenge for the remaining TTL time. The recommendation is 2-3x value a slow computer requires for solving. For the difficulty level 21, the TTL is set to 100s.

## 4. Timeouts

Timeout is protection against DB misbehavior, in order not to melt the server due to goroutine pile up.

Each request handler has a timeout. Read/lightweight endpoints - 3s., CreateHandler - 5s. If DB stalls, or there is a network hiccup, the context is canceled: the driver sends a cancellation signal to PostgreSQL or aborts the TCP connection, returns context deadline exceeded, the handler stops waiting, returns 500 or timeout. Otherwise goroutine would hang indefinitely.

Without context, if a goroutine blocks forever, connection pool slot remains busy, eventually the pool exhausts, entire app stalls, we get a cascading failure.

Context cancellation works as a short-circuit:

```
HTTP request
    → handler
        → service layer
            → db
            → redis
            → external API
```

Once it activates, everything downstream stops.

## 5. Frontend

The frontend code is App.tsx.

To fetch the listings count, ChatGPT5 prefers correctness over clarity:

```ts
async function fetchCount(signal?: AbortSignal): Promise<number> {
  const res = await fetch("/api/listings/count", { signal });
  if (!res.ok) throw new Error("count failed");
  const json = await res.json();
  return json.count as number;
}
```

The grand idea here is that

**Effects must be written as if the component can disappear at any time.**

```ts
useEffect(() => {
  const ac = new AbortController();
  fetchCount(ac.signal)
    .then(setTotalCount)
    .catch(() => {});
  return () => ac.abort();
}, []);
```

Note `[]` as the last argument, so the effect gets executed only when App mounts. However, fetch can outlive App. If App unmounts while fetch is in progress, ac.signal becomes "abort" and prevents the execution of `then`. The code jumps into `catch` which does nothing. We have achieved correct fetch abortion and App unmounting, but was it necessary?!

What is annoying about correctness and error handling here is that there are only three API endpoints, and already 500 LOC of frontend, with 24 (!) instances of "abort". ChatGPT5 produces working code here, but this needs some rewriting.

The counter is stored and updated as everything in React:

```ts
const [totalCount, setTotalCount] = useState<number | null>(null);
...
setTotalCount((n) => (n === null ? n : n + 1));
```

Cumbersome, but not as bad as abortion.

# Part III: What Will Survive?

- VPS or PaaS? VPS. 20TB bandwidth @ 5 euros vs Disneyland.

- Postgres: one server per app with Docker guarantees isolation and easier setup in the future.

- Docker Compose is kind of fragile and obfuscating.

- `go build` instead of Node/Deno/Bun. Node is a disaster.

- React is "HTML with accessibility" (via shadCN). However, doing graphics in code always feels wrong for some reason (TikZ in LaTeX, Mermaid, Markdown instead of LibreOffice). We need Blender here, not virtual DOM and "components".

- Fetch and AbortController need to be dropped in favor of something more succinct and automated.

- ChatGPT is incredible in full stack web dev, but it likes to delete working code when adding something new.

VPS, Postgres, Makefile, Dockerfile, Go, sqlc, net/http, vite, React, and README.md with details and a real running app code on github rather than starter templates, GUI panels, serverless services.

One thing about any web app is that it can never be complete. One can endlessly harden routes against bots or simplify code, or optimize/learn PostgreSQL.
