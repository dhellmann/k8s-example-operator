#!/bin/sh -x

# Run the operator outside of the cluster

export OPERATOR_NAME=k8s-example-operator
operator-sdk up local --namespace=myproject
