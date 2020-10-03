FROM golang:1.14

RUN mkdir -p /go/src/github.com/boodyvo/craigslist
WORKDIR /go/src/github.com/boodyvo/craigslist
COPY . .
# BUILD DEPS
RUN go mod download
#WORKDIR /go/src/github.com/boodyvo/craigslist/scrapper

CMD ["go", "run", "main.go"]
#CMD ["go", "test", "-v"]