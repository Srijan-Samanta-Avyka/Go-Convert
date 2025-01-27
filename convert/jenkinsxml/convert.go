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

// Package jenkinsxml converts Jenkins XML pipelines to Harness pipelines.
package jenkinsxml

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Srijan-Samanta-Avyka/go-convert/convert/harnessNew"
	jenkinsxml "github.com/Srijan-Samanta-Avyka/go-convert/convert/jenkinsxml/xml"
	"github.com/Srijan-Samanta-Avyka/go-convert/internal/slug"
	"github.com/Srijan-Samanta-Avyka/go-convert/internal/store"

	"github.com/ghodss/yaml"
)

// conversion context
type context struct {
	config *jenkinsxml.Project
}

// Converter converts a Jenkins XML file to a Harness
// v1 pipeline.
type Converter struct {
	kubeEnabled   bool
	kubeNamespace string
	kubeConnector string
	dockerhubConn string
	identifiers   *store.Identifiers
}

func New(options ...Option) *Converter {
	d := new(Converter)

	// create the unique identifier store. this store
	// is used for registering unique identifiers to
	// prevent duplicate names, unique index violations.
	d.identifiers = store.New()

	// loop through and apply the options.
	for _, option := range options {
		option(d)
	}

	// set the default kubernetes namespace.
	if d.kubeNamespace == "" {
		d.kubeNamespace = "default"
	}

	// set the runtime to kubernetes if the kubernetes
	// connector is configured.
	if d.kubeConnector != "" {
		d.kubeEnabled = true
	}

	return d
}

// Convert downgrades a v1 pipeline.
func (d *Converter) Convert(r io.Reader) ([]byte, error) {
	src, err := jenkinsxml.Parse(r)
	if err != nil {
		return nil, err
	}
	return d.convert(&context{
		config: src,
	})
}

// ConvertBytes downgrades a v1 pipeline.
func (d *Converter) ConvertBytes(b []byte) ([]byte, error) {
	return d.Convert(
		bytes.NewBuffer(b),
	)
}

// ConvertString downgrades a v1 pipeline.
func (d *Converter) ConvertString(s string) ([]byte, error) {
	return d.Convert(
		bytes.NewBufferString(s),
	)
}

// ConvertFile downgrades a v1 pipeline.
func (d *Converter) ConvertFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return d.Convert(f)
}

