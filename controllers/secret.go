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
	"fmt"
	"github.com/pkg/errors"
	passgen "gomodules.xyz/password-generator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cu "kmodules.xyz/client-go/client"
	coreutil "kmodules.xyz/client-go/core/v1"
	metautil "kmodules.xyz/client-go/meta"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	msapi "kubedb.dev/mssql/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r MSSQLReconciler) ensureAuthSecret() error {
	if r.db.Spec.AuthSecret != nil && r.db.Spec.AuthSecret.ExternallyManaged {
		return r.ensureExternalAuthSecret()
	}
	if err := r.ensureInternalAuthSecret(); err != nil {
		return errors.Wrap(err, "failed to ensure database auth secret")
	}
	return nil
}

func (r MSSQLReconciler) ensureExternalAuthSecret() error {
	if r.db.Spec.AuthSecret.Name == "" {
		return fmt.Errorf("externally managed auth secret name is missing for MSSQL %s/%s", r.db.Namespace, r.db.Name)
	}
	// validate spec of the auth secret. make sure that have the keys 'username', 'password'
	var secret core.Secret
	err := r.Client.Get(r.ctx, types.NamespacedName{
		Name:      r.db.Spec.AuthSecret.Name,
		Namespace: r.db.Namespace,
	}, &secret)
	if err != nil {
		if kerr.IsNotFound(err) {
			return fmt.Errorf("externally managed auth secret \"%s\" not found for MSSQL %s/%s", r.db.Spec.AuthSecret.Name, r.db.Namespace, r.db.Name)
		}
		return err
	}
	return r.validateAuthSecret(&secret)
}

//  if internally managed :
//    if secret not exists (name either defaulted or user provided)
//      - create a secret
//    else
//      - check the secret labels are not associated with a different db
//      - validate required keys

func (r *MSSQLReconciler) ensureInternalAuthSecret() error {
	secretName := r.db.GetAuthSecretName()
	var secret core.Secret
	err := r.Client.Get(r.ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: r.db.Namespace,
	}, &secret)
	if err != nil && !kerr.IsNotFound(err) {
		return err
	} else if kerr.IsNotFound(err) {
		if err = r.createAuthSecret(); err != nil {
			return err
		}
	} else {
		// secret exists but labels indicate different db
		if secret.Labels[metautil.NameLabelKey] != r.db.ResourceFQN() ||
			secret.Labels[metautil.InstanceLabelKey] != r.db.Name {
			return fmt.Errorf(`auth secret "%v/%v" associated with %s %s but expected to be associated with %s %s`,
				r.db.Namespace,
				secretName,
				secret.Labels[metautil.NameLabelKey],
				secret.Labels[metautil.InstanceLabelKey],
				r.db.ResourceFQN(),
				r.db.Name,
			)
		}
		if err := r.validateAuthSecret(&secret); err != nil {
			return err
		}
	}

	ms, _, err := cu.CreateOrPatch(r.ctx, r.Client, &msapi.MSSQL{
		ObjectMeta: r.db.ObjectMeta,
	}, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*msapi.MSSQL)
		if in.Spec.AuthSecret == nil {
			in.Spec.AuthSecret = &dbapi.SecretReference{}
		}
		in.Spec.AuthSecret.Name = secretName
		return in
	})
	if err != nil {
		return err
	}
	r.db.Spec.AuthSecret = ms.(*msapi.MSSQL).Spec.AuthSecret
	return nil
}

func (r *MSSQLReconciler) createAuthSecret() error {
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.db.GetAuthSecretName(),
			Namespace: r.db.Namespace,
			Labels:    r.db.OffshootLabels(),
		},
		Type: core.SecretTypeBasicAuth,
		Data: map[string][]byte{
			core.BasicAuthUsernameKey: []byte(msapi.MSSQLUser),
			core.BasicAuthPasswordKey: []byte(passgen.Generate(dbapi.DefaultPasswordLength)),
		},
	}
	_, _, err := cu.CreateOrPatch(r.ctx, r.Client, secret, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.Secret)
		coreutil.EnsureOwnerReference(&in.ObjectMeta, r.getOwnerRef())
		return in
	})
	return err
}

func (r *MSSQLReconciler) validateAuthSecret(secret *core.Secret) error {
	// verify if the desired key ["password", "username"] exist or not (when secret is managed by the user)
	if _, ok := secret.Data[core.BasicAuthUsernameKey]; !ok {
		return fmt.Errorf("key \"%s\" doesn't exists inside spec data for secret %s/%s", core.BasicAuthUsernameKey, secret.Namespace, secret.Name)
	}
	if _, ok := secret.Data[core.BasicAuthPasswordKey]; !ok {
		return fmt.Errorf("key \"%s\" doesn't exists inside spec data for secret %s/%s", core.BasicAuthPasswordKey, secret.Namespace, secret.Name)
	}
	return nil
}
