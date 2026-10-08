#!/usr/bin/env bash
# Roda a suíte do backend, incluindo o teste de API ponta a ponta, contra um MariaDB local.
# Defina TEST_DATABASE_DSN apontando para um banco DEDICADO (nome terminando em _test): ele é migrado
# e apagado pelos testes. Exemplo:
#   sudo mariadb -e "CREATE DATABASE homealias_test; GRANT ALL ON homealias_test.* TO 'homealias_test'@'127.0.0.1' IDENTIFIED BY 'x'"
#   TEST_DATABASE_DSN='homealias_test:x@tcp(127.0.0.1:3306)/homealias_test?parseTime=true&multiStatements=true' backend/tests/run-api-tests.sh
set -euo pipefail
: "${TEST_DATABASE_DSN:?defina TEST_DATABASE_DSN (banco dedicado terminando em _test)}"
cd "$(dirname "$0")/.."
go test ./tests/... -count=1 "$@"
