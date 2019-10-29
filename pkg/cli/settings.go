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
	"reflect"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	Namespace        = "backyards.namespace"
	URL              = "backyards.url"
	CACert           = "backyards.cacert"
	Token            = "backyards.token"
	TrackingClientID = "backyards.trackingClientId"
	LicenseAccepted  = "backyards.licenseAccepted"
	UsePortForward   = "backyards.usePortForward"
	LocalPort        = "backyards.localPort"
)

type Setting struct {
	Flag        string
	Default     string
	Description string
	Shorthand   string
	Kind        reflect.Kind
	Env         string
}

type Settings map[string]Setting

var PersistentConfigurationSettings = Settings{
	Namespace: {
		Flag:        "namespace",
		Default:     "backyards-system",
		Description: "Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE]",
		Shorthand:   "n",
		Kind:        reflect.String,
		Env:         "BACKYARDS_NAMESPACE",
	},
	URL: {
		Flag:        "base-url",
		Description: "Custom Backyards base URL (uses port forwarding or proxying if empty)",
		Kind:        reflect.String,
	},
	CACert: {
		Flag:        "cacert",
		Description: "The CA to use for verifying Backyards' server certificate",
		Kind:        reflect.String,
	},
	Token: {
		Flag:        "token",
		Description: "Authentication token to use to communicate with Backyards",
		Kind:        reflect.String,
	},
	TrackingClientID: {
		Description: "Google Analytics tracking client ID",
		Kind:        reflect.String,
	},
	LicenseAccepted: {
		Description: "Accept the Backyards proprietary license",
		Kind:        reflect.Bool,
	},
	UsePortForward: {
		Description: "Use port forwarding instead of proxying to reach Backyards",
		Flag:        "use-portforward",
		Default:     "false",
		Kind:        reflect.Bool,
	},
	LocalPort: {
		Description: "Use this local port for port forwarding / proxying to Backyards (when set to 0, a random port will be used)",
		Shorthand:   "p",
		Flag:        "local-port",
		Default:     "-1",
		Kind:        reflect.Int,
	},
}

func (i Settings) Configure(flags *flag.FlagSet) {
	for _, item := range i {
		if item.Flag != "" {
			switch item.Kind {
			case reflect.Int:
				_ = flags.IntP(item.Flag, item.Shorthand, cast.ToInt(item.Default), item.Description)
			case reflect.String:
				_ = flags.StringP(item.Flag, item.Shorthand, cast.ToString(item.Default), item.Description)
			case reflect.Bool:
				_ = flags.BoolP(item.Flag, item.Shorthand, cast.ToBool(item.Default), item.Description)
			default:
				logrus.Errorf("Unsupported field type: %s", item.Kind)
			}
		}
	}
}

func (i Settings) Bind(viper *viper.Viper, flags *flag.FlagSet) {
	for key, item := range i {
		if item.Flag != "" {
			err := viper.BindPFlag(key, flags.Lookup(item.Flag))
			if err != nil {
				logrus.Errorf("Failed to bind flags %+v", err)
			}
			if item.Env != "" {
				err = viper.BindEnv(key, item.Env)
				if err != nil {
					logrus.Errorf("Failed to bind env %+v", err)
				}
			}
		}
	}
}
