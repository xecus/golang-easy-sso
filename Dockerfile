FROM golang

RUN mkdir -p /go/src/app

WORKDIR /go/src/app
COPY . /go/src/app

#RUN curl https://glide.sh/get | sh
#RUN glide install -v

RUN go get github.com/StephanDollberg/go-json-rest-middleware-jwt
RUN go get github.com/jinzhu/gorm
RUN go get github.com/lib/pq
RUN go get github.com/joho/godotenv
RUN cp dot.env .env

EXPOSE 8080
CMD go run main.go
