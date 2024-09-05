# accounts
simple banking service

# features
- accounts
- add money to account
- transfer money between accounts

# todo
- use money library to preserve precisions
- write tests for core functionality
- verify error messages and transaction usage

# considerations 
- Disable updates on transaction_lines table
- Every transaction results in a debit and credit
- Store lowest form of values (cents)
- Perform balance checks before transacting between accounts
- Deposits debit genesis account, whose balance represents total risk.

# setup
###### using docker (easier)
1. In the root of the project, run `docker compose up`
2. This should boot up a local postgresql instance & an instance of the accounts service.
3. Accounts service should be running on port 8080.

### barebones (doable)
1. Ensure you have a postgres db setup with the right dbname, user, password and sslmode.
2. Modify the Makefile in the project root, and set your `DB_URL` and `PORT`
3. To run, `make run`
3. To build `make build`


# interactions
```
curl --location 'localhost:6554/users' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "1@gmail.com"
}'

curl --location 'localhost:6554/accounts' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 5
}'

curl --location 'localhost:6554/transactions' \
--header 'Content-Type: application/json' \
--data '{
    "to": "985270462",
    "type": "deposit",
    "amount": 100,
    "reference": "ok"
}'

curl --location 'localhost:6554/transactions' \
--header 'Content-Type: application/json' \
--data '{
    "from": "810093581",
    "to": "985270462",
    "type": "transfer",
    "amount": 100,
    "reference": "lekkero"
}'

```