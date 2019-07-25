# TODO: I would like to move this to alpine but it'll need some more effort
# this is "good enough" for now
FROM golang:latest

LABEL description="Lollipops command-line tool to generate variant annotation diagrams"
LABEL url="https://github.com/pbnjay/lollipops"
LABEL maintainer="Jeremy Jay <jeremy@pbnjay.com>"

## add the arial font (debian)
RUN sed -i 's/main$/main contrib non-free/' /etc/apt/sources.list
RUN apt-get update && apt-get install -y ttf-mscorefonts-installer

## add the arial font (alpine)
#RUN apk --no-cache add msttcorefonts-installer fontconfig && \
#    update-ms-fonts && \
#    fc-cache -f

WORKDIR /go/src/github.com/pbnjay/lollipops
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["/go/bin/lollipops"]
CMD ["-legend", "-labels", "TP53", "R248Q#7f3333@131", "R273C", "R175H", "T125@5"]
