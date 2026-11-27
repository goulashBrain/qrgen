# qrgen

Generates **vCard QR codes** from CSV files. Scan a code with a phone camera and it offers to save the contact.

One static Go binary, two modes:

- **CLI**: reads every CSV in a folder and writes one PNG per contact.
- **Server**: small web UI + HTTP API, ready for systemd/nginx or Kubernetes.

Prebuilt binaries for Linux and Windows (amd64/arm64) and a multi-arch container image are published with every release.

---

## Contents

- [CSV format](#csv-format)
- [Install](#install)
- [Usage: CLI](#usage-cli)
- [Usage: server](#usage-server)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Development](#development)
- [Releasing](#releasing)
- [Project structure](#project-structure)
- [Security notes](#security-notes)

---

## CSV format

The header is required. Columns are case-insensitive and can be in any order. Only `first_name` **or** `last_name` is mandatory; missing columns are skipped.

```csv
first_name;last_name;org;title;phone;email
Anna;Bianchi;Example Srl;Engineer;+393400000000;anna.bianchi@example.com
```

| Column | vCard field |
|---|---|
| `first_name`, `last_name` | `N`, `FN` |
| `org` | `ORG` |
| `title` | `TITLE` |
| `phone` | `TEL;TYPE=WORK,VOICE` |
| `email` | `EMAIL;TYPE=INTERNET` |

Notes:

- Default separator is `;` (Excel with Italian locale). Use `-sep ','` for comma-separated files.
- A UTF-8 BOM from Excel is handled automatically; empty rows are skipped.
- Write phone numbers with the country code (`+39...`) so they dial correctly from any phone.
- Codes are vCard 3.0, error correction level Medium. Fewer fields = smaller, easier-to-scan codes.

---

## Install

### From a release (recommended)

Download the archive for your platform from the **Releases** page and verify it:

```bash
V=0.1.1; ARCH=amd64            # arm64 for Raspberry Pi
gh release download v$V -R goulashbrain/qrgen -p "qrgen_${V}_linux_${ARCH}.tar.gz" -p checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf qrgen_${V}_linux_${ARCH}.tar.gz qrgen
sudo install -m 0755 qrgen /usr/local/bin/qrgen
qrgen -version
```

On Windows, download `qrgen_<version>_windows_amd64.zip`, extract it and run `qrgen.exe`.

### From source

Requires Go (version per `go.mod`).

```bash
git clone git@github.com:goulashbrain/qrgen.git && cd qrgen
make build                     # -> bin/qrgen
```

### Container image

```bash
docker run --rm -p 8080:8080 ghcr.io/goulashbrain/qrgen:latest
```

---

## Usage: CLI

Passing `-in` selects CLI mode.

```bash
mkdir -p input && cp contacts.csv input/
qrgen -in ./input -out ./qrcodes
# generated 1 QR codes in ./qrcodes
```

Output files are named `NNN_lastname_firstname.png`, e.g. `001_bianchi_anna.png`.

```bash
qrgen -in ./input -out ./qrcodes -sep ',' -size 1024   # comma CSV, larger images
```

On Windows:

```powershell
.\qrgen.exe -in .\input -out .\qrcodes
```

---

## Usage: server

Without `-in`, qrgen starts an HTTP server (default `:8080`).

```bash
qrgen
# open http://localhost:8080
```

| Method | Path | Input | Output |
|---|---|---|---|
| `GET` | `/` | | Web form |
| `POST` | `/api/qr` | Form fields `first_name`, `last_name`, `org`, `title`, `phone`, `email` | `image/png` |
| `POST` | `/api/batch` | Multipart: `file` (CSV), optional `sep` | `application/zip` |
| `GET` | `/healthz` | | `ok` |

```bash
curl -d first_name=Anna -d last_name=Bianchi http://localhost:8080/api/qr -o me.png
curl -F file=@contacts.csv -F sep=';' http://localhost:8080/api/batch -o qrcodes.zip
```

Limits: CSV upload 1 MiB, max 1000 contacts per batch.

Error responses: `400` invalid input, `413` request too large, `500` internal error (logged).

The server is stateless and writes nothing to disk, so it can run with a read-only root filesystem.

---

## Configuration

| Flag | Env var | Default | Description |
|---|---|---|---|
| `-in` | | *(empty = server mode)* | CLI: input folder with `*.csv` files |
| `-out` | | `./qrcodes` | CLI: output folder (created if missing) |
| `-addr` | `QRGEN_ADDR` | `:8080` | Server: listen address |
| `-sep` | `QRGEN_SEP` | `;` | CSV separator (`,`, `;`, `tab`, ...) |
| `-size` | | `512` | PNG size in pixels |
| `-version` | | | Print version, commit and build date, then exit |

Flags override environment variables.

---

## Deployment

### systemd + nginx

```bash
sudo cp deploy/qrgen.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now qrgen
```

The unit listens on `127.0.0.1:8080`. Expose it through nginx:

```nginx
location /qrgen/ {
    proxy_pass http://127.0.0.1:8080/;   # trailing slash strips the /qrgen prefix
    client_max_body_size 1m;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Open `https://<host>/qrgen/` (keep the trailing slash; the form uses relative URLs).

**Update:** download the new release (see [Install](#from-a-release-recommended)), then `sudo systemctl restart qrgen`.

### Kubernetes

`deploy/k8s.yaml` contains a Deployment, Service and Ingress (no Helm needed). Set the image tag and Ingress host, then:

```bash
kubectl apply -f deploy/k8s.yaml
kubectl rollout status deployment/qrgen
```

**Update to a new version:**

```bash
kubectl set image deployment/qrgen qrgen=ghcr.io/goulashbrain/qrgen:0.1.1
kubectl rollout status deployment/qrgen
kubectl rollout undo deployment/qrgen      # rollback if needed
```

Keep the tag in `deploy/k8s.yaml` in sync with what is running.

If the ghcr.io package is private, create a pull secret and reference it in the Deployment (`imagePullSecrets: [{ name: ghcr }]`):

```bash
kubectl create secret docker-registry ghcr \
  --docker-server=ghcr.io --docker-username=goulashbrain --docker-password=<token-with-read:packages>
```

### Windows

Run `qrgen.exe` directly. To run it as a background service, wrap it with [NSSM](https://nssm.cc/):

```powershell
nssm install qrgen C:\Tools\qrgen\qrgen.exe -addr 127.0.0.1:8080
```

---

## Development

### Requirements

- Go (version per `go.mod`)
- [GoReleaser v2](https://goreleaser.com): `go install github.com/goreleaser/goreleaser/v2@latest`
- Docker with buildx (only for building images locally)

### Common commands

```bash
go test ./...                              # unit tests
go vet ./...
go run ./cmd/qrgen                         # server
go run ./cmd/qrgen -in ./input             # CLI
make build                                 # bin/qrgen with version info
make docker                                # local image from Dockerfile
goreleaser check                           # validate .goreleaser.yaml
goreleaser release --snapshot --clean      # full local build into ./dist, publishes nothing
```

### Branching

| Branch | Purpose |
|---|---|
| `main` | Stable. Releases are tagged here only. |
| `develop` | Integration branch. |
| `feature/<name>` | One branch per change, created from `develop`. |

```bash
git switch develop && git pull
git switch -c feature/<name>
# ...commit...
go test ./... && goreleaser release --snapshot --clean
git push -u origin feature/<name>

git switch develop && git pull
git merge --no-ff feature/<name>
git push
git branch -d feature/<name> && git push origin --delete feature/<name>
```

Commit messages follow a light [Conventional Commits](https://www.conventionalcommits.org) style (`feat:`, `fix:`, `build:`, `docs:`, `test:`, `chore:`). `docs:`, `test:` and `chore:` are excluded from the release changelog.

---

## Releasing

Releases are built by GoReleaser in GitHub Actions (`.github/workflows/release.yml`) whenever a `v*` tag is pushed.

```bash
git switch main && git pull
git merge --no-ff develop -m "Release v0.2.0"
git push
git tag -a v0.2.0 -m "v0.2.0" && git push origin v0.2.0
```

Each release produces:

- `qrgen_<version>_linux_{amd64,arm64}.tar.gz`
- `qrgen_<version>_windows_{amd64,arm64}.zip`
- `checksums.txt`
- Container image `ghcr.io/goulashbrain/qrgen:<version>` and `:latest` (linux/amd64, linux/arm64)

Archives include this README and the `deploy/` folder.

Versioning follows [SemVer](https://semver.org): patch (`v0.2.1`) for fixes, minor (`v0.3.0`) for features. The project stays on `0.x` until the CSV format and API are considered stable.

---

## Project structure

Hexagonal (ports and adapters) layout. Dependencies point inward: `internal/core` imports only the standard library.

```
qrgen/
├── cmd/qrgen/main.go              # composition root: flags, wiring, CLI or server
├── internal/
│   ├── core/
│   │   ├── domain/                # Contact, QRCode, validation, errors
│   │   ├── ports/                 # interfaces (driving + driven)
│   │   └── service/               # use cases, tested with in-memory fakes
│   └── adapters/
│       ├── csvrepo/               # ContactRepository: CSV folder or upload stream
│       ├── vcard/                 # PayloadEncoder: vCard 3.0
│       ├── qrpng/                 # ImageRenderer: skip2/go-qrcode
│       ├── fssink/                # CodeSink: PNG files in a folder
│       ├── zipsink/               # CodeSink: PNGs in a ZIP stream
│       ├── cli/                   # driving adapter: batch mode
│       └── httpapi/               # driving adapter: HTTP + embedded web UI
├── deploy/                        # k8s.yaml, qrgen.service
├── .github/workflows/release.yml  # GoReleaser on tag push
├── .goreleaser.yaml
├── Dockerfile                     # local multi-stage build
├── Dockerfile.goreleaser          # release image (copies prebuilt binary)
└── Makefile
```

| Port | Role | Adapters |
|---|---|---|
| `QRService` | Driving: use cases | `service.QRService` (used by `cli`, `httpapi`) |
| `ContactRepository` | Driven: contact source | `csvrepo.DirRepository`, `csvrepo.StreamRepository` |
| `PayloadEncoder` | Driven: QR payload | `vcard.Encoder` |
| `ImageRenderer` | Driven: image output | `qrpng.Renderer` |
| `CodeSink` | Driven: storage | `fssink.Sink`, `zipsink.Sink` |

To add a format or storage (e.g. MeCard, SVG, SQLite), implement the port in a new adapter package and wire it in `cmd/qrgen/main.go`.

---

## Security notes

- CSV files contain personal data. `*.csv`, `/input/` and `/qrcodes/` are git-ignored; use fake data under `testdata/` for samples.
- The server has **no authentication**. Keep it on the LAN, or put it behind nginx `auth_basic` / Ingress basic-auth if exposed.
- The container runs as non-root on distroless with a read-only root filesystem and all capabilities dropped.
