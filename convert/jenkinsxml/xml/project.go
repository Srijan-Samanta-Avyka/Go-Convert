// Copyright 2024 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xml

type (
	// Project defines a jenkins project.
	Project struct {
		// TODO: add support for more fields.
		ConcurrentBuild bool      `xml:"concurrentBuild,omitempty"`
		Disabled        bool      `xml:"disabled,omitempty"`
		Builders        *Builders `xml:"builders"`
		Script          string    `xml:"definition>pipelineScript>script"`
		Pipeline        Pipeline
	}

	Pipeline struct {
		Agent       Agent       `json:"agent"`
		Environment Environment `json:"environment"`
		Stages      []Stage     `json:"stages"`
		Post        Post        `json:"post"`
		Triggers    Triggers    `json:"triggers"`
		Options     Options     `json:"options"`
		Tools       Tools       `json:"tools"`
		Parameters  Parameters  `json:"parameters"` // Add Parameters struct here

	}
)
