/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1beta1"
	v12 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sphongcomv1alpha1 "github.com/seongpyoHong/elasticsearch-operator/api/v1alpha1"
)

// ElasticsearchReconciler reconciles a Elasticsearch object
type ElasticsearchReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=sphong.com.my.domain,resources=elasticsearches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sphong.com.my.domain,resources=elasticsearches/status,verbs=get;update;patch

func (r *ElasticsearchReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("elasticsearch", req.NamespacedName)

	//Fetch the Elasticsearch Cluster
	elasticsearch := &sphongcomv1alpha1.Elasticsearch{}
	err := r.Get(ctx, req.NamespacedName, elasticsearch)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Elasticsearch Resource Not Found!. Ignore since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Elasticsearch")
		return ctrl.Result{}, err
	}

	//Create Elasticsearch ConfigMap
	foundElasticsearchConf := &v12.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Namespace: elasticsearch.Namespace, Name: elasticsearch.Name + "-config"}, foundElasticsearchConf)

	if err != nil && errors.IsNotFound(err) {
		_ = r.createElasticsearchConfigMap(elasticsearch)
	}

	//Create Discovery Service
	foundMasterSvc := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-discovery", Namespace: elasticsearch.Namespace}, foundMasterSvc)

	if err != nil && errors.IsNotFound(err) {
		_ = r.createMasterService(elasticsearch)
	}

	//Create Client Service
	foundClientSvc := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-client", Namespace: elasticsearch.Namespace}, foundClientSvc)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createClientService(elasticsearch)
	}

	// Create Data Service
	foundDataSvc := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-data", Namespace: elasticsearch.Namespace}, foundDataSvc)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createDataService(elasticsearch)
	}
	//Master Node
	foundMaster := &v1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-master", Namespace: elasticsearch.Namespace}, foundMaster)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createDeploymentForMaster(elasticsearch)
	}

	//Client Node
	foundClient := &v1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-client", Namespace: elasticsearch.Namespace}, foundClient)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createDeploymentForClient(elasticsearch)
	}

	//Data Node (Hot)

	// First, Create Storage Class
	foundSsdStorageClass := &v1beta12.StorageClass{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-ssd", Namespace: elasticsearch.Namespace}, foundSsdStorageClass)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createSsdStorageClass(elasticsearch)
	}

	foundHotData := &v1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-data-hot", Namespace: elasticsearch.Namespace}, foundHotData)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createStatefulSetForHotData(elasticsearch)
	}

	//Data Node (Warm)
	//First, Create Storage Class
	foundHddStorageClass := &v1beta12.StorageClass{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-hdd", Namespace: elasticsearch.Namespace}, foundHddStorageClass)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createHddStorageClass(elasticsearch)
	}

	foundWarmData := &v1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-data-warm", Namespace: elasticsearch.Namespace}, foundWarmData)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createStatefulSetForWarmData(elasticsearch)
	}

	//Cerebro Deployment
	foundCerebroSvc := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-cerebro", Namespace: elasticsearch.Namespace}, foundCerebroSvc)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createCerebroService(elasticsearch)
	}

	foundCerebro := &v1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-cerebro", Namespace: elasticsearch.Namespace}, foundCerebro)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createDeploymentForCerebro(elasticsearch)
	}

	//Kibana Deployment
	foundKibanaSvc := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-kibana", Namespace: elasticsearch.Namespace}, foundKibanaSvc)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createKibanaService(elasticsearch)
	}

	foundKibana := &v12.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: elasticsearch.Name + "-kibana", Namespace: elasticsearch.Namespace}, foundKibana)
	if err != nil && errors.IsNotFound(err) {
		_ = r.createDeploymentForKibana(elasticsearch)
	}

	//scaling status's size => spec's size
	hotDataSize := elasticsearch.Spec.HotDataReplicas
	if *foundHotData.Spec.Replicas != hotDataSize {
		foundHotData.Spec.Replicas = &hotDataSize
		err = r.Client.Update(context.TODO(), foundHotData)
		if err != nil {
			log.Error(err, "Failed to update Deployment's Size (Hot Data Node Replica)")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	warmDataSize := elasticsearch.Spec.WarmDataReplicas
	if *foundWarmData.Spec.Replicas != warmDataSize {
		foundWarmData.Spec.Replicas = &warmDataSize
		err = r.Client.Update(context.TODO(), foundWarmData)
		if err != nil {
			log.Error(err, "Failed to update Deployment's Size")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &v12.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(elasticsearch.Namespace),
		client.MatchingLabels(map[string]string{"app": "elasticsearch-data", "type": "warm"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods")
		return ctrl.Result{}, err
	}
	//podNames := getPodNames(podList.Items)
	return ctrl.Result{}, nil
}

func (r *ElasticsearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sphongcomv1alpha1.Elasticsearch{}).
		Owns(&v1.Deployment{}).
		Owns(&v1.StatefulSet{}).
		Owns(&v1beta1.CronJob{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 3,
		}).Complete(r)
}