// converts converts a Jenkins XML pipeline to a Harness pipeline.
func (d *Converter) convert(ctx *context) ([]byte, error) {

	// create the harness pipeline spec

	dst := harnessNew.Pipeline{}
	dst.Props.CI.Codebase.Build = "<+input>"
	// create the harness pipeline resource
	// config := &harness.Config{
	// 	Version: 1,
	// 	Kind:    "pipeline",
	// 	Spec:    dst,
	// }

	config := &harnessNew.Config{
		Pipeline: dst,
	}

	// Convert the Enviornments
	// for _, envVariables := range ctx.config.Pipeline.Environment.Variables {
	// 	fmt.Print(envVariables)
	// 	convertParameters()
	// }

	dst.ID = "InOrg_Id"
	dst.Name = "Converter_Name"

	// if config.Pipeline.ID == harness.DefaultName && p.Name != "" {
	// 	config.Pipeline.ID = slug.Create(p.Name)
	// }

	dst.Org = "org_Identifier"
	dst.Project = "project_Identifier"

	//dst.Timeout =

	// Handling Pipeline Variables

	for _, param := range ctx.config.Pipeline.Parameters.Params {

		if param.Type == "string" || param.Type == "text" || param.Type == "booleanParam" || param.Type == "choice" {
			if param.Type == "booleanParam" {
				param.DefaultValue = "true"
			}
			param.Type = "String"
			if param.DefaultValue == "" {
				param.DefaultValue = "default"
			}
		}
		if param.Type == "password" {
			param.Type = "Secret"
		}

		variable := harnessNew.Variable{
			Type:  param.Type,
			Name:  param.Name,
			Value: param.DefaultValue,
		}
		dst.Variables = append(dst.Variables, &variable)
	}

	// Handling Environment Variables in Pipeline Variables

	for envKey, envValue := range ctx.config.Pipeline.Environment.Variables {
		variable := harnessNew.Variable{}
		envValue.TypeValue = "String"
		// variable := harnessNew.Variable{
		// 	Type:  envValue.TypeValue,
		// 	Name:  envKey,
		// 	Value: envValue.Value,
		// }
		variable.Type = envValue.TypeValue
		variable.Name = envKey
		if envValue.Value != "" {
			variable.Value = envValue.Value
		} else {
			variable.Value = "default"
		}
		dst.Variables = append(dst.Variables, &variable)

	}

	for _, stage := range ctx.config.Pipeline.Stages {
		var steps []harnessNew.Steps
		// convert each drone step to a harness step.
		var i int = 0

		// handling parallel block

		if len(stage.Parallel.Stages) != 0 {
			parallelStages := convertParallelStages(stage.Parallel.Stages)
			stages := harnessNew.Stages{
				Parallel: parallelStages,
			}
			dst.Stages = append(dst.Stages, &stages)
			continue
		}

		// Stages without steps won't be considered
		if len(stage.Steps) == 0 {
			continue
		}

		for _, v := range stage.Steps {
			steps = append(steps, convertSteps(v, i))
			i++
		}

		stage := &harnessNew.Stage{
			Name: stage.Name,
			ID:   slug.Create(stage.Name),
			Type: "CI",
			Spec: &harnessNew.StageCI{
				Runtime: &harnessNew.Runtime{
					Type: "Cloud",
					Spec: struct{}{},
				},
				Clone: false,
				Execution: harnessNew.Execution{
					Steps: steps,
				},
				Platform: &harnessNew.Platform{
					OS:   "Linux",
					Arch: "Arm64",
				},
			},
			When:    convertStageCondition(stage.When),
			Vars:    convertEnvironmentVariables(stage.Environment),
			TimeOut: convertStageTimeOut(&stage.Options.TimeOut),
		}
		stages := harnessNew.Stages{
			Stage: stage,
		}
		dst.Stages = append(dst.Stages, &stages)

	}

	// Handling Post block in Pipeline As A Stage This should be handled as a function

	if (ctx.config.Pipeline.Post != jenkinsxml.Post{}) {
		commandForPostStep := generateCommandStringForPost(ctx.config.Pipeline.Post)

		stageForPost := &harnessNew.Stage{
			Name: "Post",
			ID:   slug.Create("Post"),
			Type: "CI",
			Spec: &harnessNew.StageCI{
				Runtime: &harnessNew.Runtime{
					Type: "Cloud",
					Spec: struct{}{},
				},
				Execution: harnessNew.Execution{
					Steps: []harnessNew.Steps{
						{
							Step: &harnessNew.Step{
								Name: "postRunStep",
								ID:   slug.Create("postRunStep"),
								Type: "Run",
								Spec: harnessNew.StepRun{
									Command: commandForPostStep,
								},
							},
						},
					},
				},
				Clone: false,
				Platform: &harnessNew.Platform{
					OS:   "Linux",
					Arch: "Arm64",
				},
			},
		}
		stagesForPost := harnessNew.Stages{
			Stage: stageForPost,
		}

		dst.Stages = append(dst.Stages, &stagesForPost)
	}
	//Handling The Advanced Option For TimeOut

	if ctx.config.Pipeline.Options.TimeOut.Time != "" {
		dst.Timeout = ctx.config.Pipeline.Options.TimeOut.Time
		if ctx.config.Pipeline.Options.TimeOut.Unit == "MINUTES" {
			dst.Timeout = dst.Timeout + "m"
		}
		if ctx.config.Pipeline.Options.TimeOut.Unit == "SECONDS" {
			dst.Timeout = dst.Timeout + "s"
		}
	}

	config.Pipeline = dst
	out, err := yaml.Marshal(config)
	// cacheFound := false

	// create the harness stage.
	/*dstStage := &harness.Stage{
		Type: "ci",
		// When: convertCond(from.Trigger),
		Spec: &harness.StageCI{
			// Delegate: convertNode(from.Node),
			// Envs: convertVariables(ctx.config.Variables),
			// Platform: convertPlatform(from.Platform),
			// Runtime:  convertRuntime(from),
			Steps: make([]*harness.Step, 0), // Initialize the Steps slice
		},
	}
	dst.Stages = append(dst.Stages, dstStage)*/
	//stageSteps := make([]*harness.Step, 0)

	/*tasks := ctx.config.Builders.Tasks
	for _, task := range tasks {
		step := &harness.Step{}

		switch taskname := task.XMLName.Local; taskname {
		case "hudson.tasks.Shell":
			step = convertShellTaskToStep(&task)
		case "hudson.tasks.Ant":
			step = convertAntTaskToStep(&task)
		default:
			step = unsupportedTaskToStep(taskname)
		}

		stageSteps = append(stageSteps, step)
	}
	dstStage.Spec.(*harness.StageCI).Steps = stageSteps*/

	// marshal the harness yaml
	if err != nil {
		return nil, err
	}

	return out, nil
}

