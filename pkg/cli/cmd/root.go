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
	"fmt"
	"os"
	"regexp"

	"emperror.dev/errors"
	logrushandler "emperror.dev/handler/logrus"
	"github.com/AlecAivazis/survey/v2"
	ga "github.com/jpillora/go-ogle-analytics"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/config"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/login"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/graph"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

var (
	kubeconfigPath string
	kubeContext    string
	verbose        bool
	outputFormat   string
	trackingID     string
)

// RootCmd represents the root Cobra command
var RootCmd = &cobra.Command{
	Use:           "backyards",
	Short:         "Install and manage Backyards",
	SilenceErrors: true,
	SilenceUsage:  false,
}

// Init is a temporary function to set initial values in the root cmd
func Init(version string, commitHash string, buildDate string, tid string) {
	RootCmd.Version = version

	trackingID = tid

	RootCmd.SetVersionTemplate(fmt.Sprintf(
		"Backyards CLI version %s (%s) built on %s\n",
		version,
		commitHash,
		buildDate,
	))
	RootCmd.SetUsageFunc(usageFunc)
}

// GetRootCommand returns the cli root command
func GetRootCommand() *cobra.Command {
	return RootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately
// This is called by main.main(). It only needs to happen once to the RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		handler := logrushandler.New(log.New())
		handler.Handle(err)
		os.Exit(1)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()

	flags.StringVarP(&kubeconfigPath, "kubeconfig", "c", "", "path to the kubeconfig file to use for CLI requests")
	_ = viper.BindPFlag("kubeconfig", flags.Lookup("kubeconfig"))
	flags.StringVar(&kubeContext, "context", "", "name of the kubeconfig context to use")
	_ = viper.BindPFlag("kubecontext", flags.Lookup("context"))
	flags.BoolVarP(&verbose, "verbose", "v", false, "turn on debug logging")
	_ = viper.BindPFlag("verbose", flags.Lookup("verbose"))

	flags.StringVarP(&outputFormat, "output", "o", "table", "output format (table|yaml|json)")
	_ = viper.BindPFlag("output.format", flags.Lookup("output"))

	flags.Bool("formatting.force-color", false, "force color even when non in a terminal")
	flags.Bool("color", true, "use colors on non-tty outputs")
	_ = viper.BindPFlag("formatting.force-color", flags.Lookup("color"))
	flags.Bool("non-interactive", false, "never ask questions interactively")
	_ = viper.BindPFlag("formatting.non-interactive", flags.Lookup("non-interactive"))
	flags.Bool("interactive", false, "ask questions interactively even if stdin or stdout is non-tty")
	_ = viper.BindPFlag("formatting.force-interactive", flags.Lookup("interactive"))

	flags.String("persistent-config-file", "", "Backyards persistent config file to use instead of the default at ~/.banzai/backyards/")
	_ = viper.BindPFlag(cli.PersistentConfigKey, flags.Lookup("persistent-config-file"))

	settings := cli.PersistentConfigurationSettings
	settings.Configure(flags)

	cliRef := cli.NewCli(os.Stdout, RootCmd, settings)

	RootCmd.AddCommand(cmd.NewVersionCommand(cliRef))
	RootCmd.AddCommand(cmd.NewInstallCommand(cliRef))
	RootCmd.AddCommand(cmd.NewUninstallCommand(cliRef))
	RootCmd.AddCommand(cmd.NewDashboardCommand(cliRef, cmd.NewDashboardOptions()))
	RootCmd.AddCommand(istio.NewRootCmd(cliRef))
	RootCmd.AddCommand(canary.NewRootCmd(cliRef))
	RootCmd.AddCommand(demoapp.NewRootCmd(cliRef))
	RootCmd.AddCommand(routing.NewRootCmd(cliRef))
	RootCmd.AddCommand(certmanager.NewRootCmd(cliRef))
	RootCmd.AddCommand(graph.NewGraphCmd(cliRef, "base.json"))
	RootCmd.AddCommand(login.NewLoginCmd(cliRef))
	RootCmd.AddCommand(config.NewConfigCmd(cliRef))
	RootCmd.AddCommand(sidecarproxy.NewRootCmd(cliRef))

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		err := cliRef.Initialize()
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		namespaceRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
		namespace := cliRef.GetPersistentConfig().Namespace()
		if !namespaceRegex.MatchString(namespace) {
			return errors.NewWithDetails("invalid namespace", "namespace", namespace)
		}

		err = askLicense(cliRef)
		if err != nil {
			return err
		}
		config := cliRef.GetPersistentConfig()
		collectMetrics(cmd, config)
		return nil
	}

	RootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		config := cliRef.GetPersistentConfig()
		return config.PersistConfig()
	}
}

func collectMetrics(cmd *cobra.Command, config cli.PersistentConfig) {
	if config.TrackingClientID() == "" {
		config.SetTrackingClientID(uuid.New())
	}
	if trackingID != "" {
		log.Debugf("sending metrics to %s", trackingID)
		client, err := ga.NewClient(trackingID)
		if err != nil {
			log.Errorf("Failed to send metrics")
			log.Debugf("Failed to configure analytics client: %+v", err)
			return
		}
		client.UserAgentOverride("backyards-cli")
		event := ga.NewEvent("clicmd", cmd.CommandPath())
		event.Label(RootCmd.Version)
		client.AnonymizeIP(true)
		client.ClientID(config.TrackingClientID())
		err = client.Send(event)
		if err != nil {
			log.Errorf("Failed to send metrics")
			log.Debugf("Failed to send event: %+v", err)
			return
		}
	}
}

func askLicense(cliRef cli.CLI) error {
	if !cliRef.InteractiveTerminal() && !cliRef.GetPersistentConfig().LicenseAccepted() {
		return errors.New("you have to accept the license to use this product")
	}

	if cliRef.InteractiveTerminal() && !cliRef.GetPersistentConfig().LicenseAccepted() {
		accepted := false
		err := survey.AskOne(&survey.Confirm{
			Renderer: survey.Renderer{},
			Message:  "Are you cool with the license terms?",
			Default:  true,
		}, &accepted)
		if err != nil {
			return err
		}
		if !accepted {
			return errors.New("you have to accept the license to use this product")
		}
		cliRef.GetPersistentConfig().SetLicenseAccepted(accepted)
	}

	return cliRef.GetPersistentConfig().PersistConfig()
}
