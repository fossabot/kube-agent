#!/usr/bin/env bash

set -e

if [[ -n "${DEBUG}" ]]; then
    set -x
fi

rabbit() {
    docker-compose exec rabbitmq "${@}"
}

rabbitadm() {
    rabbit rabbitmqadmin -utest -ptest "${@}"
}

if [[ ! -f certificate.pem ]]; then
    openssl req -subj '/C=US/ST=""/L=""/O=""/OU=""/CN=""/emailAddress=""' \
        -newkey rsa:2048 -nodes -keyout key.pem -x509 -days 365 -out certificate.pem
fi

docker-compose up -d rabbitmq

for i in $(seq 1 10); do
    if rabbit rabbitmqctl node_health_check &>/dev/null; then
        started=1
        break
    fi
    echo "RabbitMQ is starting..."
    sleep 2
done

if [[ -z "${started}" ]]; then
    echo >&2 "Error. Failed to start RabbitMQ."
    exit 1
fi

echo "RabbitMQ has started"

docker-compose up -d

echo "Preparing rabbitmq"

queue_args='{"x-expires":86400000,"x-max-length-bytes":1048576}'

rabbitadm declare queue auto_delete=false durable=true arguments="${queue_args}" name="cluster_test"
rabbitadm declare binding source="amq.direct" destination="cluster_test" routing_key="cluster_test"

rabbitadm declare queue auto_delete=false durable=true arguments="${queue_args}" name="nodes"
rabbitadm declare exchange name="wodby.nodes" type=direct durable=true auto_delete=false
rabbitadm declare binding source="wodby.nodes" destination="nodes" routing_key="nodes"

rabbitadm declare user name="cluster_test" password="token" tags="node"
rabbitadm declare permission user="cluster_test" vhost=/ configure="^$" write="^wodby\.nodes$" read="^cluster_test$"

if ! kubectl get ns wodby &> /dev/null; then
    kubectl create ns wodby
fi

#if kubectl get deploy agent -n wodby &> /dev/null; then
#    kubectl delete deploy agent -n wodby
#fi

kubectl create -f deployment.yml

#rabbitadm publish routing_key="cluster_test" exchange="amq.direct" \
#    properties='{"content_type":"text/json"}' \
#    payload='{"type":"action","action":"is_ok","params":{},"context":{"message_uuid":"1","reply_to":"worker1"}}'

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"api\/v1\/namespaces","body":null},"context":{"message_uuid":"0468bfd4-da50-449a-a7e5-72b2a1aec52a"}}

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"api\/v1\/namespaces\/9122ff42-d53f-40d6-ab46-07de0f0d252f\/secrets","body":null},"context":{"message_uuid":"a2aa2229-b5f1-4eb4-8b0e-6c8df6dc591d"}}

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"api\/v1\/namespaces\/9122ff42-d53f-40d6-ab46-07de0f0d252f\/services","body":null},"context":{"message_uuid":"14e55369-c686-438e-8645-944ab39d4650"}}

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"apis\/extensions\/v1beta1\/namespaces\/9122ff42-d53f-40d6-ab46-07de0f0d252f\/deployments","body":null},"context":{"message_uuid":"b744d5bb-6258-4b66-8e84-14ac6b0616fa"}}

#{"type":"action","action":"is_ok","params":[],"context":{"message_uuid":"3c71f70a-a60e-4d32-a536-8e77f850a3f2"}}
