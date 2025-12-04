package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1alpha1 "github.com/example/todo-operator/api/v1alpha1"
)

type TodoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *TodoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var todo webappv1alpha1.Todo
	if err := r.Get(ctx, req.NamespacedName, &todo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Reconciling Todo", "name", todo.Name)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      todo.Name + "-cm",
			Namespace: todo.Namespace,
		},
		Data: map[string]string{
			"foo": todo.Spec.Foo,
		},
	}
	// set controller reference so garbage collection works
	if err := ctrl.SetControllerReference(&todo, cm, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existing corev1.ConfigMap
	if err := r.Get(ctx, client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, &existing); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		existing.Data = cm.Data
		if err := r.Update(ctx, &existing); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *TodoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.Todo{}).
		Complete(r)
}
