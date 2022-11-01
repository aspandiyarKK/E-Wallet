# Wallet-service

Для запуска сервиса

```shell
make up
```

Для запуска тестов

```shell
make test
```

Для запуска линтера

```shell
make lint
```

## Описание методов

### AddWallet (POST)

```bash
# AddDWallet
curl --location --request POST 'http://localhost:3000/api/v1/wallet' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--header 'Content-Type: application/json' \
--data-raw '{
    "owner":"Aspandiyar",
    "balance": 500  
}'
```

#### Response:

```
{
    "id": 1
}
```

### GetWallet (GET) для id = 1

```bash
# GetWallet для id = 1
params:

`?currency` - string(Examples:"USD","RUB","EUR",  ....), default:"RUB"

curl --location --request GET 'http://localhost:3000/api/v1/wallet/1' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--data-raw ''
```

#### Example Response:

```
{
    "owner": "Aspandiyar",
    "balance": 500,
    "created_at": "2022-10-25T19:12:18.705349+06:00",
    "updated_at": "2022-10-25T19:12:18.705186+06:00"
}
```

### UpdateWallet (PUT) for Id = 1

```bash
curl --location --request PUT 'http://localhost:3000/api/v1/wallet/1' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--header 'Content-Type: application/json' \
--data-raw '{
    "owner":"Aspandiyar_K",
    "balance":  3000.0
}'
```

#### Response:

```
{
    "owner": "Aspandiyar_K",
    "balance": 3000,
    "created_at": "2022-10-25T19:12:18.705349+06:00",
    "updated_at": "2022-10-25T19:37:25.900652+06:00"
}
``` 

### DepositWallet (PUT) for Id = 1

```bash
curl --location --request PUT 'http://localhost:3000/api/v1/wallet/2/deposit' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--header 'Content-Type: application/json' \
--data-raw '{"sum":5000,"uuid":"f7eb5a3b-d9d2-11ec-abbd-0242ac150004"}'
```

#### Response:

```
{
"Ok"
}
```

### Withdraw (PUT) for Id = 1

```bash
curl --location --request PUT 'http://localhost:3000/api/v1/wallet/2/withdraw' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--header 'Content-Type: application/json' \
--data-raw '{
    "sum": 5000,
    "uuid": "f7eb5a3b-d9d2-11ec-abbd-0242ac156004"
}'
```

#### Response:

```
{
"Ok"
}
```

### TransferMoney (PUT) from Id = 1 to Id = 2

```bash
curl --location --request PUT 'http://localhost:3000/api/v1/wallet/1/transfer' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--header 'Content-Type: application/json' \
--data-raw '{
    "sum": 200,
    "walletTarget": 2,
    "uuid": "f7eb5a3b-d9d2-11ec-abed-0242ac160004"
}'
```

#### Response:

```
{
    "Success transferring"
}
```

### GetTransactions (GET)

params:

`?limit` - int

`?offset` - int, default: 100

`?desc` - "true"/"false", default:"true"

`?sort` - "amount"/"date", default:"date"

```bash
curl --location --request GET 'http://localhost:3000/api/v1/wallet/2/transactions?sort=sum&desc=false&limit=2' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--data-raw ''
```

#### Example Response:

```
[
    {
        "transaction_id": 7,
        "uuid": "f7eb5a3b-d9d2-11ec-abed-0242ac130004",
        "from_id": 2,
        "to_id": null,
        "sum": 20,
        "operation": "withdraw",
        "date": "2022-10-25T22:21:50.773669+06:00"
    },
    {
        "transaction_id": 4,
        "uuid": "f7eb5a3b-d9d2-11ec-abed-0242ac160004",
        "from_id": 2,
        "to_id": 3,
        "sum": 200,
        "operation": "transfer",
        "date": "2022-10-25T22:09:44.153883+06:00"
    }
]
```

### DeleteWallet (DELETE) for Id = 1

```bash
curl --location --request DELETE 'http://localhost:3000/api/v1/wallet/1' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFzcGFuIiwiZXhwIjoxNjcwMzA4Nzg4LCJpc3MiOiJlLXdhbGxldCJ9.2acWtWXecZ4L0hu2jAhUJnRcyPTlUjsDOWU_v7NeYPA' \
--data-raw ''
```

#### Response:

```
{}
```







