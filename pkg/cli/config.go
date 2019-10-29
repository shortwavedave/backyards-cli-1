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
	"errors"
	"os"
	"path/filepath"

	emperror "emperror.dev/errors"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	PersistentConfigKey = "persistentConfig"
)

type PersistentConfig interface {
	Namespace() string
	BaseURL() string
	CACert() string
	LocalPort() int
	UsePortForward() bool
	TrackingClientID() string
	LicenseAccepted() bool
	Token() string

	SetTrackingClientID(string)
	SetLicenseAccepted(bool)
	SetToken(string)

	PersistConfig() error

	GetConfigFileUsed() string
}

type viperPersistentConfig struct {
	localConfigFile string
	viper           *viper.Viper
	settings        Settings
	flags           *flag.FlagSet
	changed         map[string]interface{}
}

func newViperPersistentConfig(persistentConfig string, settings Settings, flags *flag.FlagSet) (PersistentConfig, error) {
	config := &viperPersistentConfig{
		localConfigFile: persistentConfig,
		changed:         make(map[string]interface{}),
		settings:        settings,
		flags:           flags,
	}
	return config, config.loadConfig()
}

func (b *viperPersistentConfig) Namespace() string {
	return b.viper.GetString(Namespace)
}

func (b *viperPersistentConfig) BaseURL() string {
	return b.viper.GetString(URL)
}

func (b *viperPersistentConfig) CACert() string {
	return b.viper.GetString(CACert)
}

func (b *viperPersistentConfig) LocalPort() int {
	return b.viper.GetInt(LocalPort)
}

func (b *viperPersistentConfig) UsePortForward() bool {
	return b.viper.GetBool(UsePortForward)
}

func (b *viperPersistentConfig) TrackingClientID() string {
	return b.viper.GetString(TrackingClientID)
}

func (b *viperPersistentConfig) LicenseAccepted() bool {
	return b.viper.GetBool(LicenseAccepted)
}

func (b *viperPersistentConfig) Token() string {
	return b.viper.GetString(Token)
}

func (b *viperPersistentConfig) SetTrackingClientID(clientID string) {
	b.set(TrackingClientID, clientID)
}

func (b *viperPersistentConfig) SetLicenseAccepted(enabled bool) {
	b.set(LicenseAccepted, enabled)
}

func (b *viperPersistentConfig) SetToken(token string) {
	b.set(Token, token)
}

func (b *viperPersistentConfig) set(key string, value interface{}) {
	original := b.viper.Get(key)
	b.changed[key] = original
	b.viper.Set(key, value)
}

func (b *viperPersistentConfig) loadConfig() error {
	b.viper = viper.New()
	b.viper.SetConfigFile(b.localConfigFile)
	err := b.viper.ReadInConfig()
	if err != nil {
		var osPathError *os.PathError
		if errors.As(err, &osPathError) {
			logrus.Debugf("No configuration file has been loaded")
		} else {
			return emperror.WrapIff(err, "Failed to read config file %s", b.localConfigFile)
		}
	} else {
		logrus.Debugf("Using config: %s", b.localConfigFile)
	}
	b.settings.Bind(b.viper, b.flags)
	return nil
}

func (b *viperPersistentConfig) PersistConfig() error {
	if len(b.changed) > 0 {
		if _, err := os.Stat(filepath.Dir(b.viper.ConfigFileUsed())); os.IsNotExist(err) {
			logrus.Debug("Creating config dir")
			configPath := filepath.Dir(b.viper.ConfigFileUsed())
			err := os.MkdirAll(configPath, 0700)
			if err != nil {
				return emperror.WrapIf(err, "failed to create config dir")
			}
		}
		logrus.Debugf("Saving current config settings to %s", b.viper.ConfigFileUsed())
		logrus.Debugf("%#v", b.viper.AllSettings())
		err := b.viper.WriteConfig()
		if err != nil {
			return err
		}
		// restore initial state so that PersistConfig would be idempotent
		b.changed = make(map[string]interface{})
	}
	return nil
}

func (b *viperPersistentConfig) GetConfigFileUsed() string {
	return b.viper.ConfigFileUsed()
}
