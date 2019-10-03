// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certmanager

import (
	"context"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

const (
	CertManagerNamespace   = "cert-manager"
	certManagerReleaseName = "cert-manager"
)

func NewRootCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert-manager",
		Short: "Install and manage cert-manager",
	}

	cmd.AddCommand(
		NewInstallCommand(cli, NewInstallOptions()),
		NewUninstallCommand(cli, NewUninstallOptions()),
	)

	return cmd
}

func GetNamespace() string {
	return CertManagerNamespace
}

func crdExists(cli cli.CLI, crdName string) (bool, error) {
	cl, err := cli.GetK8sClient()
	if err != nil {
		return false, err
	}

	var crd apiextensions.CustomResourceDefinition

	err = cl.Get(context.Background(), types.NamespacedName{
		Name: crdName,
	}, &crd)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.WrapIfWithDetails(err, "could not get CRD", "name", crdName)
	}

	return true, nil
}
