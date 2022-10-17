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
	"github.com/go-logr/logr"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	msapi "kubedb.dev/mssql/api/v1alpha1"
)

// MSSQLReconciler reconciles a MSSQL object
type MSSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	ctx    context.Context
	Log    logr.Logger
	db     *msapi.MSSQL
}

//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microsoft.kubedb.com,resources=mssqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services;secrets,verbs=get;list;watch;create;patch;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;patch;update;delete

func (r *MSSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.ctx = ctx
	r.Log = log.FromContext(ctx)

	mssql, err := r.getMSSQL(req.NamespacedName)
	if err != nil {
		if kerr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return r.requeueWithError("Failed to get MSSQL", err)
	}
	r.db = mssql
	klog.Infof("Got the mssql Object : %v/%v", r.db.Namespace, r.db.Name)

	// if MSSQL instance is marked for deletion, remove the finalizers & abort reconcile
	if r.isMarkedForDeletion() {
		err = r.removeFinalizers()
		if err != nil {
			return r.requeueWithError("Failed to remove finalizers", err)
		}
		return ctrl.Result{}, nil
	}
	err = r.ensureFinalizers()
	if err != nil {
		return r.requeueWithError("Failed to ensure finalizers", err)
	}

	//TODO:
	// Update MSSQL phase from current conditions
	// r.updatePhaseFromCondition(ctx, r.db)

	err = r.ensurePrimaryService()
	if err != nil {
		return r.requeueWithError("Failed to ensure service", err)
	}

	err = r.ensureGoverningServices()
	if err != nil {
		return r.requeueWithError("Failed to ensure service", err)
	}

	err = r.ensureAuthSecret()
	if err != nil {
		return r.requeueWithError("Failed to ensure secrets", err)
	}

	err = r.ensureNodes()
	if err != nil {
		return r.requeueWithError("Failed to ensure nodes", err)
	}

	return ctrl.Result{}, nil
}

func (r *MSSQLReconciler) getMSSQL(meta types.NamespacedName) (*msapi.MSSQL, error) {
	var db msapi.MSSQL
	err := r.Client.Get(context.TODO(), meta, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MSSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&msapi.MSSQL{}).
		Complete(r)
}
