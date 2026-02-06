## 0. Introduction

**initialsdb** is a public bulletin board (initials.dev) implemented as a Go backend serving a React SPA (compiled to static JS/CSS and served by Go), together with PostgreSQL and a Docker-based infrastructure.

This repository includes application code and infrastructure so that one can host more web apps on the same VPS, not just initialsdb.

The big picture:

- Browser → Caddy: HTTPS (443).

- Caddy → app: HTTP over Docker network (external).

- App → Postgres: TCP inside Docker network (internal).

- Volumes: Postgres data and Caddy TLS state.

- Containers use debian:bookworm-slim (Alpine lacks debugging tools).

- No building inside containers. Cross-compilation to ARM64 happens on dev.

- Go compiles to binary (unlike Js/Ts/Node), needs 10x less RAM (than Js/Ts/Node).

- React is minimal and standard (vite + React, no router), it adds accessibility/EU compliance via shadcn/ui and Radix UI.

- Postgres is not as slick as SQLite, but 10x more users out of the box, and ready to shard with plugins.

This is a 12-factor app in the sense of dev and prod being mildly isomorphic and there is a bit of extra-robustness against possible shenanigans with environment variables and variable expansions inside docker-compose.yml. Postgres DNS is assembled in Go for this very reason.

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
mkdir -p volumes/postgres
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

If you are done testing and do not care about any Postgres data (!), nuke it all:

```bash
make clean
```

To simply stop/restart all the containers (data intact), use

```bash
make down
make up
```

This is enough for code updates: make down, rebuild, make up. If Dockerfile changes (rarely):

```bash
make reset
```

## 4. Release (VPS)

prod (VPS):

- adds a reverse-proxy (Caddy),
- must install and run Posgres backup.

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

This will stop everything app (initialsdb) related, except Caddy, data untouched.

To nuke the whole old app running (including data!):

VPS:

```bash
cd /opt/initialsdb
make clean
```

On dev:

```bash
cd deploy
make copy
make copy-backup-script
make install-backup-cron
```

VPS (if never run before):

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

## 5. .secrets and Dangerous Commands

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

Any of these commands irreversibly deletes the database:

```bash
make nuke-db
make clean
```

Use them only if you explicitly want to destroy all data and start everything from scratch.

For normal updates, use `make down` and `make up`, cd into folders and check what is doable with Makefile.

## 6. Extras (TBC)

1. How to add another app (cp, grep, sed, no interpolation inside docker-compose.yml).

2. How to update Postgres password without nuking data.

3. How to nuke the whole VPS (all the containers and data).
