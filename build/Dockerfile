FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/k8s-example-operator /usr/local/bin/k8s-example-operator
