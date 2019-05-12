FROM alpine:3.8

RUN set -x && \
	apk update && \
	apk add curl

COPY fs /

EXPOSE 3000:3000

CMD ["/healthy-grpc"]

