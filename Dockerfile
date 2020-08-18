FROM golang:1.13-alpine3.12 as builder
RUN mkdir -p /go/src/gobins
COPY . /go/src/gobins
RUN cd /go/src/gobins \
  && go mod verify \
  && go test -v ./cmd/parser/... \ 
  && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/ -a ./cmd/... 

FROM scratch
COPY --from=builder /go/bin/cluster_reader /go/bin/cluster_reader
COPY --from=builder /go/bin/parser /go/bin/parser
COPY --from=builder /go/src/gobins/cmd/parser/templates /templates
