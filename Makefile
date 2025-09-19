# To create and manage a local PostgreSQL database using Docker
postgres:
	docker run --name postgres18rc1 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18rc1-alpine

# To Create a new database named simple_bank
createdb: 
	docker exec -it postgres18rc1 createdb --username=root --owner=root simple_bank

# To Drop the database named simple_bank
dropdb: 
	docker exec -it postgres18rc1 dropdb simple_bank

.PHONY: postgres createdb dropdb