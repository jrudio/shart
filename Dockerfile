FROM golang:alpine as builder

RUN apk update && apk add git && apk add ca-certificates

COPY . $GOPATH/src/github.com/jrudio/shart
# RUN go get -u github.com/jrudio/shart
WORKDIR $GOPATH/src/github.com/jrudio/shart

RUN go get -u github.com/golang/dep/cmd/dep

RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -X main.version=$(git describe --always --long --dirty)" -o /go/bin/shart

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /go/bin/shart /go/bin/shart

EXPOSE 6969

ENTRYPOINT ["/go/bin/shart"]
