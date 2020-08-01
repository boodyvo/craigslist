FROM golang:1.13-alpine

RUN mkdir /go/src/github.com/
RUN mkdir /go/src/github.com/boodyvo/
RUN mkdir /go/src/github.com/boodyvo/craigslist
WORKDIR /go/src/github.com/boodyvo/craigslist
COPY . .

RUN go mod tidy && go mod vendor

CMD ["go", "run", "main.go"]