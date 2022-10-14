/*
Copyright 2022 Appscode Inc..

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myapi "kubedb.dev/mssql/api/v1alpha1"
)

// MSSQLReconciler reconciles a MSSQL object
type MSSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services;secrets,verbs=get;list;watch;create;patch;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;patch;update;delete

func (r *MSSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var db myapi.MSSQL
	err := r.Client.Get(context.TODO(), req.NamespacedName, &db)
	if err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}
	klog.Infof("Got the mssql Object : %v/%v", req.Namespace, req.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MSSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myapi.MSSQL{}).
		Complete(r)
}
