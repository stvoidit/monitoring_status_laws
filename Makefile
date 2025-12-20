SHELL:=/usr/bin/bash
NMV_SCRIPT:="${HOME}/.nvm/nvm.sh"

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

default:
	$(info db: "${PGHOST}:${PGPORT}/${PGDATABASE}")

## development
run-dev.backend:
	cd src/backend && go run main.go --config=config.json
run-dev.frontend:
	cd src/frontend && . ${NMV_SCRIPT} && nvm use && pnpm dev

check.docs:
	cd src/backend && go run main.go --checkdocs


container.build:
	@podman build --no-cache \
		--build-arg VERSION_TAG=$$(git describe --tags) \
		--build-arg=COMMIT_SHA=$$(git rev-parse --short HEAD) \
		--build-arg=GIT_BRANCH=$$(git branch --show-current) \
		--file Dockerfile .
container.run:
	@podman run -p 8080:8080 --rm --name mdf \
		-v /etc/localtime:/etc/localtime:ro \
		-v /etc/timezone:/etc/timezone:ro \
		-v ./config.json:/config.json:ro \
		-e DEBUG=${DEBUG} \
		-e TG_BOT_TOKEN=${TG_BOT_TOKEN} \
		-e SERVER_DOMAIN=${SERVER_DOMAIN} \
		-e DB_PREFIX=${DB_PREFIX} \
		--build-arg COMMIT_SHA=${CI_COMMIT_SHORT_SHA:-dev} \
		--build-arg GIT_BRANCH=${CI_COMMIT_BRANCH:-dev} \
		monitoring_draft_laws:dev



ddl:
	pg_dump --host=${PGHOST} --port=${PGPORT} --username=${PGUSER} --dbname=megaplan --schema=monitoring_draft_laws -x --format=plain --schema-only --no-owner --create > monitoring_draft_laws.sql

ddl.init:
	psql --dbname=postgres --host=${PGHOST} --port=${PGPORT} --username=${PGUSER} -a -f 'monitoring_draft_laws.sql'

db.dump:
	$(info "${PGHOST}:${PGPORT}/${PGDATABASE}")
	pg_dump --verbose \
		--host=${PGHOST} --port=${PGPORT} \
		--username=${PGUSER} --format=c --compress=8 --encoding=UTF-8 \
		--no-privileges --file ../${PGDATABASE}.dump ${PGDATABASE}

db.restore:
	$(info "${PGHOST}:${PGPORT}/${PGDATABASE}")
	pg_restore \
		--host=${PGHOST:-localhost} --port=${PGPORT:-5432} --username=${PGUSER} \
		--create --clean --if-exists --format=c --dbname=postgres \
		--no-owner --verbose -x ../${PGDATABASE}.dump
