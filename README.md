# golang-easy-sso

> golang + postgres SQL

## Build Setup

``` bash

$ cd <Repo>

$ cp dot.env .env

$ vi .env

$ docker build -t golang-easy-sso ./

$ docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d postgres

$ export POSTGRES_CONTAINER_IP=$(docker inspect postgres | jq -r ".[].NetworkSettings.Networks.bridge.IPAddress")

$ docker run --rm -it -e POSTGRES_HOST="$POSTGRES_CONTAINER_IP" -p 8080:8080 --link postgres:db golang-easy-sso

```

## How to use

### Get token for super user

```bash

$ curl \
  -XPOST \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin", "password": "admin"}' \
  http://127.0.0.1:8080/api/v1/auth

```

### Add user

```bash

$ curl \
  -XPOST \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin", "password": "admin", "enabled": true}' \
  http://127.0.0.1:8080/api/v1/users

```

### List user

```bash

$ curl \
  -H 'Authorization: Bearer <TOKEN>' \
  http://127.0.0.1:8080/api/v1/users

```

### Show user detail

```bash

$ curl \
  -H 'Authorization: Bearer <TOKEN>' \
  http://127.0.0.1:8080/api/v1/users/<ID>

```

### Update user Information

```bash

$ curl \
  -XPUT \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin", "password": "admin", "enabled": false}' \
  http://127.0.0.1:8080/api/v1/users/<ID>

```

### Delete User

```bash

$ curl \
  -H 'Authorization: Bearer <TOKEN>' \
  -XDELETE http://127.0.0.1:8080/api/v1/users/<ID>

```

### Get JWT

```bash

$ curl \
  -XPOST \
  -H 'Content-Type: application/json' \
  -d '{"username": "admin", "password": "admin"}' \
  http://127.0.0.1:8080/api/v1/auth

```
