# accounts


# setup
### Docker (easier)
1. In the root of the project, run `docker compose up`
2. This should boot up a local postgresql instance & an instance of the accounts service.
3. Accounts service should be running on port 8080.

### Barebones
1. Ensure you have a postgres db setup with the right dbname, user, password and sslmode setup.
2. Run `DB_URL=<your_db_url> PORT=<your_port> go run cmd/accounts/main.go`