// convertAntTaskToStep converts a Jenkins Ant task to a Harness step.
// func convertAntTaskToStep(task *jenkinsxml.Task) *harness.Step {
// 	antTask := &jenkinsxml.HudsonAntTask{}
// 	// TODO: wrapping task.Content with 'builders' tags is ugly.
// 	err := xml.Unmarshal([]byte("<builders>"+task.Content+"</builders>"), antTask)
// 	if err != nil {
// 		return nil
// 	}

// 	spec := new(harness.StepPlugin)
// 	spec.Image = "harnesscommunitytest/ant-plugin"
// 	spec.Inputs = map[string]interface{}{
// 		"goals": antTask.Targets,
// 	}
// 	step := &harness.Step{
// 		Name: "ant",
// 		Type: "plugin",
// 		Spec: spec,
// 	}

// 	return step
// }

// // convertShellTaskToStep converts a Jenkins Shell task to a Harness step.
// func convertShellTaskToStep(task *jenkinsxml.Task) *harness.Step {
// 	shellTask := &jenkinsxml.HudsonShellTask{}
// 	// TODO: wrapping task.Content with 'builders' tags is ugly.
// 	err := xml.Unmarshal([]byte("<builders>"+task.Content+"</builders>"), shellTask)
// 	if err != nil {
// 		return nil
// 	}

// 	spec := new(harness.StepExec)
// 	spec.Run = shellTask.Command
// 	step := &harness.Step{
// 		Name: "shell",
// 		Type: "script",
// 		Spec: spec,
// 	}

// 	return step
// }

// // unsupportedTaskToStep converts an unsupported Jenkins Task to a placeholder
// // Harness step.
// func unsupportedTaskToStep(task string) *harness.Step {
// 	spec := new(harness.StepExec)
// 	spec.Run = "echo Unsupported field " + task
// 	step := &harness.Step{
// 		Name: "shell",
// 		Type: "script",
// 		Spec: spec,
// 	}

// 	return step
// }

// func convertParameters(in jenkinsxml.Environment) map[string]*harness.Input {
// 	out := map[string]*harness.Input{}
// 	for name, param := range in.Variables {
// 		t := param.TypeValue
// 		switch t {
// 		case "integer":
// 			t = "number"
// 		case "string", "enum", "env_var_name":
// 			t = "string"
// 		case "boolean":
// 			t = "boolean"
// 		case "executor", "steps":
// 			// TODO parameter.type execution not supported
// 			// TODO parameter.type steps not supported
// 			continue // skip
// 		}
// 		out[name] = &harness.Input{
// 			Type:    t,
// 			Default: param.Value,
// 		}
// 		fmt.Print(name, param.TypeValue)
// 	}
// 	return out
// }

