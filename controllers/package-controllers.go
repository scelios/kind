package controllers

import (
    "context"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"

    corev1 "k8s.io/api/core/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    webappv1 "github.com/scelios/kind/api/v1"
)

type HelloReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *HelloReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    var hello webappv1.Hello
    if err := r.Get(ctx, req.NamespacedName, &hello); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    cm := &corev1.ConfigMap{}
    cm.Name = hello.Name
    cm.Namespace = hello.Namespace

    // desired state
    desired := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      hello.Name,
            Namespace: hello.Namespace,
        },
        Data: map[string]string{
            "message": hello.Spec.Message,
        },
    }

    // create or update ConfigMap
    if err := ctrl.SetControllerReference(&hello, desired, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }
    if err := r.Patch(ctx, desired, client.Apply, client.FieldOwner("hello-controller")); err != nil {
        logger.Error(err, "unable to apply configmap")
        return ctrl.Result{}, err
    }

    // update status (simple example)
    hello.Status.Seen = true
    _ = r.Status().Update(ctx, &hello)

    return ctrl.Result{}, nil
}

func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&webappv1.Hello{}).
        Complete(r)
}