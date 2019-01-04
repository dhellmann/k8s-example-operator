# Example Operator

## Setup with minishift

1. Install and launch minishift

  https://docs.okd.io/latest/minishift/getting-started/index.html

2. Install operator-sdk

    ```
    go get github.com/dhellmann/k8s-example-operator
    cd ~/go/src/github.com/dhellmann/k8s-example-operator
    oc --as system:admin create -f deploy/service_account.yaml
    oc --as system:admin create -f deploy/role.yaml
    oc --as system:admin create -f deploy/role_binding.yaml
    oc --as system:admin create -f deploy/crds/app_v1alpha1_appservice_crd.yaml
    ```

3. Ensure you're logged in to the correct context

    ```
    oc config use-context minishift
    ```

4. Launch the operator locally

    ```
    export OPERATOR_NAME=k8s-example-operator
    operator-sdk up local --namespace=myproject
    ```

5. Create the CR

    ```
    oc apply -f deploy/crds/app_v1alpha1_appservice_cr.yaml
    ```
