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

package controller

import (
	"context"
	"fmt"

	km "github.com/k0sproject/k0smotron/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ClusterReconciler) generateCM(kmc *km.Cluster) v1.ConfigMap {
	// TODO externalAddress cannot be hardcoded
	// TODO k0s.yaml should probably be a
	// github.com/k0sproject/k0s/pkg/apis/k0s.k0sproject.io/v1beta1.ClusterConfig
	// and then unmarshalled into json to make modification of fields reliable
	cm := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kmc.GetConfigMapName(),
			Namespace: kmc.Namespace,
		},
		Data: map[string]string{
			"k0s.yaml": fmt.Sprintf(`apiVersion: k0s.k0sproject.io/v1beta1
kind: ClusterConfig
metadata:
  name: k0s
spec:
  api:
    externalAddress: %s
    port: %d
  konnectivity:
    agentPort: %d`, kmc.Spec.ExternalAddress, kmc.Spec.Service.APIPort, kmc.Spec.Service.KonnectivityPort), // TODO: do it as a template or something like this
		},
	}

	_ = ctrl.SetControllerReference(kmc, &cm, r.Scheme)
	return cm
}

func (r *ClusterReconciler) reconcileCM(ctx context.Context, kmc km.Cluster) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling configmap")

	if kmc.Spec.Service.Type == v1.ServiceTypeLoadBalancer && kmc.Spec.ExternalAddress == "" {
		return nil
	}

	cm := r.generateCM(&kmc)

	return r.Client.Patch(ctx, &cm, client.Apply, patchOpts...)
}
