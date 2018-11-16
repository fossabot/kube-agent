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

skip="${1}"

if [[ -n "${skip}" ]]; then
    echo "Skipping deployment creation"
else
    if kubectl get deploy kube-agent -n wodby &> /dev/null; then
        kubectl delete deploy kube-agent -n wodby
        kubectl create -f deployment.yml
        exit 0
    fi
fi

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

# random uuid.
uuid="7ed990cc-fafe-4741-85a1-55d3ba03d8d4"
token="random-token"
name="kube_agent_${uuid}"

rabbitadm declare queue auto_delete=false durable=true arguments="${queue_args}" name="${name}_q"
rabbitadm declare binding source="amq.direct" destination="${name}_q" routing_key="${name}_q"

rabbitadm declare queue auto_delete=false durable=true arguments="${queue_args}" name="kube_agent_q"

rabbitadm declare exchange name="kube_agent.dx" type=direct durable=true auto_delete=false
rabbitadm declare binding source="kube_agent.dx" destination="kube_agent_q" routing_key="worker1"

rabbitadm declare user name="${name}" password="${token}" tags="kube_agent"
rabbitadm declare permission user="${name}" vhost=/ configure="^$" write="^kube_agent\.dx$" read="${name}_q"

if [[ -z "${skip}" ]]; then
    if ! kubectl get ns wodby &> /dev/null; then
        kubectl create ns wodby
        kubectl create rolebinding kube-agent --serviceaccount=wodby:kube-agent --clusterrole=cluster-admin
    fi

    kubectl create -f deployment.yml
fi

#rabbitadm publish routing_key="${name}_q" exchange="amq.direct" \
#    properties='{"content_type":"text/json", "type": "kube_api_request", "reply-to": "worker1", "delivery-mode": 2, "corellation-id": 123, "expiration": "150"}' \
#    payload='{"method":"GET","uri":"api\/v1\/namespaces","body":null}'

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"api\/v1\/namespaces\/uuid\/secrets","body":null}}

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"api\/v1\/namespaces\/uuid\/services","body":null}}

#{"type":"action","action":"kubernetes_api_call","params":{"method":"GET","uri":"apis\/extensions\/v1beta1\/namespaces\/uuid\/deployments","body":null}}

#{"type":"action","action":"is_ok","params":[]}
