This is a reproducer for a kubelet issue where it unmounts a volume of a
running pod due to connection issues to the API server. As the issue is only
reproducible in very specific scenarios, the `main.go` contains a proxy that
can be put between kubelet and the API server to simulate the issue reliably
by always failing API calls to `/api/v1/persistentvolumes/pvc-`.

1. Create a kind cluster (`v0.20.0` with the node image `1.27.3` was used in my case)

    ```bash
    kind create cluster
    ```

1. Install a CSI driver like [`csi-driver-host-path`](https://github.com/kubernetes-csi/csi-driver-host-path/blob/master/docs/deploy-1.17-and-later.md)

    ```bash
    kubectl apply -f csi/
    ```

1. Create a PVC and a Pod that mounts it

    ```bash
    kubectl apply -f sts.yaml
    ```

1. Build, copy and run the `pvcfailproxy`

    ```bash
    GOOS=linux go build -o pvcfailproxy
    docker cp pvcfailproxy kind-control-plane:/usr/bin/
    docker exec -ti kind-control-plane pvcfailproxy
    ```

1. Modify `/etc/kubernetes/kubelet.conf` to use the `pvcfailproxy` and restart kubelet

    ```bash
    # in another terminal
    docker exec -ti kind-control-plane sed -i 's/kind-control-plane:6443/kind-control-plane:8443/g' /etc/kubernetes/kubelet.conf
    docker exec -ti kind-control-plane systemctl restart kubelet
    ```

1. Observe the kubelet log if it contains the expected error and unmount operation

    ```bash
    docker exec -ti kind-control-plane journalctl -fu kubelet
    ```

    ```bash
    desired_state_of_world_populator.go:295] "Error processing volume" err="error processing PVC default/data-web-0: failed to fetch PV pvc-8308a5fc-c8b2-42c2-8f1c-c0183ea2bce5 from API server ...
    reconciler_common.go:172] "operationExecutor.UnmountVolume started for volume \"pvc-8308a5fc-c8b2-42c2-8f1c-c0183ea2bce5\"
    ```
