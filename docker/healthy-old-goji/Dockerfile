FROM alpine:3.8

RUN set -x && \
	apk update && \
	apk add curl

COPY fs /

EXPOSE 3000:8000

CMD ["/healthy-old-goji"]

