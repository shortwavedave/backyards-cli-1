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
	Token() string

	SetTrackingClientID(string)
	SetToken(string)

	PersistConfig() error

	GetConfigFileUsed() string
}

type viperPersistentConfig struct {
	viper   *viper.Viper
	changed map[string]interface{}
}

func newViperPersistentConfig(persistentConfig *viper.Viper) PersistentConfig {
	return &viperPersistentConfig{
		viper:   persistentConfig,
		changed: make(map[string]interface{}),
	}
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

func (b *viperPersistentConfig) Token() string {
	return b.viper.GetString(Token)
}

func (b *viperPersistentConfig) SetTrackingClientID(clientID string) {
	b.set(TrackingClientID, clientID)
}

func (b *viperPersistentConfig) SetToken(token string) {
	b.set(Token, token)
}

func (b *viperPersistentConfig) set(key string, value interface{}) {
	original := b.viper.Get(key)
	b.changed[key] = original
	b.viper.Set(key, value)
}

func (b *viperPersistentConfig) PersistConfig() error {
	if len(b.changed) > 0 {
		if err := persistViper(b.viper); err != nil {
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
