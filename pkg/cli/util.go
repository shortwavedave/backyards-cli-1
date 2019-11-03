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

package cli

import (
	"os"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

func createViper(file string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		var osPathError *os.PathError
		if errors.As(err, &osPathError) {
			logrus.Debugf("No configuration file has been loaded")
		} else {
			return nil, errors.WrapIff(err, "Failed to read config file %s", file)
		}
	} else {
		logrus.Debugf("Using config: %s", file)
	}
	return v, nil
}

func persistViper(v *viper.Viper) error {
	if _, err := os.Stat(filepath.Dir(v.ConfigFileUsed())); os.IsNotExist(err) {
		logrus.Debug("Creating config dir")
		configPath := filepath.Dir(v.ConfigFileUsed())
		err := os.MkdirAll(configPath, 0700)
		if err != nil {
			return errors.WrapIf(err, "failed to create config dir")
		}
	}
	logrus.Debugf("Saving current config settings to %s", v.ConfigFileUsed())
	logrus.Debugf("%#v", v.AllSettings())
	err := v.WriteConfig()
	if err != nil {
		return errors.WrapIf(err, "Failed to save config file")
	}
	return nil
}

func fileExists(file string) (bool, error) {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func getValidatedRawConfig() (*api.Config, error) {
	config, err := client.GetRawConfig(viper.GetString("kubeconfig"), viper.GetString("kubecontext"))
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get raw kubernetes config")
	}

	if len(config.Clusters) == 0 {
		return nil, errors.New("kubeconfig is invalid, no clusters defined")
	}

	var ok bool
	var currentContext *api.Context
	if currentContext, ok = config.Contexts[config.CurrentContext]; !ok {
		return nil, errors.Errorf("kubeconfig is invalid, current context data not available %s", config.CurrentContext)
	}

	if _, ok = config.Clusters[currentContext.Cluster]; !ok {
		return nil, errors.Errorf("kubeconfig is invalid, cluster data for current context not available %s", currentContext.Cluster)
	}

	return &config, nil
}
