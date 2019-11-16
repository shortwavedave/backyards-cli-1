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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	ga "github.com/jpillora/go-ogle-analytics"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/licenses"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/config"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/graph"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/login"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/sidecarproxy"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

var (
	kubeconfigPath string
	kubeContext    string
	verbose        bool
	outputFormat   string
	trackingID     *string
)

const (
	GAUserAgent       = "backyards-cli"
	GACategoryCommand = "clicmd"
	GACategoryLicense = "clilicense"
	GAActionReject    = "reject"

	AcceptLicense = "accept-license"
	LicenseLink   = "https://banzaicloud.com/docs/backyards/evaluation-license"
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

	trackingID = &tid

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
		if verbose {
			fmt.Fprintf(RootCmd.ErrOrStderr(), "Error with stack:\n%+v\n\n", err)
			if len(errors.GetDetails(err)) > 0 {
				fmt.Fprintf(RootCmd.ErrOrStderr(), "Details:\n%+v\n", errors.GetDetails(err))
			}
		} else {
			if len(errors.GetErrors(err)) == 1 {
				fmt.Fprintf(RootCmd.ErrOrStderr(), "%s\n", err)
			} else {
				for _, err := range errors.GetErrors(err) {
					fmt.Fprintf(RootCmd.ErrOrStderr(), "- %s\n", err)
				}
			}
		}
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

	flags.Bool("accept-license", false, fmt.Sprintf("Accept the license: %s", LicenseLink))

	cli.PersistentSettings.Configure(flags)
	cli.PersistentGlobalSettings.Configure(flags)

	cliRef := cli.NewCli(os.Stdout, RootCmd)

	RootCmd.AddCommand(cmd.NewVersionCommand(cliRef))
	RootCmd.AddCommand(cmd.NewInstallCommand(cliRef, cmd.NewInstallOptions(func() string {
		return *trackingID
	})))
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
	RootCmd.AddCommand(cmd.NewLicenseCommand(cliRef))
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
		return nil
	}

	RootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		sendGAEvent(cliRef, ga.NewEvent(GACategoryCommand, cmd.CommandPath()))
		config := cliRef.GetPersistentConfig()
		return config.PersistConfig()
	}
}

func askLicense(cliRef cli.CLI) error {
	if ok, _ := cliRef.GetRootCommand().Flags().GetBool(AcceptLicense); ok {
		return nil
	}

	licenseAccepted := cliRef.GetPersistentGlobalConfig().LicenseAcceptedForVersion(cmd.LicenseVersion)

	if !cliRef.InteractiveTerminal() && !licenseAccepted {
		return errors.New("you have to accept the license to use this product")
	}

	if cliRef.InteractiveTerminal() && !licenseAccepted {
		fullLicenseHaveBeenRead := false
		for {
			const (
				AnswerYes     = "Yes"
				AnswerNo      = "No"
				AnswerLicense = "Read the license"
			)

			fmt.Fprint(cliRef.Out(), heredoc.Doc(`
				You have successfully downloaded and installed Backyards, a Banzai Cloud product. Before commencing the use of the product, You must read, acknowledge and agree to the Banzai Cloud Evaluation License, which is available at https://banzaicloud.com/docs/backyards/evaluation-license/. If you would like to use the Banzai Cloud product in production, you will need a different license. Contact: sales@banzaicloud.com. 
				
				For your convenience here is a short summary of what you will agree to by commencing use of the Evaluation Product: 
				
				(a) You receive a limited license to test and evaluate the Evaluation Product;
				
				(b) You receive the Evaluation Product “as-is” and “as-available”, without any warranties of title, non-infringement or fitness for purpose;
				
				(c) Banzai Cloud disclaims and where disclaiming is not possible limits its liability for damages incurred by You or any third party;
				
				(d) any feedback on the Evaluation Product given by You will be used by Banzai Cloud freely and independently for any purpose.
				
				(e) Banzai Cloud may collect anonymized analytics data of your usage and use it for any purpose, including enhancements, statistics and business purposes. 
				
				This is a human-readable summary of (and not a substitute for) the Agreement https://banzaicloud.com/docs/backyards/evaluation-license/.

				Please read the full license and confirm that you accept the terms!
			`))

			response := ""
			err := survey.AskOne(&survey.Select{
				Renderer: survey.Renderer{},
				Options:  []string{AnswerNo, AnswerYes, AnswerLicense},
				Default:  AnswerNo,
			}, &response, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
			switch response {
			case AnswerNo:
				sendGAEvent(cliRef, ga.NewEvent(GACategoryLicense, GAActionReject))
				return errors.New("you have to accept the license to use this product")
			case AnswerYes:
				if fullLicenseHaveBeenRead {
					cliRef.GetPersistentGlobalConfig().SetLicenseAcceptedForVersion(cmd.LicenseVersion)
					return cliRef.GetPersistentGlobalConfig().PersistConfig()
				}
				moveOn := false
				err := survey.AskOne(&survey.Confirm{
					Renderer: survey.Renderer{},
					Message:  "You haven't read the complete license! Continue?",
					Default:  true,
				}, &moveOn)
				if err != nil {
					return err
				}
				if moveOn {
					cliRef.GetPersistentGlobalConfig().SetLicenseAcceptedForVersion(cmd.LicenseVersion)
					return cliRef.GetPersistentGlobalConfig().PersistConfig()
				}
			case AnswerLicense:
				if err := readLicense(); err != nil {
					return err
				}
				fullLicenseHaveBeenRead = true
			}
		}
	}

	return nil
}

func sendGAEvent(cli cli.CLI, event *ga.Event) {
	config := cli.GetPersistentConfig()
	if config.TrackingClientID() == "" {
		config.SetTrackingClientID(uuid.New())
	}
	if *trackingID != "" {
		log.Debugf("Sending metrics to %s", *trackingID)
		client, err := ga.NewClient(*trackingID)
		if err != nil {
			log.Debugf("Failed to configure analytics client: %+v", err)
			return
		}
		client.HttpClient.Timeout = time.Millisecond * 500
		client.UserAgentOverride(GAUserAgent)
		event.Label(RootCmd.Version)
		client.AnonymizeIP(true)
		client.ClientID(config.TrackingClientID())
		err = client.Send(event)
		if err != nil {
			log.Debugf("Failed to send event: %+v", err)
			return
		}
	}
}

func readLicense() error {
	f, err := licenses.Licenses.Open(fmt.Sprintf("LICENSE-v%s.txt", cmd.LicenseVersion))
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\n\nPress enter to continue", string(b))
	_, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	return nil
}
