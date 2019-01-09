#!/bin/bash -x

cd ~/go/src/github.com/dhellmann/k8s-example-operator
oc --as system:admin apply -f deploy/service_account.yaml
oc --as system:admin apply -f deploy/role.yaml
oc --as system:admin apply -f deploy/role_binding.yaml
oc --as system:admin apply -f deploy/crds/app_v1alpha1_appservice_crd.yaml

oc config use-context minishift