func convertSteps(step jenkinsxml.Steps, i int) harnessNew.Steps {

	var stepType string
	var stepCommandType string
	var stepCommand string
	if step.Echo != "" {
		stepType = "run"
		stepCommandType = "echo"
		stepCommand = step.Echo
	}
	if step.Shell != "" {
		stepType = "run"
		stepCommandType = "shell"
		stepCommand = step.Shell
	}
	// if step.Script.Echo != "" {
	// 	stepType = "run"
	// 	stepCommandType = "Script.Echo"
	// 	stepCommand = step.Script.Echo
	// }
	// if step.Script.Shell != "" {
	// 	stepType = "run"
	// 	stepCommandType = "Script.Shell"
	// 	stepCommand = step.Script.Shell
	// }
	if len(step.Script) > 0 {
		stepType = "run"
		stepCommandType = "script"
		for i = 0; i < len(step.Script); i++ {
			if step.Script[i].Echo != "" {
				stepCommand = stepCommand + "echo " + "\"" + step.Script[i].Echo + "\"" + "\n"
			} else if step.Script[i].Shell != "" {
				stepCommand = stepCommand + step.Script[i].Shell + "\n"
			} else if step.Script[i].Comment != "" {
				stepCommand = stepCommand + "# To_Do_Change " + step.Script[i].Comment + "\n"
			}
		}
	}

	harnessStep := &harnessNew.Step{
		Name: "RunStep" + strconv.Itoa(i),
		ID:   slug.Create("RunStep" + strconv.Itoa(i)),
		Type: "Run",
		Spec: struct{}{},
	}

	if stepType == "run" {
		runStep := convertStepRun(stepCommandType, stepCommand)
		harnessStep.Spec = runStep
	}

	harnessSteps := harnessNew.Steps{
		Step: harnessStep,
	}
	return harnessSteps
}

func convertStepRun(stepCommandType string, stepCommand string) harnessNew.StepRun {

	if stepCommandType == "echo" {
		return harnessNew.StepRun{
			Command: "echo " + "\"" + stepCommand + "\"",
			Shell:   "Sh",
		}
	}
	if stepCommandType == "shell" {
		return harnessNew.StepRun{
			Command: stepCommand,
			Shell:   "Sh",
		}
	}
	if stepCommandType == "script" {
		return harnessNew.StepRun{
			Command: stepCommand,
			Shell:   "Sh",
		}
	}
	return harnessNew.StepRun{}
}

func convertParallelStages(parallelStages []jenkinsxml.Stage) []*harnessNew.Stages {

	var allParallelStages []*harnessNew.Stages
	fmt.Print(allParallelStages)
	var steps []harnessNew.Steps
	// convert each drone step to a harness step.
	var i int = 0

	for _, stage := range parallelStages {

		for _, v := range stage.Steps {
			steps = append(steps, convertSteps(v, i))
			i++
		}
		stage := &harnessNew.Stage{
			Name: stage.Name,
			ID:   slug.Create(stage.Name),
			Type: "CI",
			Spec: &harnessNew.StageCI{
				Runtime: &harnessNew.Runtime{
					Type: "Cloud",
					Spec: struct{}{},
				},
				Execution: harnessNew.Execution{
					Steps: steps,
				},
				Platform: &harnessNew.Platform{
					OS:   "Linux",
					Arch: "Arm64",
				},
				Clone: false,
			},
		}
		stages := harnessNew.Stages{
			Stage: stage,
		}
		allParallelStages = append(allParallelStages, &stages)

	}
	return allParallelStages

}

func convertStageCondition(stageCondition jenkinsxml.When) *harnessNew.StageWhen {
	var harnessCondition string
	if stageCondition == (jenkinsxml.When{}) {
		return nil
	}
	if stageCondition.Branch != "" {
		harnessCondition = harnessCondition + "<+pipeline.branch> ==" + "\"" + stageCondition.Branch + "\""
	}
	return &harnessNew.StageWhen{
		PipelineStatus: "Success",
		Condition:      harnessCondition,
	}
}

