# Stage 1: Build Go binary
FROM golang:1.26.1-alpine AS go-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o server ./cmd/server/

# Stage 2: Build Astro
FROM node:22-alpine AS web-build
WORKDIR /app
COPY web/package.json web/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY web/ .
RUN yarn build

# Stage 3: Final image
FROM alpine:3.20
WORKDIR /app
COPY --from=go-build /app/server .
COPY --from=web-build /app/dist ./web/dist
EXPOSE 8080
CMD ["./server"]
