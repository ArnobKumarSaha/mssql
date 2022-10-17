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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientutil "kmodules.xyz/client-go/client"
	coreutil "kmodules.xyz/client-go/core/v1"
	api "kubedb.dev/apimachinery/apis"
	msapi "kubedb.dev/mssql/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MSSQLReconciler) ensureFinalizers() error {
	if !coreutil.HasFinalizer(r.db.ObjectMeta, api.Finalizer) {
		_, _, err := clientutil.CreateOrPatch(r.ctx, r.Client, &msapi.MSSQL{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.db.Name,
				Namespace: r.db.Namespace,
			},
		}, func(obj client.Object, createOp bool) client.Object {
			ed := obj.(*msapi.MSSQL)
			ed.ObjectMeta = coreutil.AddFinalizer(ed.ObjectMeta, api.Finalizer)
			return obj
		})
		if err != nil {
			return err
		}
		r.Log.Info("Added Finalizers")
	}

	return nil
}

func (r *MSSQLReconciler) removeFinalizers() error { // call it, only if MSSQL is marked for deletion
	if coreutil.HasFinalizer(r.db.ObjectMeta, api.Finalizer) {
		_, _, err := clientutil.CreateOrPatch(r.ctx, r.Client, &msapi.MSSQL{
			ObjectMeta: r.db.ObjectMeta,
		}, func(obj client.Object, createOp bool) client.Object {
			ed := obj.(*msapi.MSSQL)
			ed.ObjectMeta = coreutil.RemoveFinalizer(ed.ObjectMeta, api.Finalizer)
			return obj
		})
		if err != nil {
			return err
		}
		r.Log.Info("Removed Finalizers")
	}
	return nil
}

// requeueWithError is a wrapper around logging an error message
// then passes the error through to the controller manager
func (r *MSSQLReconciler) requeueWithError(msg string, err error) (ctrl.Result, error) {
	r.Log.Error(err, msg)
	return ctrl.Result{}, err
}

func (r *MSSQLReconciler) isMarkedForDeletion() bool {
	return !r.db.GetDeletionTimestamp().IsZero()
}

func (r *MSSQLReconciler) getOwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(r.db, msapi.GroupVersion.WithKind(msapi.ResourceKindMSSQL))
}
