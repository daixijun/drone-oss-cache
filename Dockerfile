FROM golang:1.12-alpine as buidler
ENV CGO_ENABLE=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
WORKDIR /workspace
COPY . .
RUN go build -ldflags "-w -s" -o drone-oss-cache .

FROM alpine:3.9
LABEL maintainer="Xijun Dai <daixijun1990@gmail.com>"
RUN apk --no-cache add ca-certificates
COPY --from=buidler /workspace/drone-oss-cache /bin/
ENTRYPOINT ["/bin/drone-oss-cache"]
