# Method to run clusterloader2 on clusters deployed by tanzu in vsphere platform



## Deploy a workload cluster by tanzu

## Sync the code
1. git clone https://github.com/luwang-vmware/perf-tests.git
2. git checkout -b tkgm_release-1.20 origin/tkgm_release-1.20
2. make softlink ``ln -s <dest_folder_of_perf_test_source_code> /src/k8s.io/perf-test``

## Sync the time of cluster with the hosting ESXi
1. check the set_time_align_esx.sh and update the vc ip and datacenter per your test env
2. ``bash set_time_align_esx.sh <your cluster name>``

## Retrieve etcd related crt from one of the controlplane
1. ``bash retrieve_crt.sh <one of controlplane ip>``

## Update the env var and export them
1. edit export_env.rc with the correct controlplane ip address
2. ``source export_env.rc``

## Run clusterloader2
1. cd /src/k8s.io/perf-test; ./run-e2e.sh   --testconfig=./testing/node-throughput/config.yaml  --report-dir=/tmp/1  --masterip=$masterip --master-internal-ip=$masterip --enable-prometheus-server=true --tear-down-prometheus-server=false --prometheus-scrape-etcd=true --prometheus-scrape-kube-proxy=true  --prometheus-scrape-node-exporter=false  --prometheus-manifest-path /src/k8s.io//perf-test/clusterloader2/pkg/prometheus/manifests/ --alsologtostderr --provider vsphere  2>&1