FROM wodby/alpine:3.8-2.1.1

COPY ./bin/linux-amd64/kube-agent /usr/local/bin/kube-agent

CMD [ "kube-agent" ]
