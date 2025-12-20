FROM golang:1.25.5-alpine3.22 AS backend
WORKDIR /app
COPY src/backend .
ENV CGO_ENABLED=0 GOARCH=amd64 GOOS=linux
RUN go build -mod=vendor -ldflags "-s -w" -o ./build/monitoring_draft_laws ./main.go

FROM node:24.12-alpine3.22 AS frontend
ENV PNPM_HOME="/pnpm" PATH+=":$PNPM_HOME"
RUN corepack enable
WORKDIR /app
COPY src/frontend/package.json src/frontend/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
COPY src/frontend/vite.config.ts src/frontend/index.html src/frontend/eslint.config.ts src/frontend/tsconfig.json ./
COPY src/frontend/public public
COPY src/frontend/src src
RUN pnpm build

FROM alpine:3.22
ARG GIT_BRANCH COMMIT_SHA
LABEL git.repository="https://git.ias.cnd/ias/go/monitoring_draft_laws"
LABEL git.remote.origin.url="git@git.ias.cnd:ias/go/monitoring_draft_laws.git"
LABEL git.remote.origin.branch="${GIT_BRANCH}"
LABEL git.remote.origin.commit_sha="${COMMIT_SHA}"
COPY --from=frontend app/dist /www/data/static
COPY --from=backend app/build/monitoring_draft_laws /usr/local/bin/monitoring_draft_laws
CMD [ "monitoring_draft_laws", "--config=config.json" ]
