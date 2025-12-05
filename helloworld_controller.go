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
    "fmt"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/apimachinery/pkg/util/intstr"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    logf "sigs.k8s.io/controller-runtime/pkg/log"

    cachev1alpha1 "github.com/scelios/kind/api/v1alpha1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.localhost,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.localhost,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cache.localhost,resources=helloworlds/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := logf.FromContext(ctx)

    // Fetch the HelloWorld instance
    helloWorld := &cachev1alpha1.HelloWorld{}
    if err := r.Get(ctx, req.NamespacedName, helloWorld); err != nil {
        log.Error(err, "unable to fetch HelloWorld")
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Log the message from the spec
    log.Info("HelloWorld reconciled", "name", helloWorld.Name, "message", helloWorld.Spec.Message, "replicas", helloWorld.Spec.Replicas)

    // Create ConfigMap with HTML content
    configMap := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      helloWorld.Name + "-html",
            Namespace: helloWorld.Namespace,
        },
        Data: map[string]string{
            "index.html": fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Hello World Operator</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .container {
            text-align: center;
            background: rgba(255, 255, 255, 0.1);
            padding: 50px;
            border-radius: 20px;
            box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37);
        }
        h1 {
            font-size: 3em;
            margin: 0;
            animation: fadeIn 1s;
        }
        p {
            font-size: 1.5em;
            margin-top: 20px;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-20px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1> %s</h1>
        <p>Replicas: %d</p>
        <p><small>Deployed by Kubernetes Operator</small></p>
    </div>
</body>
</html>`, helloWorld.Spec.Message, *helloWorld.Spec.Replicas),
        },
    }

    if err := controllerutil.SetControllerReference(helloWorld, configMap, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }

    foundConfigMap := &corev1.ConfigMap{}
    err := r.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, foundConfigMap)
    if err != nil && errors.IsNotFound(err) {
        log.Info("Creating ConfigMap", "name", configMap.Name)
        if err = r.Create(ctx, configMap); err != nil {
            return ctrl.Result{}, err
        }
    } else if err == nil {
        log.Info("Updating ConfigMap", "name", configMap.Name)
        if err = r.Update(ctx, configMap); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Create Deployment
    replicas := int32(1)
    if helloWorld.Spec.Replicas != nil {
        replicas = *helloWorld.Spec.Replicas
    }

    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      helloWorld.Name,
            Namespace: helloWorld.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": helloWorld.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": helloWorld.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "nginx",
                            Image: "nginx:alpine",
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: 80,
                                    Protocol:      corev1.ProtocolTCP,
                                },
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "html",
                                    MountPath: "/usr/share/nginx/html",
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "html",
                            VolumeSource: corev1.VolumeSource{
                                ConfigMap: &corev1.ConfigMapVolumeSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: configMap.Name,
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    if err := controllerutil.SetControllerReference(helloWorld, deployment, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }

    foundDeployment := &appsv1.Deployment{}
    err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
    if err != nil && errors.IsNotFound(err) {
        log.Info("Creating Deployment", "name", deployment.Name)
        if err = r.Create(ctx, deployment); err != nil {
            return ctrl.Result{}, err
        }
    } else if err == nil {
        log.Info("Updating Deployment", "name", deployment.Name)
        foundDeployment.Spec = deployment.Spec
        if err = r.Update(ctx, foundDeployment); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Create Service
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      helloWorld.Name,
            Namespace: helloWorld.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{
                "app": helloWorld.Name,
            },
            Type: corev1.ServiceTypeNodePort,
            Ports: []corev1.ServicePort{
                {
                    Port:       80,
                    TargetPort: intstr.FromInt(80),
                    Protocol:   corev1.ProtocolTCP,
                },
            },
        },
    }

    if err := controllerutil.SetControllerReference(helloWorld, service, r.Scheme); err != nil {
        return ctrl.Result{}, err
    }

    foundService := &corev1.Service{}
    err = r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
    if err != nil && errors.IsNotFound(err) {
        log.Info("Creating Service", "name", service.Name)
        if err = r.Create(ctx, service); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Update status with the message
    helloWorld.Status.Message = helloWorld.Spec.Message
    if err := r.Status().Update(ctx, helloWorld); err != nil {
        log.Error(err, "unable to update HelloWorld status")
        return ctrl.Result{}, err
    }

    log.Info("Successfully reconciled HelloWorld", "name", helloWorld.Name)
    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&cachev1alpha1.HelloWorld{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&corev1.ConfigMap{}).
        Named("helloworld").
        Complete(r)
}