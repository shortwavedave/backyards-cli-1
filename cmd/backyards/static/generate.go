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
// +build ignore

package main

import (
	"github.com/shurcooL/vfsgen"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static"
)

func main() {
	err := vfsgen.Generate(static.BackyardsChartSource, vfsgen.Options{
		Filename:     "static/backyards/chart.gogen.go",
		PackageName:  "backyards",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.IstioOperatorChartSource, vfsgen.Options{
		Filename:     "static/istio_operator/chart.gogen.go",
		PackageName:  "istio_operator",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CanaryOperatorChartSource, vfsgen.Options{
		Filename:     "static/canary_operator/chart.gogen.go",
		PackageName:  "canary_operator",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.BackyardsDemoChartSource, vfsgen.Options{
		Filename:     "static/backyards_demo/chart.gogen.go",
		PackageName:  "backyards_demo",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.IstioAssetsSource, vfsgen.Options{
		Filename:     "static/istio_assets/assets.gogen.go",
		PackageName:  "istio_assets",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CertManagerChartSource, vfsgen.Options{
		Filename:     "static/certmanager/chart.gogen.go",
		PackageName:  "certmanager",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CertManagerCainjectorChartSource, vfsgen.Options{
		Filename:     "static/certmanagercainjector/chart.gogen.go",
		PackageName:  "certmanagercainjector",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CertManagerCRDSource, vfsgen.Options{
		Filename:     "static/certmanagercrds/chart.gogen.go",
		PackageName:  "certmanagercrds",
		VariableName: "CRDs",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.GraphTemplates, vfsgen.Options{
		Filename:     "static/graphtemplates/graphtemplates.gogen.go",
		PackageName:  "graphtemplates",
		VariableName: "GraphTemplates",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.Licenses, vfsgen.Options{
		Filename:     "static/licenses/licenses.gogen.go",
		PackageName:  "licenses",
		VariableName: "Licenses",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.PeerClusterAssetsSource, vfsgen.Options{
		Filename:     "static/peercluster/assets.gogen.go",
		PackageName:  "peercluster",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.NodeExporterChartSource, vfsgen.Options{
		Filename:     "static/nodeexporter/chart.gogen.go",
		PackageName:  "nodeexporter",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
