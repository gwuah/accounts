# accounts
simple banking service 

# features
- accounts
- add money to account
- transfer money between accounts

# considerations 
- Disable updates & deletes on transaction_lines table
- Every transaction results in a debit and credit
- We store lowest form of values (cents)
- We perform balance checks before transacting between accounts
- Deposits debit genesis account, whose balance represents total risk
- We assume currency USD with 2 decimal places of precision
- We use transaction references to prevent duplicate transactions (idempotency key)

# improvements
- async processing of transactions
- support for multiple currencies
- robust user authn & authz

# setup
#### using docker (easier)
1. In the root of the project, run `docker compose up`
2. This should boot up a local postgresql instance & an instance of the accounts service.
3. Accounts service should be running on port 8080.

#### barebones (doable)
1. Ensure you have a postgres db setup with the right dbname, user, password and sslmode.
2. Modify the Makefile in the project root, and set your `DB_URL` and `PORT`
3. To run, `make run`
4. To build `make build`
5. To test `make test`


# interactions
```
curl --location 'localhost:8080/users' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "1@gmail.com"
}'

curl --location 'localhost:8080/accounts' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 5
}'

curl --location 'localhost:8080/transactions' \
--header 'Content-Type: application/json' \
--data '{
    "to": "985270462",
    "type": "deposit",
    "amount": 100,
    "reference": "ok"
}'

curl --location 'localhost:8080/transactions' \
--header 'Content-Type: application/json' \
--data '{
    "from": "810093581",
    "to": "985270462",
    "type": "transfer",
    "amount": 100,
    "reference": "lekkero"
}'

curl --location 'localhost:8080/accounts/715733003'
```

# notes
This is an improvement on [cashapp](https://github.com/gwuah/cashapp), which I wrote 4 years ago. Code is much cleaner to follow & easier to extend. It also has quite good test coverage.