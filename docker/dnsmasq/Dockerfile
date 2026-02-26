FROM alpine:latest
RUN apk add --no-cache dnsmasq
EXPOSE 53/tcp 53/udp
ENTRYPOINT ["dnsmasq", "-k"]
