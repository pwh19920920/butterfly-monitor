FROM golang:1.17.2 as builder
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o spider


FROM alpine
ENV TIME_ZONE=Asia/Shanghai
COPY --from=builder /usr/share/zoneinfo/$TIME_ZONE /etc/localtime
RUN echo $TIME_ZONE > /etc/timezone

WORKDIR /dist
COPY --from=builder /app/spider .
COPY --from=builder /app/conf ./conf
EXPOSE 8081

ARG PARAMS
CMD /dist/spider $PARAMS