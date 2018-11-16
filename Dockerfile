FROM wodby/alpine:3.8-2.1.1

COPY ./bin/linux-amd64/kube-agent /usr/local/bin/kube-agent
COPY ./vendor/github.com/streadway/amqp/LICENSE /usr/share/doc/github.com/streadway/amqp/
COPY ./vendor/k8s.io/client-go/LICENSE /usr/share/doc/k8s.io/client-go/

CMD [ "kube-agent" ]
