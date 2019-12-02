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

package graphql

import (
	"context"
	"time"

	"github.com/MakeNowJust/heredoc"
)

type OverviewResponse struct {
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	Clusters        int       `json:"clusters"`
	Services        int       `json:"services"`
	ServicesInMesh  int       `json:"servicesInMesh"`
	Workloads       int       `json:"workloads"`
	WorkloadsInMesh int       `json:"workloadsInMesh"`
	Pods            int       `json:"pods"`
	PodsInMesh      int       `json:"podsInMesh"`
	ErrorRate       float32   `json:"errorRate"`
	Latency         float32   `json:"latency"`
	RPS             float32   `json:"rps"`
}

func (c *client) Overview(evaluationDurationSeconds uint) (OverviewResponse, error) {
	request := heredoc.Doc(`
	query overview($evaluationDurationSeconds: Int) {
		overview(evaluationDurationSeconds: $evaluationDurationSeconds) {
			start
			end
			clusters
			services
			servicesInMesh
			workloads
			workloadsInMesh
			pods
			podsInMesh
			errorRate
			latency
			rps
		}
	}
`)

	r := c.NewRequest(request)
	r.Var("evaluationDurationSeconds", evaluationDurationSeconds)

	// run it and capture the response
	var respData map[string]OverviewResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return OverviewResponse{}, err
	}

	return respData["overview"], nil
}
