// StageCI defines a continuous integration stage.
package harnessNew

type (
	Config struct {
		Pipeline Pipeline `json:"pipeline" yaml:"pipeline"`
	}

	Pipeline struct {
		ID        string      `json:"identifier,omitempty"        yaml:"identifier,omitempty"`
		Name      string      `json:"name,omitempty"              yaml:"name,omitempty"`
		Desc      string      `json:"description,omitempty"       yaml:"description,omitempty"`
		Account   string      `json:"accountIdentifier,omitempty" yaml:"accountIdentifier,omitempty"`
		Project   string      `json:"projectIdentifier,omitempty" yaml:"projectIdentifier,omitempty"`
		Org       string      `json:"orgIdentifier,omitempty"     yaml:"orgIdentifier,omitempty"`
		Props     Properties  `json:"properties,omitempty"        yaml:"properties,omitempty"`
		Stages    []*Stages   `json:"stages,omitempty"            yaml:"stages"`
		Variables []*Variable `json:"variables,omitempty"         yaml:"variables,omitempty"`
		Timeout   string      `json:"timeout,omitempty"         yaml:"timeout,omitempty"`
	}

	StageCI struct {
		Cache          *Cache          `json:"caching,omitempty"             yaml:"caching,omitempty"`
		Clone          bool            `json:"cloneCodebase"       yaml:"cloneCodebase"`
		Execution      Execution       `json:"execution,omitempty"           yaml:"execution,omitempty"`
		Infrastructure *Infrastructure `json:"infrastructure,omitempty"      yaml:"infrastructure,omitempty"`
		Platform       *Platform       `json:"platform,omitempty"            yaml:"platform,omitempty"`
		Runtime        *Runtime        `json:"runtime,omitempty"            yaml:"runtime,omitempty"`
		Services       []*Service      `json:"serviceDependencies,omitempty" yaml:"serviceDependencies,omitempty"`
		SharedPaths    []string        `json:"sharedPaths,omitempty"         yaml:"sharedPaths,omitempty"`
	}

	// Properties defines pipeline properties.
	Properties struct {
		CI CI `json:"ci,omitempty" yaml:"ci,omitempty"`
	}

	// CI defines CI pipeline properties.
	CI struct {
		Codebase Codebase `json:"codebase,omitempty" yaml:"codebase,omitempty"`
	}

	// Cache defines the cache settings.
	Cache struct {
		Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Key     string   `json:"key,omitempty"     yaml:"key,omitempty"`
		Paths   []string `json:"paths,omitempty"   yaml:"paths,omitempty"`
	}

	// Codebase defines a codebase.
	Codebase struct {
		Name  string `json:"repoName,omitempty"     yaml:"repoName,omitempty"`
		Conn  string `json:"connectorRef,omitempty" yaml:"connectorRef,omitempty"`
		Build string `json:"build,omitempty"        yaml:"build,omitempty"` // branch|tag
	}

	Stages struct {
		Stage    *Stage    `json:"stage,omitempty"    yaml:"stage,omitempty"`
		Parallel []*Stages `json:"parallel,omitempty" yaml:"parallel,omitempty"`
	}

	Stage struct {
		ID          string      `json:"identifier,omitempty"   yaml:"identifier,omitempty"`
		Description string      `json:"description,omitempty"  yaml:"description,omitempty"`
		Name        string      `json:"name,omitempty"         yaml:"name,omitempty"`
		Spec        interface{} `json:"spec,omitempty"         yaml:"spec,omitempty"`
		Type        string      `json:"type,omitempty"         yaml:"type,omitempty"`
		Vars        []*Variable `json:"variables,omitempty"    yaml:"variables,omitempty"`
		When        *StageWhen  `json:"when,omitempty"         yaml:"when,omitempty"`
		Strategy    *Strategy   `json:"strategy,omitempty"     yaml:"strategy,omitempty"`
		TimeOut     string      `json:"timeout,omitempty"     yaml:"timeout,omitempty"`
	}

	StageWhen struct {
		PipelineStatus string `json:"pipelineStatus,omitempty" yaml:"pipelineStatus,omitempty"`
		Condition      string `json:"condition,omitempty" yaml:"condition,omitempty"`
	}

	Strategy struct {
		Matrix map[string]interface{} `json:"matrix,omitempty" yaml:"matrix,omitempty"`
	}

	// Infrastructure provides pipeline infrastructure.
	Infrastructure struct {
		Type string     `json:"type,omitempty"          yaml:"type,omitempty"`
		From string     `json:"useFromStage,omitempty"  yaml:"useFromStage,omitempty"` // this is also weird
		Spec *InfraSpec `json:"spec,omitempty"          yaml:"spec,omitempty"`
	}

	// InfraSpec describes pipeline infastructure.
	InfraSpec struct {
		Namespace string `json:"namespace,omitempty"    yaml:"namespace,omitempty"`
		Conn      string `json:"connectorRef,omitempty" yaml:"connectorRef,omitempty"`
	}

	Platform struct {
		OS   string `json:"os,omitempty"   yaml:"os,omitempty"`
		Arch string `json:"arch,omitempty" yaml:"arch,omitempty"`
	}

	Runtime struct {
		Type string      `json:"type,omitempty"   yaml:"type,omitempty"`
		Spec interface{} `json:"spec,omitempty"   yaml:"spec,omitempty"`
	}

	Variable struct {
		Name  string `json:"name,omitempty"  yaml:"name,omitempty"`
		Type  string `json:"type,omitempty"  yaml:"type,omitempty"` // Secret|Text
		Value string `json:"value,omitempty" yaml:"value,omitempty"`
	}

	Execution struct {
		Steps []Steps `json:"steps,omitempty" yaml:"steps,omitempty"` // Un-necessary
	}

	Steps struct {
		Step     *Step    `json:"step,omitempty" yaml:"step,omitempty"` // Un-necessary
		Parallel []*Steps `json:"parallel,omitempty" yaml:"parallel,omitempty"`
	}

	Step struct { // TODO missing failure strategies
		ID          string            `json:"identifier,omitempty"        yaml:"identifier,omitempty"`
		Description string            `json:"description,omitempty"       yaml:"description,omitempty"`
		Name        string            `json:"name,omitempty"              yaml:"name,omitempty"`
		Skip        string            `json:"skipCondition,omitempty"     yaml:"skipCondition,omitempty"`
		Spec        interface{}       `json:"spec,omitempty"              yaml:"spec,omitempty"`
		Type        string            `json:"type,omitempty"              yaml:"type,omitempty"`
		Env         map[string]string `json:"envVariables,omitempty"      yaml:"envVariables,omitempty"`
		Strategy    *Strategy         `json:"strategy,omitempty"     yaml:"strategy,omitempty"`
	}

	StepRun struct {
		Env             map[string]string `json:"envVariables,omitempty"    yaml:"envVariables,omitempty"`
		Command         string            `json:"command,omitempty"         yaml:"command,omitempty"`
		ConnRef         string            `json:"connectorRef,omitempty"    yaml:"connectorRef,omitempty"`
		Image           string            `json:"image,omitempty"           yaml:"image,omitempty"`
		ImagePullPolicy string            `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
		Outputs         []*Output         `json:"outputVariables,omitempty" yaml:"outputVariables,omitempty"`
		Privileged      bool              `json:"privileged,omitempty"      yaml:"privileged,omitempty"`
		Resources       *Resources        `json:"resources,omitempty"       yaml:"resources,omitempty"`
		RunAsUser       string            `json:"runAsUser,omitempty"       yaml:"runAsUser,omitempty"`
		Reports         *Report           `json:"reports,omitempty"         yaml:"reports,omitempty"`
		Shell           string            `json:"shell,omitempty"           yaml:"shell,omitempty"`
	}

	Report struct {
		Type string       `json:"type" yaml:"type,omitempty"` // JUnit|JUnit
		Spec *ReportJunit `json:"spec" yaml:"spec,omitempty"` // TODO
	}

	ReportJunit struct {
		Paths []string `json:"paths" yaml:"paths,omitempty"`
	}

	Service struct {
		ID   string       `json:"identifier,omitempty"   yaml:"identifier,omitempty"`
		Name string       `json:"name,omitempty"         yaml:"name,omitempty"`
		Type string       `json:"type,omitempty"         yaml:"type,omitempty"` // Service
		Desc string       `json:"description,omitempty"  yaml:"description,omitempty"`
		Spec *ServiceSpec `json:"spec,omitempty"         yaml:"spec,omitempty"`
	}

	Output struct {
		Name  string `json:"name,omitempty"      yaml:"name,omitempty"`
		Type  string `json:"type,omitempty" yaml:"type,omitempty"`
		Value string `json:"value,omitempty" yaml:"value,omitempty"`
	}

	ServiceSpec struct {
		Env        map[string]string `json:"envVariables,omitempty"   yaml:"envVariables,omitempty"`
		Entrypoint []string          `json:"entrypoint,omitempty"     yaml:"entrypoint,omitempty"`
		Args       []string          `json:"args,omitempty"           yaml:"args,omitempty"`
		Conn       string            `json:"connectorRef,omitempty"   yaml:"connectorRef,omitempty"`
		Image      string            `json:"image,omitempty"          yaml:"image,omitempty"`
		Resources  *Resources        `json:"resources,omitempty"      yaml:"resources,omitempty"`
	}

	FailureStrategy struct {
		OnFailure OnFailure `json:"onFailure,omitempty" yaml:"onFailure,omitempty"`
	}

	OnFailure struct {
		Errors []string
	}

	Resources struct {
		Limits Limits `json:"limits,omitempty" yaml:"limits,omitempty"`
	}

	Limits struct {
	}
)
