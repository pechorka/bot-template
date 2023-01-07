FROM umputun/baseimage:buildgo-latest as build

ADD . /build
WORKDIR  /build
RUN go test ./app/...
RUN cd app && go build -o /target/bot-template


FROM alpine:3.17
ENV \
    APP_USER=app               \
    APP_UID=1001
COPY --from=build /target/bot-template /srv/bot-template

RUN adduser -s /bin/sh -D -u $APP_UID $APP_USER && chown -R $APP_USER:$APP_USER /home/$APP_USER && \
    mkdir -p /srv && chown -R $APP_USER:$APP_USER /srv && \
    rm -rf /var/cache/apk/*
RUN chown -R app:app /srv

CMD ["/srv/bot-template"]