FROM ghcr.io/skip-mev/slinky-dev-base AS builder

WORKDIR /src/slinky

COPY go.mod .

RUN go mod download

COPY . .

RUN make build

FROM ubuntu:rolling
EXPOSE 8080 8002

COPY --from=builder /src/slinky/build/* /usr/local/bin/
RUN apt-get update && apt-get install jq -y \
    make \
    jq \
    curl \
    wget \
    tree \
    ca-certificates

WORKDIR /usr/local/bin/

RUN mkdir -p /oracle
RUN scripts --use-core=true --temp-file=/oracle/markets.json

CMD ["tail", "-f", "/dev/null"]
