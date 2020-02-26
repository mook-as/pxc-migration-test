FROM golang AS builder

WORKDIR /go/src/app
COPY . .
RUN go install .

FROM scratch
COPY --from=builder /go/bin/pxc-migration-test /
ENTRYPOINT [/pxc-migration-test]