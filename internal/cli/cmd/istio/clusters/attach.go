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

package clusters

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/peercluster"
	cmdCommon "github.com/banzaicloud/backyards-cli/internal/cli/cmd/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
	"github.com/banzaicloud/backyards-cli/pkg/k8s/resourcemanager"
	"github.com/banzaicloud/backyards-cli/pkg/monitoring"
	"github.com/banzaicloud/backyards-cli/pkg/nodeexporter"
)

type attachCommand struct {
	cli cli.CLI
}

type AttachOptions struct {
	name           string
	kubeconfigPath string
}

func NewAttachOptions() *AttachOptions {
	return &AttachOptions{}
}

func NewAttachCommand(cli cli.CLI, options *AttachOptions) *cobra.Command {
	c := &attachCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "attach [path-to-kubeconfig]",
		Args:  cobra.ExactArgs(1),
		Short: "Attach peer cluster to the mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			options.kubeconfigPath = args[0]

			return c.run(options)
		},
	}

	cmd.Flags().StringVar(&options.name, "name", options.name, "Name override for the peer cluster")

	return cmd
}

func (c *attachCommand) run(options *AttachOptions) error {
	client, err := cmdCommon.GetGraphQLClient(c.cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}
	defer client.Close()

	clusters, err := client.Clusters()
	if err != nil {
		return errors.WrapIf(err, "could not get clusters")
	}

	err = c.confirmKubeconfig(options.kubeconfigPath)
	if err != nil {
		return err
	}

	k8sconfig, err := k8sclient.GetConfigWithContext(options.kubeconfigPath, "")
	if err != nil {
		return errors.WrapIf(err, "could not get k8s config")
	}

	if options.name == "" {
		options.name, err = c.getClusterNameFromKubeconfig(options.kubeconfigPath)
		if err != nil {
			return err
		}
	}

	ok, _ := clusters.GetClusterByName(options.name)
	if ok {
		return errors.Errorf("peer cluster '%s' already exists", options.name)
	}

	ok, hostCluster := clusters.GetHostCluster()
	if !ok {
		return errors.New("host cluster not found")
	}

	k8sclient, err := k8sclient.NewClient(k8sconfig, k8sclient.Options{})
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}

	err = c.createServiceAccount(k8sclient, hostCluster.Namespace)
	if err != nil {
		return errors.WrapIf(err, "could not create service account on peer cluster")
	}

	s, err := c.getServiceAccountSecret(k8sclient, types.NamespacedName{
		Name:      "istio-operator",
		Namespace: hostCluster.Namespace,
	})
	if err != nil {
		return errors.WrapIf(err, "could not get secret for service account")
	}

	caData := s.Data["ca.crt"]
	if !bytes.Contains(caData, k8sconfig.CAData) {
		caData = append(append(caData, []byte("\n")...), k8sconfig.CAData...)
	}

	ok, err = client.AttachPeerCluster(graphql.AttachPeerClusterRequest{
		Name:                     options.name,
		URL:                      k8sconfig.Host,
		CertificateAuthorityData: base64.StdEncoding.EncodeToString(caData),
		ServiceAccountToken:      string(s.Data["token"]),
	})
	if err != nil {
		return errors.WrapIf(err, "could not attach peer cluster")
	}

	if ok && c.cli.Interactive() {
		log.Infof("attaching cluster '%s' is started successfully.\n", options.name)
	}

	log.Info("waiting for Istio sidecar injector to be available before install monitoring.")

	err = k8s.WaitForResourcesConditions(k8sclient, []k8s.NamespacedNameWithGVK{
		{
			NamespacedName: types.NamespacedName{
				Name:      "istio-sidecar-injector",
				Namespace: hostCluster.Namespace,
			},
			GroupVersionKind: appsv1.SchemeGroupVersion.WithKind("Deployment"),
		},
	}, wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    60,
	}, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	err = c.installMonitoring(k8sclient, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *attachCommand) installMonitoring(client k8sclient.Client, options *AttachOptions) error {
	labelManager := c.cli.LabelManager()
	m, err := monitoring.NewManager(resourcemanager.New(client, labelManager), c.cli.GetPersistentConfig().Namespace(), monitoring.Options{
		ClusterName: options.name,
	})
	if err != nil {
		return err
	}

	masterClient, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	m.Install()

	for _, r := range m.Resources() {
		if r.GroupVersionKind().GroupKind().Kind != "Service" {
			continue
		}
		labels := r.UnstructuredObject().GetLabels()
		if labels["backyards.banzaicloud.io/federated-prometheus"] != "true" {
			continue
		}
		err = k8s.ApplyResources(masterClient, labelManager, object.K8sObjects([]*object.K8sObject{r}))
		if err != nil {
			return err
		}
	}

	m.Install().Resources()
	err = m.Do()
	if err != nil {
		return err
	}

	neMgr, err := nodeexporter.NewNodeExporterManager(resourcemanager.New(client, labelManager), c.cli.GetPersistentConfig().Namespace())
	if err != nil {
		return err
	}
	err = neMgr.Install().Do()
	if err != nil {
		return err
	}

	return nil
}

