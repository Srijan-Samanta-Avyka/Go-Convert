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

import "encoding/xml"

type (
	Builders struct {
		Tasks       []Task `xml:",any"`
		BuilderName string `xml:"builderName"`
		BuildType   string `xml:"buildType"`
	}

	Task struct {
		XMLName xml.Name `xml:"task"`
		Content string   `xml:",innerxml"`
	}

	Agent struct {
		Label  string `json:"label"`
		Docker string `json:"docker"`
		None   string `json:"none"`
	}

	Environment struct {
		Variables map[string]Variable `json:"variables"`
	}

	Variable struct {
		Value     string
		TypeValue string
	}

	Stage struct {
		Name        string                   `json:"name"`
		Steps       []Steps                  `json:"steps"`
		Post        Post                     `json:"post"`
		When        When                     `json:"when"`
		Parallel    Parallel                 `json:"parallel"`
		Environment Environment              `json:"environment"`
		Matrix      map[string][]interface{} `json:"strategy"`
		Options     Options
	}

	Steps struct {
		Echo   string   `json:"echo"`
		Shell  string   `json:"shell"`
		Script []Script `json:"script"`
	}

	Script struct {
		Echo    string
		Shell   string
		Comment string
	}

	When struct {
		Branch          string        `json:"branch"`        // Git branch conditio
		Tag             string        `json:"tag"`           // Git tag condition
		Expression      string        `json:"expression"`    // Custom expression for condition
		BuildingTag     bool          `json:"buildingTag"`   // Build is triggered by a tag
		Changelog       string        `json:"changelog"`     // SCM changelog matches regex
		Changeset       string        `json:"changeset"`     // SCM changeset matches pattern
		Changerequest   ChangeRequest `json:"changeRequest"` // Build is a change request (PR/MR)
		Environment     string        `json:"environment"`   // Matches environment variable condition
		NotCondition    []Condition   `json:"notCondition,omitempty"`
		AllOfConditions []Condition   `json:"allOfConditions,omitempty"`
		AnyOfConditions []Condition   `json:"anyOfConditions,omitempty"`
		// Equals        string `json:"equals"`        // Compares expected vs actual values
		// // AllOf         []When `json:"allOf"`         // All conditions must be true
		// // AnyOf         []When `json:"anyOf"`         // At least one condition must be true
		// Not           *When  `json:"not"`           // Negates the condition
		// TriggeredBy   string `json:"triggeredBy"`   // Triggered by specific cause (e.g., SCMTrigger)
	}

	ChangeRequest struct {
		Enabled           bool   `json:"enabled"`           // Whether to check for a change request
		ID                string `json:"id"`                // ID of the change request (PR/MR)
		Target            string `json:"target"`            // Target branch of the change request
		Branch            string `json:"branch"`            // Source branch of the change request
		Fork              string `json:"fork"`              // Fork repository info
		URL               string `json:"url"`               // URL of the change request
		Title             string `json:"title"`             // Title of the change request
		Author            string `json:"author"`            // Author of the change request
		AuthorDisplayName string `json:"authorDisplayName"` // Display name of the author
		AuthorEmail       string `json:"authorEmail"`       // Author's email
	}

	Condition struct {
		Branch        string `json:"branch"`
		Tag           string `json:"tag"`
		ChangeRequest bool   `json:"changeRequest"`
	}

	Parallel struct {
		Stages []Stage `json:"stages"`
	}

	Matrix struct {
		Matrix map[string][]interface{}
	}

	Post struct {
		Always  Always  `json:"always"`
		Success Success `json:"success"`
		Failure Failure `json:"failure"`
	}

	Always struct {
		Echo string `json:"steps"`
	}

	Success struct {
		Echo string `json:"steps"`
	}

	Failure struct {
		Echo string `json:"steps"`
	}

	Triggers struct {
		Cron    string `json:"cron"`
		PollSCM string `json:"pollSCM"`
	}

	Options struct {
		TimeOut  TimeOut
		Retry    string
		Upstream string
	}

	TimeOut struct {
		Time string
		Unit string
	}

	Tools struct {
		Jdk   string `json:"jdk"`
		Maven string `json:"maven"`
	}

	Parameters struct {
		Params []Parameter `json:"parameters"`
	}

	// Parameter is a unified struct that represents different types of parameters
	Parameter struct {
		Name         string   `json:"name"`
		DefaultValue string   `json:"defaultValue,omitempty"`
		Description  string   `json:"description,omitempty"`
		Type         string   `json:"type"`                   // Type of parameter (e.g., string, text, boolean, choice, password)
		Choices      []string `json:"choices,omitempty"`      // Only for "choice" type parameters
		BooleanValue bool     `json:"booleanValue,omitempty"` // Only for "boolean" type parameters
	}
	Input struct {
		Message    string           `json:"message"`
		OkText     string           `json:"ok"`
		Submitter  string           `json:"submitter"`
		Parameters []InputParameter `json:"parameters"`
	}

	// InputParameter defines parameters that will be provided as input.
	InputParameter struct {
		Name         string `json:"name"`
		DefaultValue string `json:"defaultValue"`
		Description  string `json:"description"`
	}
)
