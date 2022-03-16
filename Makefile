default: build

.PHONY: build
build: backend frontend

.PHONY: backend
backend:
	make -C ./backend build

.PHONY: frontend
frontend:
	make -C ./frontend build

.PHONY: up
up:
	docker-compose up --build -d

.PHONY: down
down:
	docker-compose down

.PHONY: restart
restart:
	docker-compose up --build -d
