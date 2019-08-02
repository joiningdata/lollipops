FROM golang:alpine AS builder

RUN apk update && \
    apk add --no-cache git ca-certificates msttcorefonts-installer tzdata && \
    update-ca-certificates && \
    update-ms-fonts && \
    adduser -D -g '' lollipops

WORKDIR /go/src/github.com/pbnjay/lollipops
COPY . .

RUN go get -d -v
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/lollipops


########################################################################
FROM scratch

LABEL description="Lollipops command-line tool to generate variant annotation diagrams"
LABEL url="https://github.com/pbnjay/lollipops"
LABEL maintainer="Jeremy Jay <jeremy@pbnjay.com>"

# Pull in a number of files from builder image
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/fonts/truetype/msttcorefonts/Arial.ttf /usr/share/fonts/truetype/msttcorefonts/arial.ttf
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /go/bin/lollipops /bin/lollipops

USER lollipops

WORKDIR /data
VOLUME /data
CMD ["/bin/lollipops", "-legend", "-labels", "TP53", "R248Q#7f3333@131", "R273C", "R175H", "T125@5"]
