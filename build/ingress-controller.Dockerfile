FROM golang:1.19 as builder

WORKDIR /build

COPY . .

RUN make build_ingress-controller  \
    && mkdir /output \
    && mv ./ingress-controller /output

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /output .

RUN apk add --no-cache libc6-compat \
    && rm -rf /tmp/* \
    && rm -rf /var/cache/apk/*

ENTRYPOINT ["./ingress-controller"]
