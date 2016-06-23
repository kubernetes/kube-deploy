function(cfg1)
	local cfg = cfg1.phase1;
	local master_private_ip = cfg.azure.master_private_ip;
	local names = {
		resource_group: "%(instance_prefix)s" % cfg,
		master_public_ip: "%(instance_prefix)s-master-pip" % cfg,
		availability_set: "%(instance_prefix)s-as" % cfg,
		storage_account: "${replace(\"%(instance_prefix)sstrg\", \"-\", \"\")}" % cfg,
		storage_container: "kube0000%(instance_prefix)s" % cfg,
		vnet: "%(instance_prefix)s-vnet" % cfg,
		subnet: "%(instance_prefix)s-subnet" % cfg,
		route_table: "%(instance_prefix)s" % cfg,
		security_group: "%(instance_prefix)s-nsg" % cfg,
		master_nic: "%(instance_prefix)s-master-nic" % cfg,
		master_vm: "%(instance_prefix)s-master" % cfg,
		node_nic: "%(instance_prefix)s-node-nic" % cfg,
		node_vm: "%(instance_prefix)s-node" % cfg,
	};
	{
		variable: {
			subscription_id: { default: cfg.azure.subscription_id },
			tenant_id: { default: cfg.azure.tenant_id },
			client_id: { default: cfg.azure.client_id },
			client_secret: { default: cfg.azure.client_secret },
		},
		provider: {
			azurerm: {
				subscription_id: "${var.subscription_id}",
				tenant_id: "${var.tenant_id}",
				client_id: "${var.client_id}",
				client_secret: "${var.client_secret}",
			}
		},
		resource: {
			azurerm_resource_group: {
				rg: {
					name: names.resource_group,
					location: cfg.azure.location,
				}
			},
			azurerm_storage_account: {
				sa: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					name: names.storage_account,
					location: "${azurerm_resource_group.rg.location}",
					account_type: "Standard_LRS"
				}
			},
			azurerm_storage_container: {
				sc: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					storage_account_name: "${azurerm_storage_account.sa.name}",
					name: names.storage_container,
					container_access_type: "private"
				}
			},
			azurerm_availability_set: {
				as: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					name: names.availability_set,
					location: "${azurerm_resource_group.rg.location}"
				}
			},
			azurerm_virtual_network: {
				vnet: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: "${azurerm_resource_group.rg.location}",
					name: names.vnet,
					address_space: ["10.0.0.0/8"],
				}
			},
			azurerm_route_table: {
				rt: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: "${azurerm_resource_group.rg.location}",
					name: names.route_table,
				}
			},
			azurerm_subnet: {
				subnet: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					name: names.subnet,
					virtual_network_name: "${azurerm_virtual_network.vnet.name}",
					address_prefix: "10.0.0.0/16",
					network_security_group_id: "${azurerm_network_security_group.sg.id}",
					route_table_id: "${azurerm_route_table.rt.id}",
				}
			},
			azurerm_network_security_group: {
				sg: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: cfg.azure.location,
					name: names.security_group,
				},
			},
			azurerm_network_security_rule: {
				[cfg.instance_prefix+"-master-ssh"]: {
						name: "%(instance_prefix)s-master-ssh" % cfg,
						priority: 100,
						direction: "Inbound",
						access: "Allow",
						protocol: "Tcp",
						source_port_range: "*",
						destination_port_range: "22",
						source_address_prefix: "*",
						destination_address_prefix: "*",
						resource_group_name: "${azurerm_resource_group.rg.name}",
						network_security_group_name: "${azurerm_network_security_group.sg.name}",
				},
				[cfg.instance_prefix+"-master-ssl"]: {
						name: "%(instance_prefix)s-master-ssl" % cfg,
						priority: 110,
						direction: "Inbound",
						access: "Allow",
						protocol: "Tcp",
						source_port_range: "*",
						destination_port_range: "443",
						source_address_prefix: "*",
						destination_address_prefix: "*",
						resource_group_name: "${azurerm_resource_group.rg.name}",
						network_security_group_name: "${azurerm_network_security_group.sg.name}",
				},
			},
			azurerm_public_ip: {
				pip: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: cfg.azure.location,
					name: names.master_public_ip,
					public_ip_address_allocation: "static",
					provisioner:  [{
						"local-exec": {
							command: |||
								cat <<EOF > ../../phase1b/crypto/san-extras
								DNS.1 = kubernetes
								DNS.2 = kubernetes.default
								DNS.3 = kubernetes.default.svc
								DNS.4 = kubernetes.default.svc.cluster.local
								DNS.5 = names.master_vm
								IP.1 = ${azurerm_public_ip.pip.ip_address}
								IP.2 = 10.3.0.1
								IP.3 = %s
								EOF
							||| % master_private_ip
						}
					}]
				}
			},
			azurerm_network_interface: {
				master_nic: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: "${azurerm_resource_group.rg.location}",
					name: names.master_nic,
					ip_configuration: {
						name: "ipconfig",
						subnet_id: "${azurerm_subnet.subnet.id}",
						private_ip_address_allocation: "static",
						private_ip_address: master_private_ip,
						public_ip_address_id: "${azurerm_public_ip.pip.id}",
					},
					enable_ip_forwarding: true,
				},
				node_nic: {
					resource_group_name: "${azurerm_resource_group.rg.name}",
					location: "${azurerm_resource_group.rg.location}",
					name: names.node_nic+"-${count.index}",
					ip_configuration: {
						name: "ipconfig",
						subnet_id: "${azurerm_subnet.subnet.id}",
						private_ip_address_allocation: "Dynamic",
					},
					enable_ip_forwarding: true,
					count: cfg.num_nodes
				}
			},
			template_file: {
				azure_json: {
					template: "${file(\"azure.json\")}",
					vars: {
						tenantId: "${var.tenant_id}",
						subscriptionId: "${var.subscription_id}",
						adClientId: "${var.client_id}",
						adClientSecret: "${var.client_secret}",
						resourceGroup: "${azurerm_resource_group.rg.name}",
						location: "${azurerm_resource_group.rg.location}",
						subnetName: "${azurerm_subnet.subnet.name}",
						securityGroupName: "${azurerm_network_security_group.sg.name}",
						vnetName: "${azurerm_virtual_network.vnet.name}",
					},
				},
				configure_vm: {
					template: "${file(\"configure-vm.sh\")}",
					vars: {
						apiserver_tar: "${base64encode(file(\"../../phase1b/crypto/apiserver.tar\"))}",
						root_tar: "${base64encode(file(\"../../phase1b/crypto/root.tar\"))}",
						kubelet_tar: "${base64encode(file(\"../../phase1b/crypto/kubelet.tar\"))}",
						k8s_config: "${base64encode(file(\"../../.config.json\"))}",
						azure_json: "${base64encode(template_file.azure_json.rendered)}",
					},
				}
			},
			azurerm_virtual_machine: {
				master_vm: {
					resource_group_name: names.resource_group,
					location: "${azurerm_resource_group.rg.location}",
					name: names.master_vm,
					network_interface_ids: ["${azurerm_network_interface.master_nic.id}"],
					vm_size: cfg.azure.master_vm_size,
					availability_set_id: "${azurerm_availability_set.as.id}",

					storage_image_reference: {
						publisher: cfg.azure.image_publisher,
						offer: cfg.azure.image_offer,
						sku: cfg.azure.image_sku,
						version: cfg.azure.image_version
					},

					storage_os_disk: {
						name: names.master_vm+"-osdisk",
						vhd_uri: "${azurerm_storage_account.sa.primary_blob_endpoint}${azurerm_storage_container.sc.name}/"+names.master_vm+"-osdisk.vhd",
						caching: "ReadWrite",
						create_option: "FromImage"
					},

					os_profile: {
						computer_name: names.master_vm,
						admin_username: cfg.azure.admin_username,
						admin_password: cfg.azure.admin_password,
						custom_data: "${base64encode(template_file.configure_vm.rendered)}",
					},

					os_profile_linux_config: {
						disable_password_authentication: false
					},
				},
				node_vm: {
					resource_group_name: names.resource_group,
					location: "${azurerm_resource_group.rg.location}",
					name: names.node_vm+"-${count.index}",
					network_interface_ids: ["${element(azurerm_network_interface.node_nic.*.id, count.index)}"],
					vm_size: cfg.azure.node_vm_size,
					availability_set_id: "${azurerm_availability_set.as.id}",

					storage_image_reference: {
						publisher: cfg.azure.image_publisher,
						offer: cfg.azure.image_offer,
						sku: cfg.azure.image_sku,
						version: cfg.azure.image_version
					},

					storage_os_disk: {
						name: names.node_vm+"-${count.index}-osdisk",
						vhd_uri: "${azurerm_storage_account.sa.primary_blob_endpoint}${azurerm_storage_container.sc.name}/"+names.node_vm+"${count.index}-osdisk.vhd",
						caching: "ReadWrite",
						create_option: "FromImage"
					},

					os_profile: {
						computer_name: names.node_vm+"-${count.index}",
						admin_username: cfg.azure.admin_username,
						admin_password: cfg.azure.admin_password,
						custom_data: "${base64encode(template_file.configure_vm.rendered)}",
					},

					os_profile_linux_config: {
						disable_password_authentication: false
					},

					count: cfg.num_nodes
				}
			},
			null_resource: {
				crypto_assets: {
					depends_on: [
						"azurerm_public_ip.pip",
					],
					provisioner: [{
						"local-exec": {	
							# clean is covering up a bug, perhaps in the makefile?
							command: "make -C ../../phase1b/crypto clean && make -C ../../phase1b/crypto"
						},
					}],
				},
			},
		},
	}
