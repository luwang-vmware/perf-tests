export GOVC_URL='Administrator@vsphere.local:Admin!23@20.20.0.6'
export GOVC_INSECURE=true
PREFIX=$1


VMS=$(govc find /tkg-dc -type m -name "${PREFIX}-*" | tr '\n' ' ')
for VM in ${VMS[@]}
do
      echo $VM
          govc vm.change  -vm ${VM} -sync-time-with-host=True
        done
