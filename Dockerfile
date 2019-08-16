FROM alpine:3.9
LABEL maintainer="Xijun Dai <daixijun1990@gmail.com>"
RUN apk --no-cache add tzdata ca-certificates
COPY drone-oss-cache /bin/
ENTRYPOINT ["/bin/drone-oss-cache"]
