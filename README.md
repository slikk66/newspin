# news-pin

Pin news articles you want to come back to.

Search [NewsAPI](https://newsapi.org), pin articles to your account, and curate a personal feed. Pins for the special `FRONT_PAGE` user are surfaced as the public featured list.

## Stack

- **Backend** — Go (`net/http`), JWT auth, bcrypt passwords
- **Frontend** — Astro + React + Tailwind (built into static assets, served by the Go binary)
- **Storage** — DynamoDB (users + pins, pay-per-request)
- **Infra** — Pulumi (AWS: ECR, DynamoDB, IAM, GitHub OIDC)
- **Deploy** — Docker → ECR → EKS via Helm repo bump (GitHub Actions)

## Layout

```
cmd/server/      Go HTTP server entrypoint
internal/auth    JWT + bcrypt + middleware
internal/db      DynamoDB client (users, pins)
internal/news    NewsAPI client
internal/handlers HTTP handlers
web/             Astro frontend (built to web/dist)
infra/           Pulumi program (AWS resources, OIDC role)
.github/workflows GitHub Actions (build/push image, bump helm repo)
```

## API

Public:

- `POST /api/register` — `{username, password}`
- `POST /api/login` — returns `{token}` (24h JWT)
- `GET  /api/news?q=...` — search NewsAPI
- `GET  /api/news/featured` — pins under the `FRONT_PAGE` user

Authenticated (`Authorization: Bearer <token>`):

- `GET    /api/pins`
- `POST   /api/pins` — `{articleId, title, url, description, imageUrl, source}`
- `DELETE /api/pins?articleId=...`

Static frontend served from `/` (Astro build).

## Run locally

Requires Go 1.26+, Node 22, yarn, AWS creds for DynamoDB.

```sh
# build frontend
cd web && yarn install && yarn build && cd ..

# env
export USERS_TABLE=newspin-dev-users
export PINS_TABLE=newspin-dev-pins
export NEWS_API_KEY=...
export JWT_SECRET=...

make run
```

Server listens on `:8080`.

## Deploy

Infra (Pulumi, AWS):

```sh
cd infra
PULUMI_CONFIG_PASSPHRASE="" pulumi up
```

App image (manual):

```sh
make deploy AWS_ACCOUNT=... AWS_PROFILE=... TAG=...
```

Push to `main` triggers `.github/workflows/deploy-app.yaml`: builds the image, pushes to ECR, and bumps the image tag in `slikk66/eks-helm` for ArgoCD/Flux to pick up.
