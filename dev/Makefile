# Makefile for GoSight dev setup

REPO=https://github.com/aaronlmathis/gosight.git
PG_CONTAINER=postgres-dev
PG_USER=gosight
PG_PASS=devpassword
PG_DB=gosight
PG_PORT=5432
VM_IMAGE=victoriametrics/victoria-metrics:v1.115.0
VM_CONTAINER=victoriametrics-dev
VM_PORT=8428

OS := $(shell uname -s)
PKG_MANAGER := $(shell command -v apt >/dev/null 2>&1 && echo apt || (command -v dnf >/dev/null 2>&1 && echo dnf))

.PHONY: all setup clone postgres vm schema certs build clean

all: setup clone postgres vm schema certs build

setup:
	@echo "Installing required packages..."
ifeq ($(PKG_MANAGER),apt)
	sudo apt update && sudo apt install -y git curl golang postgresql-client build-essential libsystemd-dev openssl
endif
ifeq ($(PKG_MANAGER),dnf)
	sudo dnf install -y git curl golang postgresql gcc make systemd-devel openssl
endif

clone:
	@echo "Cloning GoSight repo..."
	git clone --depth=1 --branch main $(REPO) gosight
	mkdir -p gosight/configs
	cp gosight/agent/config.yaml gosight/configs/agent.yaml
	cp gosight/server/config.yaml gosight/configs/server.yaml

postgres:
	@echo "🐘 Starting Postgres container..."
	docker run -d --rm --name $(PG_CONTAINER) \
		-e POSTGRES_USER=$(PG_USER) \
		-e POSTGRES_PASSWORD=$(PG_PASS) \
		-e POSTGRES_DB=$(PG_DB) \
		-p $(PG_PORT):5432 \
		postgres:15
	@sleep 5

vm:
	@echo "Running VictoriaMetrics..."
	mkdir -p victoria-metrics-data
	docker pull $(VM_IMAGE)
	docker run -d --rm --name $(VM_CONTAINER) \
		-v $(CURDIR)/victoria-metrics-data:/victoria-metrics-data \
		-p $(VM_PORT):8428 \
		$(VM_IMAGE) --selfScrapeInterval=5s -storageDataPath=/victoria-metrics-data

schema:
	@echo "Loading DB schema..."
	PGPASSWORD=$(PG_PASS) psql -h localhost -U $(PG_USER) -d $(PG_DB) -p $(PG_PORT) -f gosight/dev/gosight_db_schema.sql

certs:
	@echo "Generating certs..."
	mkdir -p gosight/certs
	cd gosight && bash install/generate_certs_with_san.sh

build:
	@echo "Building GoSight binaries..."
	cd gosight && \
		go work init && \
		go work use ./shared ./server ./agent && \
		GOSIGHT_AGENT_CONFIG=./configs/agent.yaml go build ./agent/cmd -o ./agent/gosight-agent && \
		GOSIGHT_SERVER_CONFIG=./configs/server.yaml go build ./server/cmd -o ./server/gosight

clean:
	@echo "Cleaning up..."
	-docker stop $(PG_CONTAINER)
	-docker stop $(VM_CONTAINER)
	rm -rf gosight victoria-metrics-data
