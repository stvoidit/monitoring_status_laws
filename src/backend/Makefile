TERN_MIGRATIONS=${CURDIR}/store/migrations

.PHONY: migrate tern-debug

tern-debug:
	$(info "${TERN_MIGRATIONS}")
	@tern print-connstring


migrate: tern-debug
	@tern migrate -m ${TERN_MIGRATIONS}

migrate.undo: tern-debug
	@tern migrate --destination -1 -m ${TERN_MIGRATIONS}

migrate.init:
	@tern init

migrate.gengen:
	@tern gengen > ${TERN_MIGRATIONS}/generate-migrations.sql
