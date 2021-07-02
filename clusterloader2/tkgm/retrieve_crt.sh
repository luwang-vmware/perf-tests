set -e                                                                          
                                                                                
MASTER_IP=$1                                                                    
ssh   -o StrictHostKeyChecking=no capv@${MASTER_IP} "sudo cp /etc/kubernetes/pki/etcd/server.* .; sudo chmod 644 server.*;"
ssh   -o StrictHostKeyChecking=no capv@${MASTER_IP} "sudo cp /etc/kubernetes/pki/etcd/ca.* .; sudo chmod 644 ca.*;"
scp   -o StrictHostKeyChecking=no capv@${MASTER_IP}:~/server.* .                
scp   -o StrictHostKeyChecking=no capv@${MASTER_IP}:~/ca.* .                    
                                                                                  
kubectl create ns monitoring                                                        
kubectl create secret generic kube-etcd-client-certs --from-file=server.crt=server.crt --from-file=server.key=server.key -n monitoring
