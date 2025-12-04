/*
Copyright 2025.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webappv1alpha1 "github.com/example/todo-operator/api/v1alpha1"

	// add these imports:
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TodoReconciler reconciles a Todo object
type TodoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.example.com,resources=todoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.example.com,resources=todoes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.example.com,resources=todoes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Todo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *TodoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var todo webappv1alpha1.Todo
	if err := r.Get(ctx, req.NamespacedName, &todo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      todo.Name + "-cm",
			Namespace: todo.Namespace,
		},
		Data: map[string]string{
			"foo": todo.Spec.Foo,
		},
	}

	if err := controllerutil.SetControllerReference(&todo, cm, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existing corev1.ConfigMap
	if err := r.Get(ctx, client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, &existing); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		// Update mutable fields
		existing.Data = cm.Data
		if err := r.Update(ctx, &existing); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TodoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.Todo{}).
		Named("todo").
		Complete(r)
}
