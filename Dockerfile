FROM umputun/baseimage:buildgo-latest as build

ADD . /build
WORKDIR  /build
RUN go test ./app/...
RUN cd app && CGO_ENABLED=0 GOOS=linux go build -o /target/bot-template -ldflags \
    "-X main.revision=$(git rev-parse --abbrev-ref HEAD)-$(git describe --abbrev=7 --always --tags)-$(date +%Y%m%d-%H:%M:%S)"


FROM umputun/baseimage:app
COPY --from=build /target/bot-template /srv/bot-template

RUN chown -R app:app /srv

CMD ["/srv/bot-template"]