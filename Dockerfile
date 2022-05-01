FROM golang:alpine3.15 as first

WORKDIR /app


COPY ./go/go.mod ./  
COPY ./go/go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /demo

EXPOSE 8082

CMD [ "/demonstration" ]


FROM alpine:3.15.4

WORKDIR /
RUN addgroup -S demogroup && adduser -S demouser -G demogroup
COPY --from=first /demo /demo
EXPOSE 8082

USER demouser

ENTRYPOINT ["/demo"]
