# 🐳 Dockwatch

> Bring Kubernetes-style secret & config reload automation to Docker & Docker Compose

![Dockwatch Logo](/home/user/.gemini/antigravity/brain/c1ac7e7a-01d5-47aa-8e0b-552a8a33908d/dockwatch_logo_1774327264407.png)

⭐ If this project helps you, please give it a star!

---

## 🔥 Why This Tool Matters

In the Kubernetes world, tools like Stakater Reloader are standard for keeping applications in sync with their configurations. However, **Docker and Docker Compose lack this native capability**.

**The Pain Points:**
- ❌ **No auto-reload** when mounted `ConfigMaps`/secrets change.
- ❌ **Manual restarts** are constantly required after editing `.env` files or certificates.
- ❌ **Security risks** arise when old API keys or stale secrets remain in memory because someone forgot to restart the container.

**The Solution:**
Dockwatch seamlessly watches your host or volume-mounted files and automatically applies changes to running containers using Restart, Signal, or HTTP webhooks—driven entirely by native Docker Labels. No bespoke sidecars required.

---

## ⚡ Key Use Cases

Automate your infrastructure operations instantly:
- 🔐 **TLS Certificate Renewal:** Automatically send `SIGHUP` to Nginx/Traefik when Let's Encrypt renews your certificates.
- 🔑 **API Key & Password Rotation:** Safely trigger an app reload when backend database credentials or third-party API keys are rotated.
- ⚙️ **Zero-Downtime Config Updates:** Hot-reload application configuration logic on the fly without dropping active user connections.

---

## 🏗️ Architecture

```text
 +----------------+           +-------------------------+         +------------------+
 |                |           |                         |         |                  |
 | Mounted Files  |--(watch)--|        Dockwatch        |--(API)--|  Docker Daemon   |
 |/secrets, /certs|           |                         |         |/var/run/docker...|
 +----------------+           +-------------------------+         +------------------+
         |                                 |                               |
         v                                 v                               v
  SHA256 Validated                   Event Engine                   Container Action
(Content vs Timestamp)           (Debounce & Cooldown)           (Restart / Kill / Exec)
```
*The Watcher Engine monitors files via `fsnotify`. Upon a write, it uses SHA256 hashing to guarantee the content changed before querying the Docker API for matching containers and dispatching the configured action.*

---

## ⚙️ Supported Reload Strategies

Choose the right strategy for your application via the `reloader.mode` label:

1. 🔄 **`restart` (Default)** 
   - **Safest & Universal:** Performs a graceful Docker container restart.
   - *When to use:* Your application does not support hot-reloading or you updated environment variables.
2. ⚡ **`signal`**
   - **Zero Downtime:** Sends a specific UNIX signal (e.g., `SIGHUP`, `SIGUSR1`) to the container's PID 1.
   - *When to use:* Applications designed to catch signals and reload their own config (like Nginx, Prometheus, or Node.js apps handling custom signals).
3. 🌐 **`http`**
   - **Advanced Integration:** Executes an HTTP `POST` or `GET` webhook inside the container.
   - *When to use:* Frameworks like Spring Boot (`/actuator/refresh`) or custom microservices exposing internal reload endpoints.

---

## 🚀 Quick Start

### 1. Run as a Docker Compose Sidecar

Here's an example of automatically reloading Nginx when its config changes.

```yaml
version: '3.8'

services:
  app:
    image: nginx:alpine
    labels:
      reloader.enable: "true"
      reloader.mode: "signal"
      reloader.signal: "SIGHUP"
      reloader.watch: "/etc/nginx/nginx.conf"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    ports:
      - "8080:80"

  reloader:
    image: dockwatch:latest
    volumes:
      # Required for finding containers and sending signals:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      # Required to watch the exact same files the target mounts:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    command: ["--watch-dir=/etc/nginx", "--log-level=debug"]
```

### 2. Modify Your File

