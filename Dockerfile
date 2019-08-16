FROM daixijun1990/scratch
LABEL maintainer="Xijun Dai <daixijun1990@gmail.com>"
COPY drone-oss-cache /bin/
ENTRYPOINT ["/bin/drone-oss-cache"]
