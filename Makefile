include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | \
		sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${INVOXA_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${INVOXA_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	migrate -path ./migrations -database ${INVOXA_DB_DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy and vendor module dependencies, and format and modernize all .go files
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify
	go mod vendor
	go fix ./...
	go fmt ./...

## audit: run quality control checks
.PHONY: audit
audit:
	go mod tidy -diff
	go mod verify
	go vet ./...
	go tool staticcheck ./...
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application for local development
.PHONY: build/api
build/api:
	go build -ldflags='-s' -o=./bin/api ./cmd/api

## build/api-linux: build the cmd/api application for Linux (production)
.PHONY: build/api-linux
build/api-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #
production_host_ip = '104.248.251.249'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh fastvoxa@${production_host_ip}

## production/deploy/api: deploy the api, migrations, and systemd service to production
.PHONY: production/deploy/api
production/deploy/api:
	@echo 'Compiling application for Linux...'
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/linux_amd64/api ./cmd/api
	@echo 'Transferring files to server...'
	rsync -P ./bin/linux_amd64/api fastvoxa@${production_host_ip}:~
	rsync -rP --delete ./migrations fastvoxa@${production_host_ip}:~
	rsync -P ./remote/production/api.service fastvoxa@${production_host_ip}:~
	rsync -P ./remote/production/Caddyfile fastvoxa@${production_host_ip}:~
	@echo 'Executing remote deployment commands...'
	ssh -t fastvoxa@${production_host_ip} '\
		migrate -path ~/migrations -database $$FASTVOXA_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/fastvoxa.service \
		&& sudo mv ~/Caddyfile /etc/caddy/Caddyfile \
		&& sudo systemctl daemon-reload \
		&& sudo systemctl restart fastvoxa \
		&& sudo systemctl reload caddy \
	'