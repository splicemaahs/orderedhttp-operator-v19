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
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	orderedhttpv1alpha1 "github.com/splicemaahs/orderedhttp-operator/api/v1alpha1"
)

// OrderedHttpReconciler reconciles a OrderedHttp object
type OrderedHttpReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=orderedhttp.splicemachine.io,resources=orderedhttps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=orderedhttp.splicemachine.io,resources=orderedhttps/status,verbs=get;update;patch

func (r *OrderedHttpReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("orderedhttp", req.NamespacedName)

	// your logic here
	// Fetch the OrderedHttp instance
	orderedHttp := &orderedhttpv1alpha1.OrderedHttp{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, orderedHttp)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// List all pods owned by this OrderedHttp instance
	lbls := labels.Set{
		"app":     orderedHttp.Name,
		"version": "0.1",
	}
	var allPodsReady bool
	allPodsReady = true
	existingPods := &corev1.PodList{}
	err = r.Client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     req.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		},
	)
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the orderedHttp pod")
		return ctrl.Result{}, err
	}
	existingPodNames := []string{}

	// Count the pods that are pending or running as available
	for _, pod := range existingPods.Items {
		reqLogger.Info("1.1 Loop Pods", "PodName: ", pod.Name)
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		reqLogger.Info("1.2 Pod Phase", "Phase: ", pod.Status.Phase)
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
		if pod.Status.Phase != corev1.PodRunning {
			allPodsReady = false
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready == false {
				allPodsReady = false
			}
			reqLogger.Info("1.2.1 Container Statuses: ", "Ready: ", strconv.FormatBool(containerStatus.Ready), "All Pods Ready: ", strconv.FormatBool(allPodsReady))
		}
	}
	reqLogger.Info("2. Checking orderedHttp", "expected size", orderedHttp.Spec.Replicas, "Pod.Names", existingPodNames)

	// List the pods for this deployment
	// podList := &corev1.PodList{}
	// podNames := getPodNames(podList.Items)
	// orderedHttp.Status.PodNames = podNames
	// reqLogger.Info("0. Setting Pod Names in status", "Pod.Names", existingPodNames)
	orderedHttp.Status.PodNames = existingPodNames
	err = r.Client.Status().Update(context.TODO(), orderedHttp)
	if err != nil {
		reqLogger.Error(err, "failed to update the orderedHttp pod")
		return ctrl.Result{}, err
	}

	// Scale Down Pods
	if int32(len(existingPodNames)) > orderedHttp.Spec.Replicas {
		// When scaling down, just delete one, and allow the process to continue, the next loop will determine additional removals
		reqLogger.Info("2.1 Deleting a pod in orderedHttp set", "expected size", orderedHttp.Spec.Replicas, "Pod.Names", existingPodNames)
		pod := existingPods.Items[0]
		err = r.Client.Delete(context.TODO(), &pod)
		if err != nil {
			reqLogger.Error(err, "failed to delete a pod")
			return ctrl.Result{}, err
		}
	}

	// Scale Up Pods
	if int32(len(existingPodNames)) < orderedHttp.Spec.Replicas {
		// When scaling up, just add one, and allow the process to contiue, the next loop will add more pods if needed.
		reqLogger.Info("2.2 Pod Ready Check", "All Pods Ready: ", strconv.FormatBool(allPodsReady))
		if allPodsReady == true {
			reqLogger.Info("2.2.1 Adding a pod in orderedHttp set", "expected size", orderedHttp.Spec.Replicas, "Pod.Names", existingPodNames)
			pod := newPodForCR(orderedHttp)
			if err := controllerutil.SetControllerReference(orderedHttp, pod, r.Scheme); err != nil {
				reqLogger.Error(err, "unable to set owner reference on new pod")
				return ctrl.Result{}, err
			}
			reqLogger.Info("2.2.2 Create Pod")
			err = r.Client.Create(context.TODO(), pod)
			if err != nil {
				reqLogger.Error(err, "failed to create a pod")
				return ctrl.Result{}, err
			}
		}
	}

	// end my logic here

	return ctrl.Result{}, nil
}

func (r *OrderedHttpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orderedhttpv1alpha1.OrderedHttp{}).
		Complete(r)
}

// getPodNames returns the pod names of the array of pods passed in.
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *orderedhttpv1alpha1.OrderedHttp) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name,
		"version": "0.1",
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-",
			Namespace:    cr.Namespace,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-delay",
					Image: "splicemaahs/nginx-delay:latest",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: 80,
								},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
					},
				},
			},
		},
	}
}
