# Docker Dashboard

![Dashboard Screenshot](docs/images/docker-dashboard.png)

## Description

**Docker Dashboard** is a simple web interface for real-time monitoring of Docker containers. The project allows you to view the list of running containers, their status, tags, creation time, uptime, labels, and other useful information.

## Features

- View all running Docker containers
- Display container details: name, ID, image, tags, creation time, status, health, uptime, labels, and more
- Automatic data refresh every 5 seconds
- Web interface built with Vue.js (SPA)
- Backend in Go using Gorilla Mux and fsouza/go-dockerclient

## Architecture

- **Backend (Go):**
  - REST API to fetch container information (`/api/containers`)
  - Interacts with Docker via go-dockerclient
  - Serves static frontend files
- **Frontend (Svelte):**
  - SPA displaying containers and their parameters
  - Automatic polling for updates

## Quick Start

### Requirements
- Docker
- Docker Compose (optional)
- Linux (recommended, requires access to `/var/run/docker.sock`)

### Build Docker Image

```sh
docker build . -t docker-dashboard:latest -f build/Dockerfile
```

### Run Container

```sh
docker run -it --publish 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock:ro docker-dashboard:latest
```

### Run with Docker Compose

```yaml
version: '3.8'
services:
  dashboard:
    image: avt0x/docker-dashboard:latest
    container_name: docker_dashboard
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

### Access the Dashboard

Open your browser: [http://localhost:8080](http://localhost:8080)

## Environment Variables

- `LABEL_PREFIX` — prefix for filtering container labels (default: `org.example`)

## API

- `GET /api/containers` — get a list of containers with detailed information

## Dependencies

- Go 1.22+
- [github.com/fsouza/go-dockerclient](https://github.com/fsouza/go-dockerclient)
- [github.com/gorilla/mux](https://github.com/gorilla/mux)
- Vue.js (CDN)
- Axios (CDN)

## License

MIT 