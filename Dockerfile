FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY ./bin/watcher  .
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
CMD ["/app/watcher", "hosts"]
