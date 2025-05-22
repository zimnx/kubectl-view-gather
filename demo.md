1. Show a sample must-gather archive structure

    ```
    tree ~/Downloads/must-gather-sample/
    ```

2. List all namespaces

    ```
    kubectl view gather -p ~/Downloads/must-gather-sample/ -- get namespace
    ```
   
3. List pods in all namespaces

    ```
    kubectl view gather -p ~/Downloads/must-gather-sample/ -- get pods -A
    ```

4. List pods in a specific namespace

    ```
    kubectl view gather -p ~/Downloads/must-gather-sample/ -- get pods -n scylla-operator
    ```


5. Get a Scylla cluster with a specific name

    ```
    kubectl view gather -p ~/Downloads/must-gather-sample/ -- get scyllacluster -n scylla scylla
    ```
   
6. Get a Scylla cluster represented as yaml

    ```
    kubectl view gather -p ~/Downloads/must-gather-sample/ -- get scyllacluster -n scyla scylla -o yaml | less
    ```
