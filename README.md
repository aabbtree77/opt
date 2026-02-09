## 0. Introduction

**initialsdb** is a public bulletin board (message store) implemented as a Go backend serving a React SPA, together with PostgreSQL and Docker Compose infrastructure. It uses proof of work and rate limiting to fight bots.

This repository includes application code and infrastructure so that one can host more such web apps on the same VPS, not just initialsdb. The included precise working Makefile, Dockerfile, docker-compose.yml files and this readme.md show how to replicate the whole system end-to-end.

The big picture:

- Browser → Caddy: HTTPS (443).

- Caddy → app: HTTP over Docker network (external).

- App → Postgres: TCP inside Docker network (internal).

- Volumes: Postgres data and Caddy TLS state.

- Containers use debian:bookworm-slim (Alpine lacks debugging tools).

- No building inside containers. Cross-compilation to ARM64 happens on dev.

- Go compiles to binary (unlike Js/Ts/Node), needs 10x less RAM (than Js/Ts/Node).

- React solves accessibility/EU compliance via shadcn/ui and Radix UI.

- Postgres is more web-ready than SQLite.

Each application is fully self-contained using Docker Compose, including its own database, enabling reproducible deployments and strict isolation. Infrastructure services such as TLS termination are shared per host.

This is a "12-factor app" in the sense of dev and prod being mildly isomorphic and there is a bit of extra-robustness against possible shenanigans with environment variables and variable expansions inside docker-compose.yml. Postgres DNS is assembled in Go for this very reason.

I have tried to keep this flat and simple, but Postgres and Docker Compose are simple only in the perfect world. We lose a single binary ideal, we add .yml layers. This is a serious trade-off. Reproducibility and isolation, at the expense of debugging inside a box inside a box...

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

## 7. Adding a New Application

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

## 8. Adding an Extra (Backup) SSH Key

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

## 9. initialsdb

An example app is `initialsdb` which is Go with sqlc and net/http (no frameworks). Go also serves the React SPA, which is a Js artifact from vite + React placed in the `web` folder.

Whenever db/queries.sql are updated, say adding the global counter

```sql
-- name: CountVisibleListings :one
SELECT COUNT(*)::bigint
FROM listings
WHERE is_hidden = FALSE;
```

one must run

```bash
~/opt/initialsdb/src/backend
sqlc generate
```

which will create the code inside db/queries.sql.go:

```go
func (q *Queries) CountVisibleListings(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countVisibleListings)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}
```

This is the brilliance of the sqlc: it is a static generator. Ask AI to write SQL queries, save on tokens as the Go code will be generated by the sqlc. No ORMs, no SQL strings inside Go.

## 10. Inspecting DB on the VPS

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

## 11. VPS Reboot

It changes nothing, tested! What actually happens on VPS reboot:

- Linux boots.

- systemd starts services.

- Docker daemon starts automatically.

- Docker looks at containers it knows about.

- Containers with a restart policy are handled.

- All the containers have `restart: unless-stopped` policy in their docker-compose.yml files.

## 12. Bare Linux vs Docker vs Docker Compose

### 12.1 Classical Postgres on the VPS (systemd), Shared by Multiple Apps

One OS-level Postgres service, managed by systemd. Multiple apps connect via TCP. Apps may or may not be containerized.

Pros:

- Maximum debuggability: psql, logs, configs are first-class citizens.

- Lower operational complexity.

- Lower memory footprint.

Cons:

- Tight coupling to the host.

- Harder to migrate cleanly.

- Version upgrades affect all apps.

- Less isolation (one bad migration hurts more).

Best use case: when you are the DBA. This is the “Unix adult” setup.

### 12.2 Dockerized Postgres (Single Container), Multiple Apps

Postgres runs in Docker. One DB container. Apps connect over Docker network or localhost.

Pros:

- Reproducible environment.

- Easy upgrades via image pinning.

- Cleaner host system.

- One place to back up.

Cons:

- Still shared fate between apps.

- Debugging slightly indirect.

- Container lifecycle tied to DB uptime.

- Volume handling must be respected.

Best use case: mixed environments (dev, staging, prod).

### Docker Compose per App (This Setup)

One docker-compose.yml per dev and prod inside app. Each app has: its own Postgres, its own network. One shared Caddy container per VPS.

Pros:

- Maximum isolation.

- Each app is fully self-contained.

- Easy teardown and rebuild.

- Environment parity (dev ⇄ prod).

Cons:

- Higher resource usage.

- More layers.

Best use case: SaaS prototypes.
