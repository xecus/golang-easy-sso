# golang-easy-sso

> golang + postgres SQL

## Build Setup

``` bash

$ cd <Repo>

$ docker build -t golang-easy-sso ./

$ docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d postgres

$ docker run --rm -it -p 8080:8080 --link postgres golang-easy-sso

```

