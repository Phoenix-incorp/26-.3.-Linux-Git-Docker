FROM golang
RUN mkdir -p /go/src/website
WORKDIR /go/src/website
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Ivan Ivanov<test@test.ru>"
WORKDIR /root/
COPY --from=0 /go/bin/website .
ENTRYPOINT ./website
EXPOSE 8080
