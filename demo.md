1. Show a sample must-gather archive structure

    ```
    tree ~/Downloads/must-gather/
    ```

1. List api-resources

   ```
   kubectl view gather -p ~/Downloads/must-gather/ -- api-resources
   ```

1. List all namespaces

    ```
    kubectl view gather -p ~/Downloads/must-gather/ -- get namespace
    ```
   
1. List pods in all namespaces

    ```
    kubectl view gather -p ~/Downloads/must-gather/ -- get pods -A
    ```

1. List pods in a specific namespace

    ```
    kubectl view gather -p ~/Downloads/must-gather/ -- get pods -n scylla-operator
    ```

1. Get pod logs

   ```
   kubectl view gather -p ~/Downloads/must-gather/ -- logs -n scylla-operator scylla-operator-cc9b58cf5-48nl8
   ```

1. List all scylla clusters

    ```
   kubectl view gather -p ~/Downloads/must-gather/ -- get scyllaclusters -A
    ```

1. Get a Scylla cluster with a specific name

    ```
    kubectl view gather -p ~/Downloads/must-gather/ -- get scyllacluster -n ims ims-scylla-stage
    ```
   
1. Get a Scylla cluster represented as yaml

    ```
    kubectl view gather -p ~/Downloads/must-gather/ -- get scyllacluster -n ims ims-scylla-stage -o yaml
    ```
