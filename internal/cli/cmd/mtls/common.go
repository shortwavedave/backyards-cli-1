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

package mtls

import (
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-client-go/pkg/authentication/v1alpha1"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/util"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

const (
	meshWidePolicy        = "mesh"
	meshNotSupportedError = "operation not supported on mesh wide policy"
)

type mTLSOptions struct {
	resourceID   string
	resourceName types.NamespacedName
	portName     string
	portNumber   int
}

func newMTLSOptions() *mTLSOptions {
	return &mTLSOptions{}
}

func parseMTLSArgs(options *mTLSOptions, args []string, cli cli.CLI, meshSupported bool, portsSupported bool) error {
	var err error

	if len(args) > 0 {
		options.resourceID = args[0]
	}

	if options.resourceID == "" {
		return errors.New("resource must be specified")
	}

	if options.resourceID != meshWidePolicy && !isAutoMTLSEnabled(cli) {
		return errors.New("auto mTLS needs to be enabled to use this feature")
	}

	switch {
	case options.resourceID == meshWidePolicy:
		if !meshSupported {
			return errors.New(meshNotSupportedError)
		}
		options.resourceName = types.NamespacedName{
			Name: meshWidePolicy,
		}
	case !strings.Contains(options.resourceID, "/"):
		if !util.IsValidK8sResourceName(options.resourceID) {
			return errors.Errorf("%s is not in a valid format", options.resourceID)
		}
		options.resourceName = types.NamespacedName{
			Namespace: options.resourceID,
		}
	case !strings.Contains(options.resourceID, ":"):
		options.resourceName, err = util.ParseK8sResourceID(options.resourceID)
		if err != nil {
			return errors.WrapIf(err, "could not parse resource ID")
		}
	default:
		if portsSupported {
			parts := strings.Split(options.resourceID, ":")
			if len(parts) != 2 {
				return errors.Errorf("invalid resource ID: '%s': format must be <namespace>/<name>:<port>", options.resourceID)
			}

			options.resourceName, err = util.ParseK8sResourceID(parts[0])
			if err != nil {
				return errors.WrapIf(err, "could not parse resource ID")
			}

			portNumber, err := strconv.Atoi(parts[1])
			if err != nil {
				options.portName = parts[1]
			} else {
				options.portNumber = portNumber
			}
		}
	}

	return nil
}

func isAutoMTLSEnabled(cli cli.CLI) bool {
	cl, err := cli.GetK8sClient()
	if err != nil {
		panic(errors.WrapIf(err, "could not get k8s client"))
	}

	istioCR, err := istio.FetchIstioCR(cl)
	if err != nil {
		panic(err)
	}

	return istioCR.Spec.AutoMTLS
}

func getMesh(cli cli.CLI, options *mTLSOptions, client graphql.Client) error {
	meshPolicy, err := client.GetMeshWithMTLS()
	if err != nil {
		return errors.Wrap(err, "couldn't query mesh with mTLS")
	}

	if meshPolicy == nil {
		log.Info("no meshPolicy found for mesh")
		return nil
	}

	return Output(cli, options.resourceName, getMeshPolicyMap([]graphql.MeshPolicy{*meshPolicy}))
}

func getNamespace(cli cli.CLI, options *mTLSOptions, client graphql.Client) error {
	namespace, err := client.GetNamespaceWithMTLS(options.resourceName.Namespace)
	if err != nil {
		return errors.Wrap(err, "couldn't query namespace with mTLS")
	}

	if namespace.Namespace.Name == "" {
		return errors.New(fmt.Sprintf("namespace %s not found", options.resourceName.Namespace))
	}

	return Output(cli, options.resourceName, getPolicyMap([]graphql.Policy{namespace.Namespace.Policy}))
}

func getService(cli cli.CLI, options *mTLSOptions, client graphql.Client) error {
	service, err := client.GetServiceWithMTLS(options.resourceName.Namespace, options.resourceName.Name)
	if err != nil {
		return errors.Wrap(err, "couldn't query service with mTLS")
	}

	if len(service.Policies) == 0 {
		log.Infof("no policy found for %s", options.resourceName)
		return nil
	}

	return Output(cli, options.resourceName, getPolicyMap(service.Policies))
}

func setMTLS(cli cli.CLI, options *mTLSOptions, client graphql.Client, mode mTLSMode) error {
	req := prepareApplyPolicyPeersRequest(options, mode)

	switch {
	case options.resourceName.Name == meshWidePolicy:
		req := prepareApplyMeshPolicyRequest(mode)
		err := applyMeshPolicy(client, req)
		if err != nil {
			return err
		}

		log.Infof("switched global mTLS to %s successfully", mode)
		return nil
	case options.resourceName.Name == "":
		err := applyPolicyPeers(client, req)
		if err != nil {
			return err
		}

		log.Infof("policy peers for %s set successfully\n\n", options.resourceName)
		return getNamespace(cli, options, client)
	default:
		_, err := client.GetServiceWithMTLS(options.resourceName.Namespace, options.resourceName.Name)
		if err != nil {
			return errors.New("couldn't query service with mTLS, check the resource ID")
		}

		err = applyPolicyPeers(client, req)
		if err != nil {
			return err
		}

		log.Infof("policy peers for %s set successfully\n\n", options.resourceName)
		return getService(cli, options, client)
	}
}

func unsetMTLS(cli cli.CLI, options *mTLSOptions, client graphql.Client) error {
	req := prepareDisablePolicyPeersRequest(options)

	switch {
	case options.resourceName.Name == meshWidePolicy:
		return errors.New(meshNotSupportedError)
	case options.resourceName.Name == "":
		err := disablePolicyPeers(client, req)
		if err != nil {
			return err
		}

		log.Infof("policy peers for %s unset successfully\n\n", options.resourceName)
		return getNamespace(cli, options, client)
	default:
		_, err := client.GetServiceWithMTLS(options.resourceName.Namespace, options.resourceName.Name)
		if err != nil {
			return errors.New("couldn't query service with mTLS, check the resource ID")
		}

		err = disablePolicyPeers(client, req)
		if err != nil {
			return err
		}

		log.Infof("policy peers for %s unset successfully\n\n", options.resourceName)
		return getService(cli, options, client)
	}
}

func prepareApplyMeshPolicyRequest(mode mTLSMode) graphql.ApplyMeshPolicyInput {
	mtlsMode := graphql.MTLSModeInput(mode)
	var req = graphql.ApplyMeshPolicyInput{
		MTLSMode: &mtlsMode,
	}

	return req
}

func prepareApplyPolicyPeersRequest(options *mTLSOptions, mode mTLSMode) graphql.ApplyPolicyPeersInput {
	var req = graphql.ApplyPolicyPeersInput{
		Selector: &graphql.PolicySelectorInput{
			Namespace: options.resourceName.Namespace,
		},
		Peers: []*graphql.PeerAuthenticationMethodInput{
			{
				Mtls: &graphql.MutualTLSInput{},
			},
		},
	}

	switch mode {
	case ModePermissive:
		req.Peers[0].Mtls.Mode = graphql.AuthTLSModeInputToPointer(graphql.AuthTLSModeInputPermissive)
	case ModeStrict:
		req.Peers[0].Mtls.Mode = graphql.AuthTLSModeInputToPointer(graphql.AuthTLSModeInputStrict)
	case ModeDisabled:
		req.Peers[0].Mtls.Mode = nil
	}

	if options.resourceName.Name != "" {
		req.Selector.Target = &graphql.TargetSelectorInput{
			Name: options.resourceName.Name,
		}
	}

	if options.portName != "" {
		req.Selector.Target.Port = &graphql.AuthPortSelectorInput{
			Name: &options.portName,
		}
	} else if options.portNumber != 0 {
		req.Selector.Target.Port = &graphql.AuthPortSelectorInput{
			Number: &options.portNumber,
		}
	}

	return req
}

func prepareDisablePolicyPeersRequest(options *mTLSOptions) graphql.DisablePolicyPeersInput {
	var req = graphql.DisablePolicyPeersInput{
		Selector: &graphql.PolicySelectorInput{
			Namespace: options.resourceName.Namespace,
		},
	}

	if options.resourceName.Name != "" {
		req.Selector.Target = &graphql.TargetSelectorInput{
			Name: options.resourceName.Name,
		}
	}

	if options.portName != "" {
		req.Selector.Target.Port = &graphql.AuthPortSelectorInput{
			Name: &options.portName,
		}
	} else if options.portNumber != 0 {
		req.Selector.Target.Port = &graphql.AuthPortSelectorInput{
			Number: &options.portNumber,
		}
	}

	return req
}

func applyMeshPolicy(client graphql.Client, req graphql.ApplyMeshPolicyInput) error {
	response, err := client.ApplyMeshPolicy(req)
	if err != nil {
		return errors.WrapIf(err, "could not apply mesh policy")
	}

	if !response {
		return errors.New("unknown internal error: could not apply mesh policy")
	}

	return nil
}

func applyPolicyPeers(client graphql.Client, req graphql.ApplyPolicyPeersInput) error {
	response, err := client.ApplyPolicyPeers(req)
	if err != nil {
		return errors.WrapIf(err, "could not apply policy peers")
	}

	if !response {
		return errors.New("unknown internal error: could not apply policy peers")
	}

	return nil
}

func disablePolicyPeers(client graphql.Client, req graphql.DisablePolicyPeersInput) error {
	response, err := client.DisablePolicyPeers(req)
	if err != nil {
		return errors.WrapIf(err, "could not unset policy peers")
	}

	if !response {
		return errors.New("unknown internal error: could not unset policy peers")
	}

	return nil
}

func getPolicyMap(policies []graphql.Policy) map[string][]*v1alpha1.PolicySpec {
	policySpecs := make(map[string][]*v1alpha1.PolicySpec)
	for _, p := range policies {
		p := p
		policySpecs[p.Namespace+"/"+p.Name] = []*v1alpha1.PolicySpec{&p.Spec}
	}
	return policySpecs
}

func getMeshPolicyMap(policies []graphql.MeshPolicy) map[string][]*v1alpha1.PolicySpec {
	policySpecs := make(map[string][]*v1alpha1.PolicySpec)
	for _, p := range policies {
		p := p
		policySpecs[p.Namespace+"/"+p.Name] = []*v1alpha1.PolicySpec{&p.Spec}
	}
	return policySpecs
}