Edit `./nginx.conf`. Dockwatch will instantly detect the change, verify its SHA256 hash, and apply the `SIGHUP` signal to your Nginx container seamlessly.

---

## 🏷️ Configuration Labels

Target containers are exclusively managed via Docker labels:

| Label | Required | Default | Description |
|---|---|---|---|
| `reloader.enable` | **Yes** | `false` | Enable automatic reloading for this container. |
| `reloader.mode` | No | `restart` | Action to take: `restart`, `signal`, or `http`. |
| `reloader.signal` | No | `SIGHUP` | UNIX signal to send (if `mode=signal`). |
| `reloader.watch` | No | *None* | Comma-separated list of absolute paths to monitor. |
| `reloader.endpoint` | No | *None* | Webhook URL to hit inside the container (if `mode=http`). |

---

## 📊 Deployment Modes

1. **Standalone Container:** Run it globally on a Docker host to monitor common mounts (e.g., `/etc/ssl/certs`).
2. **Compose Sidecar:** Include it per-stack in `docker-compose.yml` scoped to that specific application's configs.
3. **Docker Swarm Mode:** *(Coming soon - see roadmap)*

---

## 🔐 Security Considerations

**⚠️ Docker Socket Risk**
Giving a container access to `/var/run/docker.sock` implies root-level access to the host. 
- Always mount the socket as **read-only (`:ro`)**.
- For highly secure environments, consider using a **Docker Socket Proxy** (like `tecnativa/docker-socket-proxy`) to strictly limit API access to just `GET /containers` and `POST /containers/{id}/restart|kill|exec`.

---

## 🧠 Limitations

Transparency is key. Before deploying, please note:
- **Environment Variables:** Changing `.env` files modifying container environments *requires* a container recreation by `docker compose up`, which this tool does not handle. It handles mounted files.
- **Application Support:** `signal` and `http` modes strictly require your application to be programmed to understand and react to them.
- **Volume Mounts:** Reloader needs access to the *host paths* or shared volumes to watch them with `fsnotify`.

---

## 📈 Observability (Prometheus)

This tool exposes basic Prometheus metrics deeply integrating into your DevOps observability stack natively on `:9090/metrics`.

Key exported metrics:
- `reloader_actions_total{container_name, action, status}`: Tracks successful and failed reloads per container and strategy.
- `reloader_file_changes_total{path}`: Tracks raw `fsnotify` events that passed SHA256 validation.

---

## 🥊 Kubernetes Reloader vs. Dockwatch

| Feature | K8s Stakater Reloader | Dockwatch |
|---|---|---|
| **Environment** | Kubernetes Clusters | Docker, Compose, Single-Node |
| **Requirements** | Master Node / Control Plane | **Lightweight (No cluster needed)** |
| **Triggers** | ConfigMap / Secret | File system events (fsnotify) |
| **Action** | Rolling Pod Restart | Restart, Unix Signal, HTTP Webhook |
| **Configuration** | Setup via Annotations | Setup via Docker Labels |

---

## 🤝 Contributing

Contributions are heavily encouraged! 

**Code Structure Overview:**
- `cmd/reloader/`: The main entrypoint.
- `internal/watcher/`: Uses `fsnotify` safely wrapped with SHA256 checks.
- `internal/docker/`: Docker client wrapper for API discovery and triggers.
- `internal/reloader/`: The core orchestration, debouncing, and rate-limiting.

**PR Guidelines:**
1. Fork and create a feature branch.
2. Provide test coverage for new functionality (`go test ./...`).
3. Maintain zero runtime dependencies outside of standard Go + Docker SDK.

---

## 🗺️ Roadmap

- [ ] Kubernetes CRD-like centralized config (Yaml-based config for hosts)
- [ ] Web UI dashboard (Basic React view for active watchers)
- [ ] External secret integration (HashiCorp Vault / AWS Secrets Manager)
- [ ] Dependency-aware reloads (Wait for DB before restarting App)
- [ ] Docker Swarm mode native global service scaling

---
*Made with ❤️ for the DevOps community.*