func generateCommandStringForPost(postBlockData jenkinsxml.Post) string {
	var commandBuilder strings.Builder

	if postBlockData.Always.Echo != "" {
		commandBuilder.WriteString(fmt.Sprintf("if [ \"<+pipeline.status>\" = \"success\" ]; then\n    echo \"%s\"\n", postBlockData.Always))
	}

	if postBlockData.Success.Echo != "" {
		if postBlockData.Always.Echo != "" {
			commandBuilder.WriteString(fmt.Sprintf("elif [ \"<+pipeline.status>\" = \"success\" ]; then\n    echo \"%s\"\n", postBlockData.Success))
		} else {
			commandBuilder.WriteString(fmt.Sprintf("if [ \"<+pipeline.status>\" = \"success\" ]; then\n    echo \"%s\"\n", postBlockData.Success))
		}
	}

	if postBlockData.Failure.Echo != "" {
		if postBlockData.Success.Echo != "" || postBlockData.Always.Echo != "" {
			commandBuilder.WriteString(fmt.Sprintf("elif [ \"<+pipeline.status>\" = \"failure\" ]; then\n    echo \"%s\"\n", postBlockData.Failure))
		} else {
			commandBuilder.WriteString(fmt.Sprintf("if [ \"<+pipeline.status>\" = \"failure\" ]; then\n    echo \"%s\"\n", postBlockData.Failure))
		}
	}

	if commandBuilder.Len() > 0 {
		commandBuilder.WriteString("fi\n")
	} else {
		commandBuilder.WriteString("# Post Deployment Action cannot be converted\n")
	}

	return commandBuilder.String()
}

func convertEnvironmentVariables(stageVariables jenkinsxml.Environment) []*harnessNew.Variable {
	var harnessStageVariables []*harnessNew.Variable
	for envKey, envValue := range stageVariables.Variables {
		variable := harnessNew.Variable{}
		if envValue.TypeValue == "string" {
			envValue.TypeValue = "String"
			// variable := harnessNew.Variable{
			// 	Type:  envValue.TypeValue,
			// 	Name:  envKey,
			// 	Value: envValue.Value,
			// }
			variable.Type = envValue.TypeValue
			variable.Name = envKey
			variable.Value = envValue.Value
			harnessStageVariables = append(harnessStageVariables, &variable)
		}

	}
	return harnessStageVariables
}

func convertStageTimeOut(stageTimeOut *jenkinsxml.TimeOut) string {
	// dst.Timeout = ctx.config.Pipeline.Options.TimeOut.Time
	// if ctx.config.Pipeline.Options.TimeOut.Unit == "MINUTES" {
	// 	dst.Timeout = dst.Timeout + "m"
	// }
	// if ctx.config.Pipeline.Options.TimeOut.Unit == "SECONDS" {
	// 	dst.Timeout = dst.Timeout + "s"
	// }
	var TimeOut string
	if stageTimeOut.Time != "" {
		if stageTimeOut.Unit == "MINUTES" {
			TimeOut = stageTimeOut.Time + "m"
		}
		if stageTimeOut.Unit == "SECONDS" {
			TimeOut = stageTimeOut.Time + "s"
		}
	}
	return TimeOut
}

// func (d *Downgrader) convertStepRun(src *v1.Step) *v0.Step {
// 	spec_ := src.Spec.(*v1.StepExec)
// 	var id = d.identifiers.Generate(
// 		slug.Create(src.Id),
// 		slug.Create(src.Name),
// 		slug.Create(src.Type))
// 	if src.Name == "" {
// 		src.Name = id
// 	}

// 	// Convert outputs
// 	var outputs []*v0.Output
// 	for _, output := range spec_.Outputs {
// 		outputs = append(outputs, convertOutput(output))
// 	}

// 	return &v0.Step{
// 		ID:      id,
// 		Name:    convertName(src.Name),
// 		Type:    v0.StepTypeRun,
// 		Timeout: convertTimeout(src.Timeout),
// 		Spec: &v0.StepRun{
// 			Env:             spec_.Envs,
// 			Command:         spec_.Run,
// 			ConnRef:         d.dockerhubConn,
// 			Image:           convertImage(spec_.Image, d.defaultImage),
// 			ImagePullPolicy: convertImagePull(spec_.Pull),
// 			Outputs:         outputs, // Add this line
// 			Privileged:      spec_.Privileged,
// 			RunAsUser:       spec_.User,
// 			Reports:         convertReports(spec_.Reports),
// 			Shell:           strings.Title(spec_.Shell),
// 		},
// 		When:     convertStepWhen(src.When, id),
// 		Strategy: convertStrategy(src.Strategy),
// 	}
// }
