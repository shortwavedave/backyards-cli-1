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

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"istio.io/operator/pkg/object"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/backyards"
	kafkav1beta1 "github.com/banzaicloud/backyards-cli/internal/apis/kafka-operator/v1beta1"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/kafka"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	requirementNotFoundErrorTemplate = "Unable to install Backyards: %s\n"
	defaultReleaseName               = "backyards"
)

var (
	certManagerPodLabels = map[string]string{
		"app": "cert-manager",
	}
)

type installCommand struct {
	cli                      cli.CLI
	shouldInstallIstio       bool
	shouldInstallCanary      bool
	shouldInstallCertManager bool
	shouldInstallKafka       bool
	shouldRunDemo            bool
}

type InstallOptions struct {
	releaseName    string
	istioNamespace string
	dumpResources  bool

	enableAuditSink   bool
	anonymousAuth     bool
	installEverything bool
	runDemo           bool

	apiImage       string
	webImage       string
	trackingIDFunc func() string

	installKafka bool
}

// patchStringValue specifies a patch operation for a string value
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func NewInstallOptions(trackingIDFunc func() string) *InstallOptions {
	return &InstallOptions{
		trackingIDFunc: trackingIDFunc,
	}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:         "install [flags]",
		Args:        cobra.NoArgs,
		Short:       "Install Backyards",
		Annotations: map[string]string{util.CommandGroupAnnotationKey: util.InstallCommand},
		Long: `Installs Backyards.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.

The command can install every component at once with the '--install-everything' option.`,
		Example: `  # Default install.
  backyards install

  # Install Backyards into a non-default namespace.
  backyards install -n backyards-system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err = c.shouldInstallComponents(options)
			if err != nil {
				return err
			}

			err = c.runSubcommands(options)
			if err != nil {
				return err
			}

			err = c.run(options)
			if err != nil {
				return err
			}

			err = c.runDemoInstall(options)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", defaultReleaseName, "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", istio.DefaultNamespace, "Namespace of Istio sidecar injector")

	cmd.Flags().BoolVarP(&options.installEverything, "install-everything", "a", options.installEverything, "Install all required components at once")
	cmd.Flags().BoolVar(&options.runDemo, "run-demo", options.runDemo, "Install demo application, send load and open up the dashboard")

	cmd.Flags().BoolVar(&options.enableAuditSink, "enable-auditsink", options.enableAuditSink, "Enable deploying the auditsink service and sending audit logs over http")
	cmd.Flags().BoolVar(&options.anonymousAuth, "anonymous-auth", options.anonymousAuth, "Switch to anonymous mode")

	cmd.Flags().StringVar(&options.apiImage, "api-image", options.apiImage, "Image for the API")
	cmd.Flags().StringVar(&options.webImage, "web-image", options.webImage, "Image for the frontend")

	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", options.dumpResources, "Dump resources to stdout instead of applying them")

	cmd.Flags().BoolVar(&options.installKafka, "with-kafka-cluster", options.installKafka, "Enable deploying a 3 broker kafka cluster with zookeeper")

	return cmd
}

func (c *installCommand) run(options *InstallOptions) error {
	err := c.validate(options)
	if err != nil {
		errors := multierr.Errors(err)
		var errorItems string
		for _, e := range errors {
			errorItems += "\n - " + e.Error()
		}
		fmt.Fprintf(os.Stderr, requirementNotFoundErrorTemplate, errorItems)
		return nil
	}

	values, err := getValues(options.releaseName, options.istioNamespace, func(values *Values) {
		if options.enableAuditSink {
			values.AuditSink.Enabled = true
		}

		if c.shouldCertManagerBeEnabled(options) {
			values.CertManager.Enabled = true
		}
		if options.anonymousAuth {
			values.Auth.Mode = anonymous
			values.Impersonation.Enabled = false
		} else {
			values.Auth.Mode = impersonation
			values.Impersonation.Enabled = true
		}
		if options.apiImage != "" {
			imageParts := strings.Split(options.apiImage, ":")
			values.Application.Image.Repository = imageParts[0]
			if len(imageParts) > 1 {
				values.Application.Image.Tag = imageParts[1]
			} else {
				values.Application.Image.Tag = "latest"
			}
		}
		if options.webImage != "" {
			imageParts := strings.Split(options.webImage, ":")
			values.Web.Image.Repository = imageParts[0]
			if len(imageParts) > 1 {
				values.Web.Image.Tag = imageParts[1]
			} else {
				values.Web.Image.Tag = "latest"
			}
		}
		if options.trackingIDFunc != nil {
			values.Web.Env["GOOGLE_ANALYTICS_ID"] = options.trackingIDFunc()
		}
	})
	if err != nil {
		return err
	}

	err = c.setTracingAddress(values)
	if err != nil {
		return err
	}

	objects, err := getBackyardsObjects(values, c.cli)
	if err != nil {
		return err
	}

	objects.Sort(helm.InstallObjectOrder())

	if !options.dumpResources {
		client, err := c.cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResources(client, c.cli.LabelManager(), objects)
		if err != nil {
			return err
		}

		err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objects), wait.Backoff{
			Duration: time.Second * 5,
			Factor:   1,
			Jitter:   0,
			Steps:    24,
		}, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
		if err != nil {
			return err
		}
	} else {
		yaml, err := objects.YAMLManifest()
		if err != nil {
			return err
		}
		fmt.Fprintf(c.cli.Out(), yaml)
	}

	return nil
}

func getValues(releaseName, istioNamespace string, valueOverrideFunc func(values *Values)) (Values, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(backyards.Chart)
	if err != nil {
		return Values{}, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return Values{}, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.SetDefaults(releaseName, istioNamespace)

	if valueOverrideFunc != nil {
		valueOverrideFunc(&values)
	}

	return values, nil
}

func getBackyardsObjects(values Values, cli cli.CLI) (object.K8sObjects, error) {
	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(backyards.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "backyards",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: cli.GetPersistentConfig().Namespace(),
	}, "backyards")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func (c *installCommand) setTracingAddress(values Values) error {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}

	istioCR, err := istio.FetchIstioCR(cl)
	if err != nil {
		return err
	}

	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/tracing/zipkin/address",
		Value: fmt.Sprintf("%s.%s:%d", values.Tracing.Service.Name, c.cli.GetPersistentConfig().Namespace(), values.Tracing.Service.ExternalPort),
	}}
	payloadBytes, _ := json.Marshal(payload)

	err = cl.Patch(context.Background(), &istioCR, client.ConstantPatch(types.JSONPatchType, payloadBytes))
	if err != nil {
		return errors.WrapIf(err, "could not patch Istio CR")
	}

	return nil
}

func (c *installCommand) validate(options *InstallOptions) error {
	var istioHealthy bool
	var combinedErr error

	istioExists, istioHealthy, err := c.istioRunning(options.istioNamespace)
	if err != nil {
		return errors.WrapIf(err, "failed to check Istio state")
	}

	if !istioExists {
		combinedErr = errors.Combine(combinedErr,
			errors.Errorf("could not find Istio sidecar injector in '%s' namespace, "+
				"use the --install-istio flag", options.istioNamespace))
	}
	if istioExists && !istioHealthy {
		combinedErr = errors.Combine(combinedErr,
			errors.Errorf("Istio sidecar injector not healthy yet in '%s' namespace", options.istioNamespace))
	}

	if c.shouldCertManagerBeEnabled(options) {
		certManagerExists, certManagerHealthy, err := c.certManagerRunning()
		if err != nil {
			return errors.WrapIf(err, "failed to check cert-manager state")
		}

		if !certManagerExists {
			combinedErr = errors.Combine(combinedErr,
				errors.Errorf("could not find cert-manager controller in '%s' namespace, "+
					"use the --install-cert-manager flag or disable it using --disable-cert-manager "+
					"which disables dependent services as well", certmanager.CertManagerNamespace))
		}
		if certManagerExists && !certManagerHealthy {
			combinedErr = errors.Combine(combinedErr,
				errors.Errorf("cert-manager controller not healthy yet in '%s' namespace", certmanager.CertManagerNamespace))
		}
	}

	return combinedErr
}

func (c *installCommand) istioRunning(istioNamespace string) (exists bool, healthy bool, err error) {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		err = errors.WrapIf(err, "could not get k8s client")
		return
	}
	var pods v1.PodList
	err = cl.List(context.Background(), &pods, client.InNamespace(istioNamespace), client.MatchingLabels(util.SidecarPodLabels))
	if err != nil {
		err = errors.WrapIf(err, "could not list istio pods")
		return
	}
	if len(pods.Items) > 0 {
		exists = true
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			healthy = true
			break
		}
	}
	return
}

func (c *installCommand) kafkaInstalled() (exists bool, healthy bool, err error) {
	kafkaNamespace := kafka.GetNamespace()

	cl, err := c.cli.GetK8sClient()
	if err != nil {
		err = errors.WrapIf(err, "could not get k8s client")
		return
	}
	var clusters kafkav1beta1.KafkaClusterList
	err = cl.List(context.Background(), &clusters, client.InNamespace(kafkaNamespace))
	if err != nil {
		err = errors.WrapIf(err, "could not list kafka clusters")
		return
	}
	if len(clusters.Items) > 0 {
		exists = true
	}
	for _, cluster := range clusters.Items {
		if cluster.Status.State == kafkav1beta1.KafkaClusterRunning {
			healthy = true
			break
		}
	}
	return
}

func (c *installCommand) certManagerRunning() (exists bool, healthy bool, err error) {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		err = errors.WrapIf(err, "could not get k8s client")
		return
	}
	var certManagerPods v1.PodList
	err = cl.List(context.Background(), &certManagerPods, client.InNamespace(certmanager.CertManagerNamespace),
		client.MatchingLabels(certManagerPodLabels))
	if err != nil {
		err = errors.WrapIf(err, "failed to list cert-manager controller pods")
		return
	}
	if len(certManagerPods.Items) > 0 {
		exists = true
	}
	for _, pod := range certManagerPods.Items {
		if pod.Status.Phase == v1.PodRunning {
			healthy = true
			break
		}
	}
	return
}

func (c *installCommand) shouldInstallComponents(options *InstallOptions) error {
	installIstioInteractively := false

	shouldAskComponents := !options.runDemo && !options.installEverything

	kafkaExists, _, err := c.kafkaInstalled()
	if err != nil {
		return errors.WrapIf(err, "failed to check Istio state")
	}

	installKafkaInteractively := false
	if !kafkaExists && !options.installKafka && shouldAskComponents && c.cli.InteractiveTerminal() {
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  "Install kafka cluster (recommended). Press enter to accept",
			Default:  true,
		}, &installKafkaInteractively)
		if err != nil {
			return err
		}
	}
	c.shouldInstallKafka = !kafkaExists && (options.installKafka || options.installEverything || installKafkaInteractively)

	istioExists, istioHealthy, err := c.istioRunning(options.istioNamespace)
	if err != nil {
		return errors.WrapIf(err, "failed to check Istio state")
	}

	if (!istioExists || !istioHealthy) && !c.shouldInstallKafka && shouldAskComponents && c.cli.InteractiveTerminal() {
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  "Install istio-operator (recommended). Press enter to accept",
			Default:  true,
		}, &installIstioInteractively)
		if err != nil {
			return err
		}
	}
	c.shouldInstallIstio = options.installEverything || installIstioInteractively || c.shouldInstallKafka

	if c.shouldCertManagerBeEnabled(options) {
		installCertManagerInteractively := false
		if !c.shouldInstallKafka && shouldAskComponents && c.cli.InteractiveTerminal() {
			err := survey.AskOne(&survey.Confirm{
				Renderer: survey.Renderer{},
				Message:  "Install cert-manager (recommended). Press enter to accept",
				Default:  true,
			}, &installCertManagerInteractively)
			if err != nil {
				return err
			}
		}
		c.shouldInstallCertManager = options.installEverything || installCertManagerInteractively
	}

	installCanaryInteractively := false
	if shouldAskComponents && c.cli.InteractiveTerminal() {
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  "Install canary-operator (recommended). Press enter to accept",
			Default:  true,
		}, &installCanaryInteractively)
		if err != nil {
			return err
		}
	}
	c.shouldInstallCanary = options.installEverything || installCanaryInteractively

	runDemoInteractively := false
	if shouldAskComponents && c.cli.InteractiveTerminal() {
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  "Install and run demo application (optional). Press enter to skip",
			Default:  false,
		}, &runDemoInteractively)
		if err != nil {
			return err
		}
	}
	c.shouldRunDemo = options.runDemo || runDemoInteractively

	return nil
}

func (c *installCommand) runSubcommands(options *InstallOptions) error {
	var err error
	var scmd *cobra.Command

	if c.shouldInstallIstio {
		scmdOptions := istio.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = istio.NewInstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Istio mesh install")
		}
	}

	if c.shouldInstallCertManager {
		scmdOptions := certmanager.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = certmanager.NewInstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during cert-manager install")
		}
	}

	if c.shouldInstallKafka {
		scmdOptions := kafka.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = kafka.NewInstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during kafka install")
		}
	}

	if c.shouldInstallCanary {
		scmdOptions := canary.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = canary.NewInstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Canary feature install")
		}
	}

	return nil
}

func (c *installCommand) runDemoInstall(options *InstallOptions) error {
	var err error
	var scmd *cobra.Command

	if c.shouldRunDemo {
		scmdOptions := demoapp.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = demoapp.NewInstallCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during demo application install")
		}
	}

	if c.shouldRunDemo {
		scmdOptions := demoapp.NewLoadOptions()
		scmdOptions.Nowait = true
		scmd := demoapp.NewLoadCommand(c.cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during sending load to demo application")
		}

		dbOptions := NewDashboardOptions()
		dbOptions.QueryParams["namespaces"] = demoapp.GetNamespace()
		dbCmd := NewDashboardCommand(c.cli, dbOptions)
		err = dbCmd.RunE(dbCmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during opening dashboard")
		}
	}

	return nil
}

func (c *installCommand) shouldCertManagerBeEnabled(options *InstallOptions) bool {
	return options.enableAuditSink || c.shouldInstallKafka
}