func (c *attachCommand) confirmKubeconfig(kubeconfigPath string) error {
	config, err := k8sclient.GetRawConfig(kubeconfigPath, "")
	if err != nil {
		return errors.WrapIf(err, "could not get k8s config")
	}

	message := fmt.Sprintf("Are you sure to use the following context? %s (API Server: %s)",
		config.CurrentContext, config.Clusters[config.Contexts[config.CurrentContext].Cluster].Server)
	confirmed := false
	err = c.cli.IfConfirmed(message, func() error {
		confirmed = true

		return nil
	})
	if err != nil {
		return err
	}

	if !confirmed {
		return errors.New("refusing to use context")
	}

	return nil
}

func (c *attachCommand) getClusterNameFromKubeconfig(kubeconfigPath string) (string, error) {
	rawk8sconfig, err := k8sclient.GetRawConfig(kubeconfigPath, "")
	if err != nil {
		return "", errors.WrapIf(err, "could not get k8s config")
	}

	for name := range rawk8sconfig.Clusters {
		return name, nil
	}

	return "", errors.New("could not determine cluster name from kubeconfig")
}

func (c *attachCommand) createServiceAccount(k8sclient k8sclient.Client, namespace string) error {
	if c.cli.Interactive() {
		log.Info("creating service account and rbac permissions")
	}

	objects, err := c.getPeerClusterManifests(namespace)
	if err != nil {
		return errors.WrapIf(err, "could not get manifests for peer cluster setup")
	}
	objects.Sort(helm.InstallObjectOrder())

	err = k8s.ApplyResources(k8sclient, c.cli.LabelManager(), objects)
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
	}

	return nil
}

func (c *attachCommand) getServiceAccountSecret(k8sclient k8sclient.Client, namespacedName types.NamespacedName) (*corev1.Secret, error) {
	if c.cli.Interactive() {
		log.Info("retrieving service account token")
	}

	var sa corev1.ServiceAccount
	err := k8sclient.Get(context.Background(), namespacedName, &sa)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get service account")
	}

	if len(sa.Secrets) == 0 {
		return nil, errors.New("invalid service account")
	}

	var s corev1.Secret
	err = k8sclient.Get(context.Background(), types.NamespacedName{
		Name:      sa.Secrets[0].Name,
		Namespace: namespacedName.Namespace,
	}, &s)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get secret for service account")
	}

	return &s, nil
}

func (c *attachCommand) getPeerClusterManifests(namespace string) (object.K8sObjects, error) {
	objects, err := helm.Render(peercluster.Assets, "", helm.ReleaseOptions{
		Name:      "peer-cluster-manifests",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "peer-cluster-manifests")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}
