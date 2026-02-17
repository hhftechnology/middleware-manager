<div align="center">
    <h1 align="center"><a href="https://github.com/hhftechnology/middleware-manager-traefik">Traefik/Pangolin Middleware Manager</a></h1>

[![Docker](https://img.shields.io/docker/pulls/hhftechnology/middleware-manager?style=flat-square)](https://hub.docker.com/r/hhftechnology/middleware-manager)
![Stars](https://img.shields.io/github/stars/hhftechnology/middleware-manager?style=flat-square)
[![Discord](https://img.shields.io/discord/994247717368909884?logo=discord&style=flat-square)](https://discord.gg/HDCt9MjyMJ)
[![Documentation](https://img.shields.io/website?url=https://middleware-manager.hhf.technology&up_message=Please%20Visit%20the%20docs%20for%20in%20depth%20details&up_color=Green&down_message=Docs%20are%20getting%20Updated&down_color=pink&style=flat-square)](https://middleware-manager.hhf.technology)

</div>

<h3 align="center">Comprehensive Middleware, Router, and Service Management for Traefik</h3>

<div align="center">
  <h5>

<img width="1920" height="1440" alt="dashboardV4" src="https://github.com/user-attachments/assets/3d05dc41-94a5-4167-a2d0-f0561c113902" />
<img width="1920" height="1440" alt="resources" src="https://github.com/user-attachments/assets/07efdb25-b012-4fd3-b0a4-32044806b9f2" />
<img width="1920" height="1440" alt="Resource-Manage-1_so" src="https://github.com/user-attachments/assets/0091c6a9-a3f4-4801-bacf-7a16915cfa13" />
<img width="1920" height="1440" alt="Resource-Manage_so" src="https://github.com/user-attachments/assets/904eae11-d79a-40f0-ba6b-edbbbe53a926" />
<img width="1920" height="1440" alt="Middleware_ManageV4" src="https://github.com/user-attachments/assets/0fd2b5f9-8f59-4f8d-a40d-a8243466a605" />
<img width="1920" height="1440" alt="Middleware-Config_so" src="https://github.com/user-attachments/assets/f35a6cd6-43f5-4a03-9000-14d98c5ffdab" />
<img width="1920" height="1440" alt="Services_ManageV4" src="https://github.com/user-attachments/assets/032d4cda-04d0-4957-95b7-7291c3a18228" />
<img width="1920" height="1440" alt="Security-Main_so" src="https://github.com/user-attachments/assets/31013be9-e6ed-4d8d-b316-49d4ccf6ef90" />
<img width="1920" height="1440" alt="Security-Setup_so" src="https://github.com/user-attachments/assets/6e903478-e5b9-4415-9b7d-7f00eea9aa43" />
<img width="1920" height="1440" alt="Security-1_so" src="https://github.com/user-attachments/assets/a6de3b5e-8317-45b8-be40-dd690f6990ac" />
<img width="1920" height="1440" alt="Security-2_so" src="https://github.com/user-attachments/assets/e31be3ea-6a02-4ac0-aa36-53fe0c112660" />
<img width="1920" height="1440" alt="Security-3_so" src="https://github.com/user-attachments/assets/553a2299-d25f-466f-a1d6-7088de65c8c8" />
<img width="1920" height="1440" alt="Security-4_so" src="https://github.com/user-attachments/assets/1a9697a6-9b17-4daa-ae4a-a9ad915e0864" />
<img width="1920" height="1440" alt="Security-5_so" src="https://github.com/user-attachments/assets/f724608f-8a17-4e73-a1b7-db7ea1cd1f80" />
<img width="1920" height="1440" alt="plugin_hubV4-1" src="https://github.com/user-attachments/assets/f45cecb9-3c9a-4d1b-a3b6-d9662bfc7ca8" />
<img width="1920" height="1440" alt="plugin_hubV4-2" src="https://github.com/user-attachments/assets/9389dd41-438f-423f-a5b6-4ccc109d2e80" />
<img width="1920" height="1440" alt="648shots_so" src="https://github.com/user-attachments/assets/27cfe275-df09-4cf9-a7e8-d7265cfc76ad" />
<img width="5852" height="2740" alt="diagram" src="https://github.com/user-attachments/assets/d69eecbd-0502-4057-94eb-97027c6c7202" />

      
  </h5>
</div>

## Overview

Middleware Manager lets you attach Traefik middlewares, manage custom services, tune routers, and control plugins/mTLS from a single UI. It works with **Pangolin** or directly with **Traefik** as the data source and automatically generates Traefik dynamic configuration in api and creates an override api. it will not disturb the original api(if you revert the original api in the traefik file it will restore all default router and services.

## Middleware-Manager container has to be kept running in order to keep the overide middleware deployed. 

### Who this is for

*   Operators deploying with Pangolin or standalone Traefik.
*   Platform teams managing routers, services, plugins, and mTLS.
*   Contributors extending the Go API or React/Vite UI.

### What you can do

*   **Discover resources** from Pangolin or Traefik and override routers/services safely.
*   **Create and assign Traefik middlewares** with priorities and templates.
*   **Define custom services** (load balancer, weighted, mirroring, failover).
*   **Install/manage plugins** (writes to Traefik static config) and restart Traefik as needed.
*   **Enable mTLS** via the `mtlswhitelist` plugin with per-resource rules.
*   **Inspect Traefik routers/services/middlewares** via the built-in explorer.

---

## Quick Start Guide

### 1. Prerequisites

*   Traefik v2/v3 reachable from Middleware Manager.
*   Docker runtime for the provided containers.
*   Mount points for database and Traefik rules.

### 2. Choose Deployment Path

#### Option A: Deploy with Pangolin (Recommended)

Use this when Pangolin already manages Traefik and you want Middleware Manager to add middlewares, custom services, and plugin/mTLS control.

```yaml
middleware-manager:
  image: hhftechnology/middleware-manager:latest
  volumes:
    - ./data:/data
    - ./config/middleware-manager:/app/config
    - ./config/traefik:/etc/traefik         # Must match Traefik's static config dir
  environment:
    - ACTIVE_DATA_SOURCE=pangolin
    - PANGOLIN_API_URL=http://pangolin:3001/api/v1
    - TRAEFIK_STATIC_CONFIG_PATH=/etc/traefik/traefik_config.yml # Path inside container
    - TRAEFIK_CONF_DIR=/conf
    - DB_PATH=/data/middleware.db
    - PORT=3456
  ports:
    - "3456:3456"
```

#### Option B: Deploy with Traefik (Standalone)

Use this when Traefik is your only provider and you want Middleware Manager to own dynamic config and plugins/mTLS.

```yaml
middleware-manager:
  image: hhftechnology/middleware-manager:latest
  volumes:
    - ./middleware_manager_data:/data
    - ./traefik_config/static:/etc/traefik  # Must match Traefik's static config dir
    - ./middleware_manager_config:/app/config
  environment:
    - ACTIVE_DATA_SOURCE=traefik
    - TRAEFIK_API_URL=http://traefik:8080
    - TRAEFIK_CONF_DIR=/conf
    - TRAEFIK_STATIC_CONFIG_PATH=/etc/traefik/traefik.yml
    - DB_PATH=/data/middleware.db
    - PORT=3456
  ports:
    - "3456:3456"
```

**Note:** Ensure `TRAEFIK_STATIC_CONFIG_PATH` points to the static file *inside the Middleware Manager container*. If the path is wrong, plugin install/removal will fail silently.

### 3. Configure Traefik Static Config

To allow Middleware Manager to dynamically update Traefik, you **must** configure the HTTP provider in your `traefik_config.yml` (or `traefik.yml`).

Replace any existing Pangolin endpoints with the Middleware Manager API endpoint:

```yaml
providers:
  http:
    endpoint: "http://middleware-manager:3456/api/v1/traefik-config"
    pollInterval: 5s
```

**Troubleshooting Connection:**
If the container name `middleware-manager` is not reachable (e.g., DNS issues), find the container's IP address:

```bash
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' middleware-manager
```

Then use the IP in your config:
```yaml
endpoint: "http://172.17.1.3:3456/api/v1/traefik-config"
```

#### Full Example `traefik_config.yml`

<details>
<summary>Click to view full example config</summary>

```yaml
accessLog:
    bufferingSize: 100
    fields:
        defaultMode: drop
        headers:
            defaultMode: drop
            names:
                Authorization: redact
                Content-Type: keep
                Cookie: redact
                User-Agent: keep
                X-Forwarded-For: keep
                X-Forwarded-Proto: keep
                X-Real-Ip: keep
        names:
            ClientAddr: keep
            ClientHost: keep
            DownstreamContentSize: keep
            DownstreamStatus: keep
            Duration: keep
            RequestMethod: keep
            RequestPath: keep
            RequestProtocol: keep
            RetryAttempts: keep
            ServiceName: keep
            StartUTC: keep
            TLSCipher: keep
            TLSVersion: keep
    filePath: /var/log/traefik/access.log
    filters:
        minDuration: 100ms
        retryAttempts: true
        statusCodes:
            - 200-299
            - 400-499
            - 500-599
    format: json
api:
    dashboard: true
    insecure: true
certificatesResolvers:
    letsencrypt:
        acme:
            caServer: https://acme-v02.api.letsencrypt.org/directory
            email: your-email@hhf.technology
            httpChallenge:
                entryPoint: web
            storage: /letsencrypt/acme.json
entryPoints:
    tcp-51821:
        address: :51821/tcp
        http:
            middlewares:
                - crowdsec@file
            tls:
                certResolver: letsencrypt
        transport:
            respondingTimeouts:
                readTimeout: 30m
    udp-51830:
        address: :51830/udp
    web:
        address: :80
    websecure:
        address: :443
        http:
            tls:
                certResolver: letsencrypt
        transport:
            respondingTimeouts:
                readTimeout: 30m
experimental:
    plugins:
        badger:
            moduleName: github.com/fosrl/badger
            version: v1.2.0
        crowdsec:
            moduleName: github.com/maxlerebourg/crowdsec-bouncer-traefik-plugin
            version: v1.4.4
        mtlswhitelist:
            moduleName: github.com/smerschjohann/mtlswhitelist
            version: v0.3.0
log:
    compress: true
    filePath: /var/log/traefik/traefik.log
    format: json
    level: INFO
    maxAge: 3
    maxBackups: 3
    maxSize: 100
ping:
    entryPoint: web
providers:
    file:
        filename: "/etc/traefik/dynamic_config.yml"
    http:
        endpoint: http://172.17.1.3:3456/api/v1/traefik-config
        pollInterval: 5s
serversTransport:
    insecureSkipVerify: true
```
</details>

### 4. First Boot Checklist

1.  **Start the stack**: Run `docker compose up -d`.
2.  **Open the UI**: Navigate to `http://<host>:3456`.
3.  **Test Connection**: Go to **Settings → Test Connection** for the active data source.
4.  **Verify Dashboard**: Check that Routers/Services/Middlewares counts are loaded.
5.  **Explore**: Visit **Resources** and open any resource to confirm router details.

## Documentation

For full guides on configuration, security, operations, and API usage, please refer to the contents of the `docs` folder or the hosted documentation.

## Development

### Backend
```bash
go run main.go
# Build: go build -o middleware-manager main.go
```

### Frontend
```bash
cd ui
cp src/package.json .
npm install
npm start
```

## Inspiration 

Api shim was ispired by @AstralDestiny. We love her. she is a gem of a dev and always be a friend.
https://github.com/AstralDestiny
https://gitlab.com/Astra.Destiny

## License

MIT License

## Documentation

[![Documentation](https://img.shields.io/website?url=https://middleware-manager.hhf.technology&up_message=Please%20Visit%20the%20docs%20for%20in%20depth%20details&up_color=Green&down_message=Docs%20are%20getting%20Updated&down_color=pink&style=flat-square)](https://middleware-manager.hhf.technology)



