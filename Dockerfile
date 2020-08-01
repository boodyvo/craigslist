FROM golang:1.13-alpine

RUN mkdir -p /go/src/github.com/boodyvo/craigslist
WORKDIR /go/src/github.com/boodyvo/craigslist
COPY . .
# BUILD DEPS
RUN go mod tidy && go mod vendor

CMD ["go", "run", "main.go"]