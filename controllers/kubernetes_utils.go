package controllers

import (
	"context"
	sphongcomv1alpha1 "github.com/seongpyoHong/elasticsearch-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *ElasticsearchReconciler) createKibanaService(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForKibana()
	kibanaSvcName := e.Name + "-kibana"

	kibanaSvc := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kibanaSvcName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v12.ServiceSpec{
			Selector: labels,
			Ports: []v12.ServicePort{
				v12.ServicePort{
					Name:     "ui",
					Port:     5601,
					Protocol: "TCP",
				},
			},
			Type: v12.ServiceTypeLoadBalancer,
		},
	}

	if err := r.Client.Create(context.TODO(), kibanaSvc); err != nil {
		r.Log.Error(err, "Could not create kibana service! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createDeploymentForKibana(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForKibana()
	kibanaName := e.Name + "-kibana"
	probe := &v12.Probe{
		TimeoutSeconds:      30,
		InitialDelaySeconds: 1,
		FailureThreshold:    10,
		Handler: v12.Handler{
			HTTPGet: &v12.HTTPGetAction{
				Port:   intstr.FromInt(5601),
				Path:   "/",
				Scheme: v12.URISchemeHTTP,
			},
		},
	}

	kibanaDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kibanaName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:           kibanaName,
							Image:          e.Spec.Kibana.Image,
							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name:  "ELASTICSEARCH_HOSTS",
									Value: e.Name + "-client",
								},
								v12.EnvVar{
									Name:  "SERVER_HOST",
									Value: "0.0.0.0",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "ui",
									ContainerPort: 5601,
									Protocol:      v12.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	err := r.Client.Create(context.TODO(), kibanaDeployment)
	if err != nil {
		r.Log.Error(err, "Failed To Create Kibana Deployments...")
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) createCerebroService(e *sphongcomv1alpha1.Elasticsearch) error {
	cerebroSvcName := e.Name + "-cerebro"
	labels := labelsForCerebro()
	cerebroSvc := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cerebroSvcName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v12.ServiceSpec{
			Selector: labels,
			Ports: []v12.ServicePort{
				v12.ServicePort{
					Name:     "ui",
					Port:     9000,
					Protocol: "TCP",
				},
			},
			Type: v12.ServiceTypeLoadBalancer,
		},
	}

	if err := r.Client.Create(context.TODO(), cerebroSvc); err != nil {
		r.Log.Error(err, "Could not create cerebro service! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createDeploymentForCerebro(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForCerebro()
	cerebroName := e.Name + "-cerebro"
	probe := &v12.Probe{
		TimeoutSeconds:      30,
		InitialDelaySeconds: 1,
		FailureThreshold:    10,
		Handler: v12.Handler{
			HTTPGet: &v12.HTTPGetAction{
				Port:   intstr.FromInt(9000),
				Path:   "/#/connect", //TODO since cerebro doesn't have a healthcheck url, this path is enough
				Scheme: v12.URISchemeHTTP,
			},
		},
	}

	cerebroDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cerebroName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:           cerebroName,
							Image:          e.Spec.Cerebro.Image,
							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name:  "ELASTICSEARCH_URL",
									Value: e.Name + "-client",
								},
								v12.EnvVar{
									Name:  "CEREBRO_USERNAME",
									Value: "admin",
								},
								v12.EnvVar{
									Name:  "CEREBRO_PASSWORD",
									Value: "password",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "ui",
									ContainerPort: 9000,
									Protocol:      v12.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	err := r.Client.Create(context.TODO(), cerebroDeployment)
	if err != nil {
		r.Log.Error(err, "Failed To Create Cerebro Deployments...")
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) createStatefulSetForHotData(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForHotData()
	probe := &v12.Probe{
		TimeoutSeconds:      60,
		InitialDelaySeconds: 10,
		FailureThreshold:    10,
		SuccessThreshold:    1,
		Handler: v12.Handler{
			TCPSocket: &v12.TCPSocketAction{
				Port: intstr.FromInt(9300),
			},
		},
	}
	hotDiskSize, _ := resource.ParseQuantity(e.Spec.HotDataDiskSize)
	hotDataStatefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Name + "-data-hot",
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &e.Spec.HotDataReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Affinity: &v12.Affinity{
						NodeAffinity: &v12.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v12.NodeSelector{
								NodeSelectorTerms: []v12.NodeSelectorTerm{
									{
										MatchExpressions: []v12.NodeSelectorRequirement{
											{
												Key:      "cloud.google.com/gke-nodepool",
												Operator: "In",
												Values:   []string{"ssd-node-pool"},
											},
										},
									},
								},
							},
						},
					},
					Containers: []v12.Container{
						{
							Name:  e.Name + "-data-hot",
							Image: e.Spec.ElasticsearchImage,
							SecurityContext: &v12.SecurityContext{
								Privileged: &[]bool{true}[0],
								Capabilities: &v12.Capabilities{
									Add: []v12.Capability{
										"IPC_LOCK",
									},
								},
							},
							Args:           []string{"/run.sh", "-Enode.attr.box_type=hot"},
							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name: "NAMESPACE",
									ValueFrom: &v12.EnvVarSource{
										FieldRef: &v12.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v12.EnvVar{
									Name:  "CLUSTER_NAME",
									Value: e.Spec.ElasticsearchClusterName,
								},
								v12.EnvVar{
									Name:  "NODE_MASTER",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "NODE_DATA",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "NODE_INGEST",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "HTTP_ENABLE",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "ES_JAVA_OPTS",
									Value: e.Spec.HotDataJavaOpts,
								},
								v12.EnvVar{
									Name:  "ES_CLIENT_ENDPOINT",
									Value: e.Name + "-client",
								},
								v12.EnvVar{
									Name:  "ES_PERSISTENT",
									Value: "true",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "transport",
									ContainerPort: 9300,
									Protocol:      v12.ProtocolTCP,
								},
								v12.ContainerPort{
									Name:          "dummy",
									ContainerPort: 21212,
									Protocol:      v12.ProtocolTCP,
								},
							},
							VolumeMounts: []v12.VolumeMount{
								v12.VolumeMount{
									Name:      "data",
									MountPath: "/data",
								},
								v12.VolumeMount{
									Name:      e.Name + "-config",
									MountPath: "/elasticsearch-conf",
								},
							},
						},
					},
					Volumes: []v12.Volume{
						v12.Volume{
							Name: e.Name + "-config",
							VolumeSource: v12.VolumeSource{
								ConfigMap: &v12.ConfigMapVolumeSource{
									LocalObjectReference: v12.LocalObjectReference{
										Name: e.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v12.PersistentVolumeClaim{
				v12.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "data",
						Labels: labels,
					},
					Spec: v12.PersistentVolumeClaimSpec{
						AccessModes: []v12.PersistentVolumeAccessMode{
							v12.ReadWriteOnce,
						},
						StorageClassName: &[]string{"elasticsearch-ssd"}[0],
						Resources: v12.ResourceRequirements{
							Requests: v12.ResourceList{
								v12.ResourceStorage: hotDiskSize,
							},
						},
					},
				},
			},
		},
	}

	if err := r.Client.Create(context.TODO(), hotDataStatefulSet); err != nil {
		r.Log.Error(err, "Could not create hot data node! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createStatefulSetForWarmData(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForWarmData()
	probe := &v12.Probe{
		TimeoutSeconds:      60,
		InitialDelaySeconds: 10,
		FailureThreshold:    10,
		SuccessThreshold:    1,
		Handler: v12.Handler{
			TCPSocket: &v12.TCPSocketAction{
				Port: intstr.FromInt(9300),
			},
		},
	}
	warmDiskSize, _ := resource.ParseQuantity(e.Spec.WarmDataDiskSize)
	warmDataStatefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Name + "-data-warm",
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &e.Spec.WarmDataReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Affinity: &v12.Affinity{
						NodeAffinity: &v12.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v12.NodeSelector{
								NodeSelectorTerms: []v12.NodeSelectorTerm{
									{
										MatchExpressions: []v12.NodeSelectorRequirement{
											{
												Key:      "cloud.google.com/gke-nodepool",
												Operator: "In",
												Values:   []string{"hdd-node-pool"},
											},
										},
									},
								},
							},
						},
					},
					Containers: []v12.Container{
						{
							Name:  e.Name + "-data-warm",
							Image: e.Spec.ElasticsearchImage,
							SecurityContext: &v12.SecurityContext{
								Privileged: &[]bool{true}[0],
								Capabilities: &v12.Capabilities{
									Add: []v12.Capability{
										"IPC_LOCK",
									},
								},
							},
							Args:           []string{"/run.sh", "-Enode.attr.box_type=warm"},
							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name: "NAMESPACE",
									ValueFrom: &v12.EnvVarSource{
										FieldRef: &v12.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v12.EnvVar{
									Name:  "CLUSTER_NAME",
									Value: e.Spec.ElasticsearchClusterName,
								},
								v12.EnvVar{
									Name:  "NODE_MASTER",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "NODE_DATA",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "NODE_INGEST",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "HTTP_ENABLE",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "ES_JAVA_OPTS",
									Value: e.Spec.WarmDataJavaOpts,
								},
								v12.EnvVar{
									Name:  "ES_CLIENT_ENDPOINT",
									Value: e.Name + "-client",
								},
								v12.EnvVar{
									Name:  "ES_PERSISTENT",
									Value: "true",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "transport",
									ContainerPort: 9300,
									Protocol:      v12.ProtocolTCP,
								},
								v12.ContainerPort{
									Name:          "dummy",
									ContainerPort: 21213,
									Protocol:      v12.ProtocolTCP,
								},
							},
							VolumeMounts: []v12.VolumeMount{
								v12.VolumeMount{
									Name:      "data",
									MountPath: "/data",
								},
								v12.VolumeMount{
									Name:      e.Name + "-config",
									MountPath: "/elasticsearch-conf",
								},
							},
						},
					},
					Volumes: []v12.Volume{
						v12.Volume{
							Name: e.Name + "-config",
							VolumeSource: v12.VolumeSource{
								ConfigMap: &v12.ConfigMapVolumeSource{
									LocalObjectReference: v12.LocalObjectReference{
										Name: e.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v12.PersistentVolumeClaim{
				v12.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "data",
						Labels: labels,
					},
					Spec: v12.PersistentVolumeClaimSpec{
						AccessModes: []v12.PersistentVolumeAccessMode{
							v12.ReadWriteOnce,
						},
						StorageClassName: &[]string{"elasticsearch-hdd"}[0],
						Resources: v12.ResourceRequirements{
							Requests: v12.ResourceList{
								v12.ResourceStorage: warmDiskSize,
							},
						},
					},
				},
			},
		},
	}

	if err := r.Client.Create(context.TODO(), warmDataStatefulSet); err != nil {
		r.Log.Error(err, "Could not create warm data node! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createClientService(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForClient()
	clientSvcName := e.Name + "-client"

	clientSvc := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientSvcName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v12.ServiceSpec{
			Selector: labels,
			Ports: []v12.ServicePort{
				v12.ServicePort{
					Name:     "http",
					Port:     9200,
					Protocol: "TCP",
				},
			},
			Type: v12.ServiceTypeLoadBalancer,
		},
	}

	if err := r.Client.Create(context.TODO(), clientSvc); err != nil {
		r.Log.Error(err, "Could not create client service! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createMasterService(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForMaster()
	discoverySvcName := e.Name + "-discovery"

	discoverySvc := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      discoverySvcName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v12.ServiceSpec{
			Ports: []v12.ServicePort{
				v12.ServicePort{
					Name:     "transport",
					Port:     9300,
					Protocol: "TCP",
				},
			},
			Selector:  labels,
			ClusterIP: v12.ClusterIPNone,
		},
	}

	if err := r.Client.Create(context.TODO(), discoverySvc); err != nil {
		r.Log.Error(err, "Could not create discovery service! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createDataService(e *sphongcomv1alpha1.Elasticsearch) error {
	dataSvcName := e.Name + "-data"

	dataSvc := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataSvcName,
			Namespace: e.Namespace,
			Labels: map[string]string{
				"app": "elasticsearch-data",
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: v12.ServiceSpec{
			Ports: []v12.ServicePort{
				v12.ServicePort{
					Name:     "transport",
					Port:     9300,
					Protocol: "TCP",
				},
			},
			Selector: map[string]string{
				"app": "elasticsearch-data",
			},
			ClusterIP: v12.ClusterIPNone,
		},
	}

	if err := r.Client.Create(context.TODO(), dataSvc); err != nil {
		r.Log.Error(err, "Could not create data service! ")
		return err
	}

	return nil
}

func (r *ElasticsearchReconciler) createHddStorageClass(e *sphongcomv1alpha1.Elasticsearch) error {
	hddStorageClass := &v1beta12.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "elasticsearch-hdd",
		},
		Provisioner: "kubernetes.io/gce-pd",
		Parameters:  map[string]string{"type": "pd-standard"},
	}
	err := r.Client.Create(context.TODO(), hddStorageClass)
	if err != nil {
		r.Log.Error(err, "Failed to create HDD storage class..")
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) createSsdStorageClass(e *sphongcomv1alpha1.Elasticsearch) error {
	ssdStorageClass := &v1beta12.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "elasticsearch-ssd",
		},
		Provisioner: "kubernetes.io/gce-pd",
		Parameters:  map[string]string{"type": "pd-ssd"},
	}
	err := r.Client.Create(context.TODO(), ssdStorageClass)
	if err != nil {
		r.Log.Error(err, "Failed to create SSD storage class..")
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) createDeploymentForClient(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForClient()
	clientName := e.Name + "-client"
	probe := &v12.Probe{
		TimeoutSeconds:      60,
		InitialDelaySeconds: 10,
		FailureThreshold:    10,
		SuccessThreshold:    1,
		Handler: v12.Handler{
			HTTPGet: &v12.HTTPGetAction{
				Port: intstr.FromInt(9200),
				Path: "/_cluster/health?wait_for_status=yellow&timeout=60s",
			},
		},
	}
	clientDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &e.Spec.ClientReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:  clientName,
							Image: e.Spec.ElasticsearchImage,
							SecurityContext: &v12.SecurityContext{
								Privileged: &[]bool{true}[0],
								Capabilities: &v12.Capabilities{
									Add: []v12.Capability{
										"IPC_LOCK",
									},
								},
							},

							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name: "NAMESPACE",
									ValueFrom: &v12.EnvVarSource{
										FieldRef: &v12.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v12.EnvVar{
									Name:  "CLUSTER_NAME",
									Value: e.Spec.ElasticsearchClusterName,
								},
								v12.EnvVar{
									Name:  "NODE_MASTER",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "NODE_DATA",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "NODE_INGEST",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "HTTP_ENABLE",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "ES_JAVA_OPTS",
									Value: e.Spec.ClientJavaOpts,
								},
								v12.EnvVar{
									Name:  "ES_CLIENT_ENDPOINT",
									Value: e.Name + "-client",
								},
								v12.EnvVar{
									Name:  "NETWORK_HOST",
									Value: "0.0.0.0",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "transport",
									ContainerPort: 9300,
									Protocol:      v12.ProtocolTCP,
								},
								v12.ContainerPort{
									Name:          "http",
									ContainerPort: 9200,
									Protocol:      v12.ProtocolTCP,
								},
							},
							VolumeMounts: []v12.VolumeMount{
								v12.VolumeMount{
									Name:      "data",
									MountPath: "/data",
								},
								v12.VolumeMount{
									Name:      e.Name + "-config",
									MountPath: "/elasticsearch-conf",
								},
							},
						},
					},
					Volumes: []v12.Volume{
						v12.Volume{
							Name: "data",
							VolumeSource: v12.VolumeSource{
								EmptyDir: &v12.EmptyDirVolumeSource{},
							},
						},
						v12.Volume{
							Name: e.Name + "-config",
							VolumeSource: v12.VolumeSource{
								ConfigMap: &v12.ConfigMapVolumeSource{
									LocalObjectReference: v12.LocalObjectReference{
										Name: e.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := r.Client.Create(context.TODO(), clientDeployment)
	if err != nil {
		r.Log.Error(err, "Failed To Create Client Deployments...")
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) createDeploymentForMaster(e *sphongcomv1alpha1.Elasticsearch) error {
	labels := labelsForMaster()
	masterName := e.Name + "-master"
	probe := &v12.Probe{
		TimeoutSeconds:      60,
		InitialDelaySeconds: 10,
		FailureThreshold:    10,
		SuccessThreshold:    1,
		Handler: v12.Handler{
			TCPSocket: &v12.TCPSocketAction{
				Port: intstr.FromInt(9300),
			},
		},
	}
	masterDeployments := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterName,
			Namespace: e.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &e.Spec.MasterReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:  masterName,
							Image: e.Spec.ElasticsearchImage,
							SecurityContext: &v12.SecurityContext{
								Privileged: &[]bool{true}[0],
								Capabilities: &v12.Capabilities{
									Add: []v12.Capability{
										"IPC_LOCK",
									},
								},
							},

							ReadinessProbe: probe,
							Env: []v12.EnvVar{
								v12.EnvVar{
									Name: "NAMESPACE",
									ValueFrom: &v12.EnvVarSource{
										FieldRef: &v12.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v12.EnvVar{
									Name:  "CLUSTER_NAME",
									Value: e.Spec.ElasticsearchClusterName,
								},
								v12.EnvVar{
									Name:  "NODE_MASTER",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "NODE_DATA",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "NODE_INGEST",
									Value: "false",
								},
								v12.EnvVar{
									Name:  "HTTP_ENABLE",
									Value: "true",
								},
								v12.EnvVar{
									Name:  "ES_JAVA_OPTS",
									Value: e.Spec.MasterJavaOpts,
								},
								v12.EnvVar{
									Name:  "ES_CLIENT_ENDPOINT",
									Value: e.Name + "-client",
								},
							},
							Ports: []v12.ContainerPort{
								v12.ContainerPort{
									Name:          "transport",
									ContainerPort: 9300,
									Protocol:      v12.ProtocolTCP,
								},
							},
							VolumeMounts: []v12.VolumeMount{
								v12.VolumeMount{
									Name:      "data",
									MountPath: "/data",
								},
								v12.VolumeMount{
									Name:      e.Name + "-config",
									MountPath: "/elasticsearch-conf",
								},
							},
						},
					},
					Volumes: []v12.Volume{
						v12.Volume{
							Name: "data",
							VolumeSource: v12.VolumeSource{
								EmptyDir: &v12.EmptyDirVolumeSource{},
							},
						},
						v12.Volume{
							Name: e.Name + "-config",
							VolumeSource: v12.VolumeSource{
								ConfigMap: &v12.ConfigMapVolumeSource{
									LocalObjectReference: v12.LocalObjectReference{
										Name: e.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	err := r.Client.Create(context.TODO(), masterDeployments)
	if err != nil {
		r.Log.Error(err, "Failed To Create Master Deployments...")
		return err
	}
	return nil
}

func createElasticsearchConf() map[string]string {
	ret := map[string]string{}
	ret["elasticsearch.yml"] = `
##############################   Mandantary Field  #############################
cluster:
  name: ${CLUSTER_NAME}
node:
  master: ${NODE_MASTER}
  data: ${NODE_DATA}
  ingest: ${NODE_INGEST}
  name: ${HOSTNAME}
network.host: ${NETWORK_HOST}
path:
  data: /data/data
  logs: /data/log
bootstrap.memory_lock: true
http:
  enabled: ${HTTP_ENABLE}
  compression: true
  cors:
	enabled: ${HTTP_CORS_ENABLE}
	allow-origin: ${HTTP_CORS_ALLOW_ORIGIN}
discovery:
  zen:
	ping.unicast.hosts: ${DISCOVERY_SERVICE}
	minimum_master_nodes: ${NUMBER_OF_MASTERS}
##############################   Add Here!  #############################
`

	ret["log4j2.properties"] = `status = error
    # log action execution errors for easier debugging
    logger.action.name = org.elasticsearch.action
    logger.action.level = debug
    appender.console.type = Console
    appender.console.name = console
    appender.console.layout.type = PatternLayout
    appender.console.layout.pattern = [%d{ISO8601}][%-5p][%-25c{1.}] [%node_name]%marker %m%n
    appender.rolling.type = RollingFile
    appender.rolling.name = rolling
    appender.rolling.fileName = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}.log
    appender.rolling.layout.type = PatternLayout
    appender.rolling.layout.pattern = [%d{ISO8601}][%-5p][%-25c{1.}] [%node_name]%marker %.-10000m%n
    appender.rolling.filePattern = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}-%d{yyyy-MM-dd}-%i.log.gz
    appender.rolling.policies.type = Policies
    appender.rolling.policies.time.type = TimeBasedTriggeringPolicy
    appender.rolling.policies.time.interval = 1
    appender.rolling.policies.time.modulate = true
    appender.rolling.policies.size.type = SizeBasedTriggeringPolicy
    appender.rolling.policies.size.size = 128MB
    appender.rolling.strategy.type = DefaultRolloverStrategy
    appender.rolling.strategy.fileIndex = nomax
    appender.rolling.strategy.action.type = Delete
    appender.rolling.strategy.action.basepath = ${sys:es.logs.base_path}
    appender.rolling.strategy.action.condition.type = IfFileName
    appender.rolling.strategy.action.condition.glob = ${sys:es.logs.cluster_name}-*
    appender.rolling.strategy.action.condition.nested_condition.type = IfAccumulatedFileSize
    appender.rolling.strategy.action.condition.nested_condition.exceeds = 2GB
    rootLogger.level = info
    rootLogger.appenderRef.console.ref = console
    rootLogger.appenderRef.rolling.ref = rolling
    appender.deprecation_rolling.type = RollingFile
    appender.deprecation_rolling.name = deprecation_rolling
    appender.deprecation_rolling.fileName = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_deprecation.log
    appender.deprecation_rolling.layout.type = PatternLayout
    appender.deprecation_rolling.layout.pattern = [%d{ISO8601}][%-5p][%-25c{1.}] [%node_name]%marker %.-10000m%n
    appender.deprecation_rolling.filePattern = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_deprecation-%i.log.gz
    appender.deprecation_rolling.policies.type = Policies
    appender.deprecation_rolling.policies.size.type = SizeBasedTriggeringPolicy
    appender.deprecation_rolling.policies.size.size = 1GB
    appender.deprecation_rolling.strategy.type = DefaultRolloverStrategy
    appender.deprecation_rolling.strategy.max = 4
    logger.deprecation.name = org.elasticsearch.deprecation
    logger.deprecation.level = warn
    logger.deprecation.appenderRef.deprecation_rolling.ref = deprecation_rolling
    logger.deprecation.additivity = false
    appender.index_search_slowlog_rolling.type = RollingFile
    appender.index_search_slowlog_rolling.name = index_search_slowlog_rolling
    appender.index_search_slowlog_rolling.fileName = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_index_search_slowlog.log
    appender.index_search_slowlog_rolling.layout.type = PatternLayout
    appender.index_search_slowlog_rolling.layout.pattern = [%d{ISO8601}][%-5p][%-25c] [%node_name]%marker %.-10000m%n
    appender.index_search_slowlog_rolling.filePattern = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_index_search_slowlog-%d{yyyy-MM-dd}.log
    appender.index_search_slowlog_rolling.policies.type = Policies
    appender.index_search_slowlog_rolling.policies.time.type = TimeBasedTriggeringPolicy
    appender.index_search_slowlog_rolling.policies.time.interval = 1
    appender.index_search_slowlog_rolling.policies.time.modulate = true
    logger.index_search_slowlog_rolling.name = index.search.slowlog
    logger.index_search_slowlog_rolling.level = trace
    logger.index_search_slowlog_rolling.appenderRef.index_search_slowlog_rolling.ref = index_search_slowlog_rolling
    logger.index_search_slowlog_rolling.additivity = false
    appender.index_indexing_slowlog_rolling.type = RollingFile
    appender.index_indexing_slowlog_rolling.name = index_indexing_slowlog_rolling
    appender.index_indexing_slowlog_rolling.fileName = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_index_indexing_slowlog.log
    appender.index_indexing_slowlog_rolling.layout.type = PatternLayout
    appender.index_indexing_slowlog_rolling.layout.pattern = [%d{ISO8601}][%-5p][%-25c] [%node_name]%marker %.-10000m%n
    appender.index_indexing_slowlog_rolling.filePattern = ${sys:es.logs.base_path}${sys:file.separator}${sys:es.logs.cluster_name}_index_indexing_slowlog-%d{yyyy-MM-dd}.log
    appender.index_indexing_slowlog_rolling.policies.type = Policies
    appender.index_indexing_slowlog_rolling.policies.time.type = TimeBasedTriggeringPolicy
    appender.index_indexing_slowlog_rolling.policies.time.interval = 1
    appender.index_indexing_slowlog_rolling.policies.time.modulate = true
    logger.index_indexing_slowlog.name = index.indexing.slowlog.index
    logger.index_indexing_slowlog.level = trace
    logger.index_indexing_slowlog.appenderRef.index_indexing_slowlog_rolling.ref = index_indexing_slowlog_rolling
    logger.index_indexing_slowlog.additivity = false
`
	ret["logging.yml"] = `es.logger.level: INFO
    rootLogger: ${es.logger.level}, console, file
    logger:
      # log action execution errors for easier debugging
      action: DEBUG
      # reduce the logging for aws, too much is logged under the default INFO
      com.amazonaws: WARN
      # gateway
      #gateway: DEBUG
      #index.gateway: DEBUG
      # peer shard recovery
      #indices.recovery: DEBUG
      # discovery
      #discovery: TRACE
      index.search.slowlog: TRACE, index_search_slow_log_file
      index.indexing.slowlog: TRACE, index_indexing_slow_log_file
    additivity:
      index.search.slowlog: false
      index.indexing.slowlog: false
    appender:
      console:
        type: console
        layout:
          type: consolePattern
          conversionPattern: "[%d{ISO8601}][%-5p][%-25c] %m%n"
      file:
        type: dailyRollingFile
        file: ${path.logs}/${cluster.name}.log
        datePattern: "'.'yyyy-MM-dd"
        layout:
          type: pattern
          conversionPattern: "[%d{ISO8601}][%-5p][%-25c] %m%n"
      index_search_slow_log_file:
        type: dailyRollingFile
        file: ${path.logs}/${cluster.name}_index_search_slowlog.log
        datePattern: "'.'yyyy-MM-dd"
        layout:
          type: pattern
          conversionPattern: "[%d{ISO8601}][%-5p][%-25c] %m%n"
      index_indexing_slow_log_file:
        type: dailyRollingFile
        file: ${path.logs}/${cluster.name}_index_indexing_slowlog.log
        datePattern: "'.'yyyy-MM-dd"
        layout:
          type: pattern
          conversionPattern: "[%d{ISO8601}][%-5p][%-25c] %m%n"
`
	return ret
}

func (r *ElasticsearchReconciler) createElasticsearchConfigMap(e *sphongcomv1alpha1.Elasticsearch) error {
	data := createElasticsearchConf()
	cf := v12.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e.Namespace,
			Name:      e.Name + "-config",
		},
		Data: data,
	}

	err := r.Client.Create(context.TODO(), &cf)
	if err != nil {
		r.Log.Error(err, "Failed to create new Elasticsearch ConfigMap.")
		return err
	}
	return nil
}

func labelsForMaster() map[string]string {
	return map[string]string{"app": "elasticsearch-master"}
}

func labelsForClient() map[string]string {
	return map[string]string{"app": "elasticsearch-client"}
}

func labelsForHotData() map[string]string {
	return map[string]string{"app": "elasticsearch-data", "type": "hot"}
}

func labelsForWarmData() map[string]string {
	return map[string]string{"app": "elasticsearch-data", "type": "warm"}
}

func labelsForCerebro() map[string]string {
	return map[string]string{"app": "elasticsearch-cerebro"}
}

func labelsForKibana() map[string]string {
	return map[string]string{"app": "elasticsearch-kibana"}
}
