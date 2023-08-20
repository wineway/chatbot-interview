FROM golang:1.19

COPY . .

RUN unset GOPATH && cd cmd/chatbot && go build

FROM debian:12

RUN apt-get -qq update \
    && apt-get -qq install -y --no-install-recommends ca-certificates
copy --from=0 /go/cmd/chatbot/chatbot ./chatbot
CMD ["./chatbot"]

