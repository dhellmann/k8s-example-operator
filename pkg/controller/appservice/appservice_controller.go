package appservice

import (
	"context"
	"fmt"
	"reflect"

	appv1alpha1 "github.com/dhellmann/k8s-example-operator/pkg/apis/app/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_appservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	log.Info("newReconciler")
	return &ReconcileAppService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Info("add")
	// Create a new controller
	c, err := controller.New("appservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppService
	err = c.Watch(&source.Kind{Type: &appv1alpha1.AppService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AppService
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AppService{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAppService{}

// ReconcileAppService reconciles a AppService object
type ReconcileAppService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AppService object and makes changes based on the state read
// and what is in the AppService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppService")

	// Fetch the AppService instance
	instance := &appv1alpha1.AppService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Request object not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Info("Error reading the object, requeuing")
		return reconcile.Result{}, err
	}

	// Handle deletes

	// If object hasn't been deleted and doesn't have a finalizer, add one
	// Add a finalizer to newly created objects.
	if instance.ObjectMeta.DeletionTimestamp.IsZero() &&
		!stringInList(instance.ObjectMeta.Finalizers, appv1alpha1.AppServiceFinalizer) {
		reqLogger.Info(
			"adding finalizer",
			"existingFinalizers", instance.ObjectMeta.Finalizers,
			"newValue", appv1alpha1.AppServiceFinalizer,
		)
		instance.Finalizers = append(instance.Finalizers,
			appv1alpha1.AppServiceFinalizer)
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Info(
				"failed to add finalizer to AppService object due to error",
				"error", err)
			return reconcile.Result{}, err
		}
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		reqLogger.Info(
			"marked to be deleted",
			"timestamp", instance.ObjectMeta.DeletionTimestamp,
		)
		// no-op if finalizer has been removed.
		if !stringInList(instance.ObjectMeta.Finalizers, appv1alpha1.AppServiceFinalizer) {
			reqLogger.Info("reconciling AppService object causes a no-op as there is no finalizer")
			return reconcile.Result{}, nil
		}

		// Here is where we would do something with external resources
		// not managed through CRs (those are deleted automatically).

		// Remove finalizer to allow deletion
		reqLogger.Info("removing finalizer")
		instance.ObjectMeta.Finalizers = filterStringFromList(
			instance.ObjectMeta.Finalizers, appv1alpha1.AppServiceFinalizer)
		if err := r.client.Update(context.Background(), instance); err != nil {
			reqLogger.Info("Error removing finalizer from AppService object")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil // done

	}

	// Handle creates/updates

	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(),
		types.NamespacedName{Name: instance.Name,
			Namespace: instance.Namespace},
		deployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.createDeployment(instance)
		reqLogger.Info("Creating a new Deployment",
			"Deployment.Namespace", dep.Namespace,
			"Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "failed to create new Deployment",
				"Deployment.Namespace", dep.Namespace,
				"Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		reqLogger.Info("new deployment created; requeueing")
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := instance.Spec.Size
	if *deployment.Spec.Replicas != size {
		reqLogger.Info("updating deployment spec", "size", size)
		deployment.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "failed to update Deployment",
				"Deployment.Namespace", deployment.Namespace,
				"Deployment.Name", deployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		reqLogger.Info("updated deployment spepc; requeuing")
		return reconcile.Result{Requeue: true}, nil
	}

	// Ensure the deployment labels match the desired values.
	if !labels.Equals(instance.Spec.DeploymentLabels, deployment.ObjectMeta.Labels) {
		deployment.ObjectMeta.Labels = instance.Spec.DeploymentLabels
		reqLogger.Info("updating deployment labels")
		err = r.client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "failed to update Deployment",
				"Deployment.Namespace", deployment.Namespace,
				"Deployment.Name", deployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		reqLogger.Info("updated deployment labels; requeuing")
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForApp(instance.Name))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "failed to list pods", "AppService.Namespace",
			instance.Namespace, "AppService.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	reqLogger.Info("got pod names", "podNames", podNames)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "failed to update Memcached status")
			return reconcile.Result{}, err
		}
		reqLogger.Info("updated status")
	}

	// If the list of pods we got is different than the number
	// expected the deployment is out of date and we should try to
	// update oureslves again
	if len(podNames) != int(size) {
		reqLogger.Info(fmt.Sprintf("found %d pods instead of %d; rescheduling to wait for deployment to update pods", len(podNames), size))
		return reconcile.Result{Requeue: true}, nil
	}

	// Pod already exists - don't requeue
	reqLogger.Info("No more reconcile work to do")
	return reconcile.Result{}, nil
}

// createDeployment returns a memcached Deployment object
func (r *ReconcileAppService) createDeployment(a *appv1alpha1.AppService) *appsv1.Deployment {
	ls := labelsForApp(a.Name)
	replicas := a.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "busybox",
						Image:   "busybox",
						Command: []string{"sleep", "3600"},
						// Image:   "memcached:1.4.36-alpine",
						// Name:    "memcached",
						// Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "demoapp",
						}},
					}},
				},
			},
		},
	}
	// Set instance as the owner and controller
	controllerutil.SetControllerReference(a, dep, r.scheme)
	return dep
}

// labelsForApp returns the labels for selecting the resources
// belonging to the given AppService CR name.
func labelsForApp(name string) map[string]string {
	return map[string]string{"appservice": "demo-app", "appservice_cr": name}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.AppService) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func stringInList(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}
	return false
}

func filterStringFromList(list []string, strToFilter string) (newList []string) {
	for _, item := range list {
		if item != strToFilter {
			newList = append(newList, item)
		}
	}
	return
}
