FROM alpine

COPY bin/service-registrator-linux-amd64 /home/

WORKDIR /home
RUN chmod +x service-registrator-linux-amd64

ENTRYPOINT ["./service-registrator-linux-amd64"]