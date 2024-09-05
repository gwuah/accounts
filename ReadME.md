# accounts
A simple banking service

# features
- account creation
- add money to account
- transfer money between accounts

# setup
### Docker (easier)
1. In the root of the project, run `docker compose up`
2. This should boot up a local postgresql instance & an instance of the accounts service.
3. Accounts service should be running on port 8080.

### Barebones
1. Ensure you have a postgres db setup with the right dbname, user, password and sslmode setup.
2. Modify the Makefile in the project root, and set your `DB_URL` and `PORT`
3. To run, `make run`
3. To build `make build`