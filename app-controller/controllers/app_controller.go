/*
Copyright 2023.

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
	"github.com/SeirousLee/app-controller/controllers/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoexposev1alpha1 "github.com/SeirousLee/app-controller/api/v1alpha1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=autoexpose.sinlee.com,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoexpose.sinlee.com,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=autoexpose.sinlee.com,resources=apps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	app := &autoexposev1alpha1.App{}
	// get customize resource
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// according to customize spec handle
	// handle deployment
	deployment := utils.NewDeployment(app)
	err = controllerutil.SetControllerReference(app, deployment, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	d := &v1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, d)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "create deploy failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	//2. Service的处理
	service := utils.NewService(app)
	if err := controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找指定service
	s := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, s); err != nil {
		if errors.IsNotFound(err) && app.Spec.Service {
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "create service failed")
				return ctrl.Result{}, err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Service {
			return ctrl.Result{}, err
		}
	} else {
		if app.Spec.Service {
			logger.Info("skip update")
		} else {
			if err := r.Delete(ctx, s); err != nil {
				return ctrl.Result{}, err
			}

		}
	}

	//3. Ingress的处理,ingress配置可能为空
	ingress := utils.NewIngress(app)
	if err := controllerutil.SetControllerReference(app, ingress, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	i := &netv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, i); err != nil {
		if errors.IsNotFound(err) && app.Spec.Ingress {
			if err := r.Create(ctx, ingress); err != nil {
				logger.Error(err, "create ingress failed")
				return ctrl.Result{}, err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Ingress {
			return ctrl.Result{}, err
		}
	} else {
		if app.Spec.Ingress {
			logger.Info("skip update")
		} else {
			if err := r.Delete(ctx, i); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoexposev1alpha1.App{}).
		Owns(&v1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&netv1.Ingress{}).
		Complete(r)
}
