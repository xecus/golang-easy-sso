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

