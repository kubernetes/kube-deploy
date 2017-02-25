FROM alpine:3.5

RUN apk add --update protobuf protobuf-dev openjdk8 maven tzdata
RUN cp /usr/share/zoneinfo/Europe/Brussels /etc/localtime
RUN echo "Asia/Hong_Kong" >  /etc/timezone
CMD ["sh"]

