# Multinode addons; all singlenode addons plus kube-proxy (and soon flannel)
	
cd ${TEMP_DIR} && sed -i "/^[ |\t]*{[#|%]/d" /addons/*.yaml 

cd ${TEMP_DIR} && sed -i -e "s@{{ *base_metrics_memory *}}@140Mi@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *metrics_memory *}}@300Mi@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *base_metrics_cpu *}}@80m@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *metrics_cpu *}}@100m@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *base_eventer_memory *}}@190Mi@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *eventer_memory *}}@206800Ki@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *metrics_memory_per_node *}}@4@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *eventer_memory_per_node *}}@500@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *nanny_memory *}}@94160Ki@g" /addons/heapster-cont*.yaml 
cd ${TEMP_DIR} && sed -i -e "s@{{ *metrics_cpu_per_node *}}@0.5@g" /addons/heapster-cont*.yaml 
