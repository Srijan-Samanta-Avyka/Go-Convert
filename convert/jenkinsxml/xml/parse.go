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

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Token Parser for Groovy Script

const (
	EOF = iota
	IDENTIFIER
	STRING
	KEYWORD
	NUMBER
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	COMMA
	NEWLINE
	DOT
	ERROR
	// Special characters
	EQUALS
	PLUS
	MINUS
	STAR
	SLASH
	// Additional keywords
	WHEN
	PARALLEL
	MATRIX
)

// Token struct to represent different types of tokens

// List of recognized keywords
var keywords = map[string]int{
	"stage":    KEYWORD,
	"steps":    KEYWORD,
	"when":     WHEN,
	"parallel": PARALLEL,
	"matrix":   MATRIX,
	"echo":     KEYWORD,
	"sh":       KEYWORD,
}

type Token struct {
	Value string
	Type  string
}

// Define a list of keywords

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*Project, error) {
	out := new(Project)

	// see https://github.com/golang/go/issues/25755
	// encoding/xml does not support XML 1.1, which jenkins uses
	//
	// TODO: this approach is likely brittle and will need to be revisited
	/*data, _ := io.ReadAll(r)
	res := strings.Replace(string(data), "<?xml version='1.1", "<?xml version='1.0", 1)*/
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&out)

	/*-------------- Parse The Groovy Script To Jenkins Struct --------------*/

	//pipeline, err := parseScript(out.Script)
	tokens := Tokenize(out.Script)
	pipeline := ParseScript(tokens)
	//fmt.Print(tokens)
	out.Pipeline = pipeline

	return out, err
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*Project, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*Project, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*Project, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parses the Groovy Script to Jenkins Struct

func Tokenize(input string) []Token {
	var tokens []string
	var currentToken []rune
	var Tokens []Token
	var Tok Token
	state := "default" // can be "default", "string_literal_single", "string_literal_double", "single_line_comment", or "multi_line_comment"
	var inSingleLineComment bool
	var inMultiLineComment bool

	for i := 0; i < len(input); i++ {
		ch := rune(input[i])

		switch state {
		case "default":
			if inSingleLineComment {
				// Skip rest of the line until new line
				if ch == '\n' {
					inSingleLineComment = false
				}
				continue
			}

			if inMultiLineComment {
				// Skip characters inside multi-line comments until we find the closing "*/"
				if i+1 < len(input) && input[i] == '*' && input[i+1] == '/' {
					inMultiLineComment = false
					i++ // Skip the '/' character
				}
				continue
			}

			// Handle spaces by adding the current token (if any)
			if unicode.IsSpace(ch) {
				if len(currentToken) > 0 {
					tokenValue := string(currentToken)
					tokens = append(tokens, tokenValue)
					Tok = Token{}
					Tok.Type = `IDENT`
					Tok.Value = tokenValue
					Tokens = append(Tokens, Tok)
					currentToken = nil
				}
				// Skip appending an empty token if currentToken is empty
			} else if ch == '\'' {
				// Start of single-quoted string literal
				if len(currentToken) > 0 {
					tokenValue := string(currentToken)
					tokens = append(tokens, tokenValue)
					Tok = Token{}
					Tok.Type = `IDENT`
					Tok.Value = tokenValue
					Tokens = append(Tokens, Tok)
					currentToken = nil
				}
				state = "string_literal_single"
				val1 := rune(input[i+1])
				val2 := rune(input[i+2])
				if val1 == '\'' && val2 == '\'' {
					state = "string_literal_triple"
					i += 2
				}
				currentToken = append(currentToken, ch)
			} else if ch == '"' {
				// Start of double-quoted string literal
				if len(currentToken) > 0 {
					tokenValue := string(currentToken)
					tokens = append(tokens, tokenValue)
					Tok = Token{}
					Tok.Type = `IDENT`
					Tok.Value = tokenValue
					Tokens = append(Tokens, Tok)
					currentToken = nil
				}
				state = "string_literal_double"
				val1 := rune(input[i+1])
				val2 := rune(input[i+2])
				if val1 == '"' && val2 == '"' {
					state = "string_literal_quadruple"
					i += 2
				}
				currentToken = append(currentToken, ch)
			} else if ch == '/' {
				// Possible start of comment
				if i+1 < len(input) && input[i+1] == '/' {
					inSingleLineComment = true
					i++ // Skip the next '/'
				} else if i+1 < len(input) && input[i+1] == '*' {
					inMultiLineComment = true
					i++ // Skip the next '*'
				} else {
					currentToken = append(currentToken, ch)
				}
			} else if ch == '(' || ch == ')' || ch == '{' || ch == '}' || ch == ',' || ch == '.' {
				// Treat special characters like parentheses and braces as separate tokens
				if len(currentToken) > 0 {
					tokenValue := string(currentToken)
					tokens = append(tokens, tokenValue)
					Tok = Token{}
					Tok.Type = `IDENT`
					Tok.Value = tokenValue
					Tokens = append(Tokens, Tok)
					currentToken = nil
				}
				// Add the special character as a separate token
				tokens = append(tokens, string(ch))
				Tok = Token{}
				Tok.Type = `IDENT`
				Tok.Value = string(ch)
				Tokens = append(Tokens, Tok)
			} else {
				// Normal token (word or other characters)
				currentToken = append(currentToken, ch)
			}

		case "string_literal_single":
			currentToken = append(currentToken, ch)
			if ch == '\'' {
				tokens = append(tokens, string(currentToken[1:len(currentToken)-1]))
				currentTokenVal := string(currentToken[1 : len(currentToken)-1])
				Tok = Token{}
				_, err := strconv.Atoi(currentTokenVal)
				if err == nil {
					// It's an integer
					Tok.Type = "integer"
				} else {
					normalizedVal := strings.ToLower(currentTokenVal) // Normalize case
					if normalizedVal == "true" || normalizedVal == "false" {
						Tok.Type = "boolean"
					} else {
						// It's a string
						Tok.Type = "string"
					}
				}
				Tok.Value = string(currentTokenVal)
				Tokens = append(Tokens, Tok)
				currentToken = nil
				state = "default"
			}

		case "string_literal_double":
			currentToken = append(currentToken, ch)
			if ch == '"' {
				// End of double-quoted string literal
				// Remove the first and last character (the quotes) before adding the token
				tokens = append(tokens, string(currentToken[1:len(currentToken)-1]))
				currentTokenVal := string(currentToken[1 : len(currentToken)-1])
				Tok = Token{}
				_, err := strconv.Atoi(currentTokenVal)
				if err == nil {
					// It's an integer
					Tok.Type = "integer"
				} else {
					normalizedVal := strings.ToLower(currentTokenVal) // Normalize case
					if normalizedVal == "true" || normalizedVal == "false" {
						Tok.Type = "boolean"
					} else {
						// It's a string
						Tok.Type = "string"
					}
				}
				Tok.Value = string(currentTokenVal)
				Tokens = append(Tokens, Tok)
				currentToken = nil
				state = "default"
			}

		case "string_literal_triple":
			currentToken = append(currentToken, ch)
			if ch == '\'' {
				// End of single-quoted string literal
				// Remove the first and last character (the quotes) before adding the token
				tokens = append(tokens, string(currentToken[1:len(currentToken)-1]))
				currentTokenVal := string(currentToken[1 : len(currentToken)-1])
				Tok = Token{}
				_, err := strconv.Atoi(currentTokenVal)
				if err == nil {
					// It's an integer
					Tok.Type = "integer"
				} else {
					normalizedVal := strings.ToLower(currentTokenVal) // Normalize case
					if normalizedVal == "true" || normalizedVal == "false" {
						Tok.Type = "boolean"
					} else {
						// It's a string
						Tok.Type = "string"
					}
				}
				Tok.Value = string(currentTokenVal)
				Tokens = append(Tokens, Tok)
				currentToken = nil
				state = "default"
				i += 2
			}
		case "string_literal_quadruple":
			currentToken = append(currentToken, ch)
			if ch == '"' {
				// End of single-quoted string literal
				// Remove the first and last character (the quotes) before adding the token
				tokens = append(tokens, string(currentToken[1:len(currentToken)-1]))
				currentTokenVal := string(currentToken[1 : len(currentToken)-1])
				Tok = Token{}
				_, err := strconv.Atoi(currentTokenVal)
				if err == nil {
					// It's an integer
					Tok.Type = "integer"
				} else {
					normalizedVal := strings.ToLower(currentTokenVal) // Normalize case
					if normalizedVal == "true" || normalizedVal == "false" {
						Tok.Type = "boolean"
					} else {
						// It's a string
						Tok.Type = "string"
					}
				}
				Tok.Value = string(currentTokenVal)
				Tokens = append(Tokens, Tok)
				currentToken = nil
				state = "default"
				i += 2
			}
		}

	}

	// Append any leftover token
	if len(currentToken) > 0 {
		tokenValue := string(currentToken)
		tokens = append(tokens, tokenValue)
		Tok = Token{}
		Tok.Type = `STRING`
		Tok.Value = string(currentToken)
		Tokens = append(Tokens, Tok)
	}

	return Tokens
}

func ParseScript(Tokens []Token) Pipeline {
	pipeline := Pipeline{
		Environment: Environment{Variables: make(map[string]Variable)},
	}
	stage := Stage{Environment: Environment{Variables: make(map[string]Variable)}}
	parallelInnerStage := Stage{}
	matrixInnerStage := Stage{}
	parallelBlock := false
	martrixBlock := false
	post := Post{}
	when := When{}
	keywordMap := map[string]bool{
		"pipeline":    true,
		"stage":       true,
		"stages":      true,
		"steps":       true,
		"agent":       true,
		"label":       true,
		"docker":      true,
		"none":        true,
		"echo":        true,
		"sh":          true,
		"script":      true,
		"environment": true,
		"always":      true,
		"success":     true,
		"failure":     true,
		"unstable":    true,
		"post":        true,
		"triggers":    true,
		"options":     true,
		"retry":       true,
		"tools":       true,
		"when":        true,
		"not":         true,
		"anyOf":       true,
		"allOf":       true,
		"branch":      true,
		"tag":         true,
		"buildingTag": true,
		"parallel":    true,
		"matrix":      true,
		"axis":        true,
		"name":        true,
		"values":      true,
		"exclude":     true,
		"jdk":         true,
		"maven":       true,
		"parameters":  true,
		"timeout":     true,
		"unit":        true,
		"time":        true,
		"upstream":    true,
	}

	var commentString = ""
	var currentBlock string
	var currentKeyword string
	var scripts []Script
	var parenthesisKeyword []string
	var parenthesis []bool
	var matrix = make(map[string][]interface{})
	var excludeMatrix = make(map[string]interface{})
	// matrix["exclude"] = append(matrix["exclude"], map[string]string{
	// 	"key1": "value1",
	// 	"key2": "value2",
	// })
	// matrix["exclude"] = append(matrix["exclude"], map[string]string{
	// 	"key1": "value1",
	// 	"key2": "value2",
	// })
	//var parenthesisInScriptComment int
	for i := 0; i < len(Tokens); i++ {
		// Handling the Open Braces That are Not Part of Keyword
		// if Tokens[i+1].Value == "{" {
		// 	parenthesis = append(parenthesis, false)
		// 	continue
		// }

		if _, exists := keywordMap[Tokens[i].Value]; exists {
			currentKeyword = Tokens[i].Value
			if currentKeyword == "stage" {
				parenthesisKeyword = append(parenthesisKeyword, Tokens[i].Value)
				currentBlock = Tokens[i].Value
			} else if Tokens[i+1].Value == "{" {
				parenthesisKeyword = append(parenthesisKeyword, Tokens[i].Value)
				parenthesis = append(parenthesis, true)
				currentBlock = Tokens[i].Value
				i++
			}
		} else if Tokens[i].Value == "{" {
			parenthesis = append(parenthesis, false)
			continue
		}

		//------------------- Resolving closing quotes -------------------

		if Tokens[i].Value == "}" {

			if len(parenthesis) > 0 && parenthesis[len(parenthesis)-1] == false {
				parenthesis = parenthesis[:len(parenthesis)-1]
				continue
			}

			if currentBlock == "stage" && stage.Name != "" {
				if parallelBlock {
					stage.Parallel.Stages = append(stage.Parallel.Stages, parallelInnerStage)
					parallelInnerStage = Stage{Environment: Environment{Variables: make(map[string]Variable)}}
				} else if martrixBlock {
					stage.Parallel.Stages = append(stage.Parallel.Stages, matrixInnerStage)
					matrixInnerStage = Stage{Environment: Environment{Variables: make(map[string]Variable)}}
				} else {
					pipeline.Stages = append(pipeline.Stages, stage)
					stage = Stage{Environment: Environment{Variables: make(map[string]Variable)}}
				}
			}

			if currentBlock == "exclude" {
				matrix["exclude"] = append(matrix["exclude"], excludeMatrix)
				excludeMatrix = make(map[string]interface{})
			}

			if currentBlock == "post" {
				if parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
					if parallelBlock {
						parallelInnerStage.Post = post
					} else if martrixBlock {
						matrixInnerStage.Post = post
					} else {
						stage.Post = post
					}
				} else {
					pipeline.Post = post
				}
				post = Post{}
			}

			if currentBlock == "parallel" {
				parallelBlock = false
			}

			if currentBlock == "matrix" {
				martrixBlock = false
				matrix = make(map[string][]interface{})
			}

			if currentBlock == "script" {
				// if parenthesisInScriptComment != 0 {
				// 	commentString = commentString + "}"
				// 	parenthesisInScriptComment--
				// 	continue
				// }
				if commentString != "" {
					script := Script{Comment: commentString}
					scripts = append(scripts, script)

					// step := Steps{Script: script}
					// stage.Steps = append(stage.Steps, step)
				}
				step := Steps{Script: scripts}
				scripts = nil
				if parallelBlock {
					parallelInnerStage.Steps = append(parallelInnerStage.Steps, step)
				} else if martrixBlock {
					matrixInnerStage.Steps = append(matrixInnerStage.Steps, step)
				} else {
					stage.Steps = append(stage.Steps, step)
				}

			}

			if currentBlock == "when" {
				if parallelBlock {
					parallelInnerStage.When = when
				} else if martrixBlock {
					matrixInnerStage.When = when
				} else {
					stage.When = when
				}
				when = When{}
			}

			parenthesisKeyword = parenthesisKeyword[:len(parenthesisKeyword)-1]
			parenthesis = parenthesis[:len(parenthesis)-1]
			if len(parenthesisKeyword) > 0 {
				currentBlock = parenthesisKeyword[len(parenthesisKeyword)-1]
			}
			continue
		}

		if currentBlock == "agent" {
			if currentKeyword == "label" {
				i++
				pipeline.Agent.Label = Tokens[i].Value
			}
			if currentKeyword == "docker" {
				i++
				pipeline.Agent.Docker = Tokens[i].Value
			}
			if currentKeyword == "none" {
				i++
				pipeline.Agent.None = Tokens[i].Value
			}
		}

		if currentBlock == "environment" {
			for {
				if Tokens[i].Value == "=" {
					if Tokens[i+1].Type == "string" || Tokens[i+1].Type == "integer" || Tokens[i+1].Type == "boolean" {
						if len(parenthesisKeyword) > 2 && parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
							stage.Environment.Variables[Tokens[i-1].Value] = Variable{Value: Tokens[i+1].Value, TypeValue: Tokens[i+1].Type}
						} else {
							pipeline.Environment.Variables[Tokens[i-1].Value] = Variable{Value: Tokens[i+1].Value, TypeValue: Tokens[i+1].Type}
						}
					} else {
						if parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
							stage.Environment.Variables[Tokens[i-1].Value] = Variable{Value: "changeMe", TypeValue: Tokens[i+1].Type}
						} else {
							pipeline.Environment.Variables[Tokens[i-1].Value] = Variable{Value: "changeMe", TypeValue: Tokens[i+1].Type}
						}
					}
				}
				if Tokens[i+1].Value == "}" {
					break
				}
				i++
			}
		}
		//------------------- Getting Values For Parallel Block -------------------

		if currentBlock == "parallel" {

			parallelBlock = true

		}

		//------------------- Getting Values For Matrix Block -------------------

		if currentBlock == "matrix" {

			martrixBlock = true

		}

		//------------------- Getting Values For Axis Block -------------------

		if currentBlock == "axis" {
			if parenthesisKeyword[len(parenthesisKeyword)-2] == "exclude" {
				if currentKeyword == "name" {
					i++
					axisKeyName := Tokens[i].Value
					excludeMatrix[axisKeyName] = nil
					i++
					if Tokens[i].Value == "values" {
						i++
						excludeMatrix[axisKeyName] = Tokens[i].Value
					}
				}

			} else {
				if currentKeyword == "name" {
					i++
					axisKeyName := Tokens[i].Value
					matrix[axisKeyName] = nil
					i++
					if Tokens[i].Value == "values" {
						for {
							i++
							if Tokens[i].Value != "," {
								matrix[axisKeyName] = append(matrix[axisKeyName], Tokens[i].Value)
							}
							if Tokens[i+1].Value == "}" {
								break
							}
						}
					}
				}
			}
		}

		//------------------- Getting Values For Stage Block -------------------

		if currentBlock == "stage" {
			if currentKeyword == "stage" {
				if Tokens[i+1].Value == "(" {
					i += 2
					if parallelBlock {
						parallelInnerStage.Name = Tokens[i].Value
					} else if martrixBlock {
						matrixInnerStage.Name = Tokens[i].Value
						matrixInnerStage.Matrix = matrix
					} else {
						stage.Name = Tokens[i].Value
					}
					i++
				} else {
					i++
					if parallelBlock {
						parallelInnerStage.Name = Tokens[i].Value
					} else if martrixBlock {
						matrixInnerStage.Name = Tokens[i].Value
						matrixInnerStage.Matrix = matrix
					} else {
						stage.Name = Tokens[i].Value
					}
				}
				i++
				parenthesis = append(parenthesis, true)
			}
			//fmt.Print("Test")
		}

		//------------------- Getting Values For Steps Block -------------------

		if currentBlock == "steps" {
			if currentKeyword == "sh" {
				i++
				step := Steps{Shell: Tokens[i].Value}
				if parallelBlock {
					parallelInnerStage.Steps = append(parallelInnerStage.Steps, step)
				} else if martrixBlock {
					matrixInnerStage.Steps = append(matrixInnerStage.Steps, step)
				} else {
					stage.Steps = append(stage.Steps, step)
				}
			} else if currentKeyword == "echo" {
				i++
				step := Steps{Echo: Tokens[i].Value}
				if parallelBlock {
					parallelInnerStage.Steps = append(parallelInnerStage.Steps, step)
				} else if martrixBlock {
					matrixInnerStage.Steps = append(matrixInnerStage.Steps, step)
				} else {
					stage.Steps = append(stage.Steps, step)
				}
			}
		}

		//------------------- Getting Values For Script Block -------------------

		if currentBlock == "script" {
			if currentKeyword == "echo" {
				if commentString != "" {
					script := Script{Comment: commentString}
					scripts = append(scripts, script)
					// step := Steps{Script: script}
					// stage.Steps = append(stage.Steps, step)
					commentString = ""
				}
				i++
				script := Script{Echo: Tokens[i].Value}
				scripts = append(scripts, script)
				//step := Steps{Script: script}
				// if parallelBlock {
				// 	parallelInnerStage.Steps = append(parallelInnerStage.Steps, step)
				// } else {
				// 	stage.Steps = append(stage.Steps, step)
				// }
				currentKeyword = ""
			} else if currentKeyword == "sh" {
				if commentString != "" {
					script := Script{Comment: commentString}
					scripts = append(scripts, script)
					// step := Steps{Script: script}
					// stage.Steps = append(stage.Steps, step)
					commentString = ""
				}
				i++
				script := Script{Shell: Tokens[i].Value}
				scripts = append(scripts, script)

				//step := Steps{Script: script}
				// if parallelBlock {
				// 	parallelInnerStage.Steps = append(parallelInnerStage.Steps, step)
				// } else {
				// 	stage.Steps = append(stage.Steps, step)
				// }
				currentKeyword = ""
			} else if currentKeyword == "" {
				commentString = commentString + Tokens[i].Value
				// if Tokens[i].Value == "{" {
				// 	//parenthesisInScriptComment++
				// }

			}

		}

		//-------------------- Data For When Block -------------------

		if currentBlock == "when" {
			if currentKeyword == "branch" {
				i++
				when.Branch = Tokens[i].Value
			} else if currentKeyword == "tag" {
				i++
				when.Tag = Tokens[i].Value
			} else if currentKeyword == "buildingTag" {
				when.BuildingTag = true
			} else if currentKeyword == "changeRequest" {
				// Create a new ChangeRequest object if it's not initialized
				changeRequest := ChangeRequest{}
				changeRequest.Enabled = true

				// Loop through the tokens to parse the attributes of changeRequest
				for i < len(Tokens) {
					if Tokens[i].Value == "id:" {
						i++ // Skip "id:"
						if Tokens[i].Type == "string" {
							changeRequest.ID = Tokens[i].Value
						}
					} else if Tokens[i].Value == "target:" {
						i++ // Skip "target:"
						if Tokens[i].Type == "string" {
							changeRequest.Target = Tokens[i].Value
						}
					} else if Tokens[i].Value == "branch:" {
						i++ // Skip "branch:"
						if Tokens[i].Type == "string" {
							changeRequest.Branch = Tokens[i].Value
						}
					} else if Tokens[i].Value == "fork:" {
						i++ // Skip "fork:"
						if Tokens[i].Type == "string" {
							changeRequest.Fork = Tokens[i].Value
						}
					} else if Tokens[i].Value == "url:" {
						i++ // Skip "url:"
						if Tokens[i].Type == "string" {
							changeRequest.URL = Tokens[i].Value
						}
					} else if Tokens[i].Value == "title:" {
						i++ // Skip "title:"
						if Tokens[i].Type == "string" {
							changeRequest.Title = Tokens[i].Value
						}
					} else if Tokens[i].Value == "author:" {
						i++ // Skip "author:"
						if Tokens[i].Type == "string" {
							changeRequest.Author = Tokens[i].Value
						}
					} else if Tokens[i].Value == "authorDisplayName:" {
						i++ // Skip "authorDisplayName:"
						if Tokens[i].Type == "string" {
							changeRequest.AuthorDisplayName = Tokens[i].Value
						}
					} else if Tokens[i].Value == "authorEmail:" {
						i++ // Skip "authorEmail:"
						if Tokens[i].Type == "string" {
							// changeRequest.AuthorEmail = Tokens[i].Value
							changeRequest.AuthorEmail = strings.Trim(Tokens[i].Value, "'")
						}
					}

					// Stop parsing if we encounter a closing brace or another block
					if Tokens[i].Value == "}" || (i+1 < len(Tokens) && keywordMap[Tokens[i+1].Value]) {
						break
					}

					i++ // Move to next token
				}

				// Store the parsed changeRequest in the when block
				when.Changerequest = changeRequest
			} else if currentKeyword == "changelog" {
				i++
				when.Changelog = Tokens[i].Value
			}
			// } else if currentKeyword == "not" {
			//  // Parse the nested condition inside 'not'
			//  when.NotCondition = parseCondition(Tokens[i+1:]) // You'll need a helper function to parse nested conditions
			//  i++ // Skip to next token after the 'not'
			// } else if currentKeyword == "allOf" {
			//  // Parse the nested conditions inside 'allOf'
			//  when.AllOfConditions = parseConditions(Tokens[i+1:]) // You'll need a helper function to parse nested conditions
			//  i++ // Skip to next token after the 'allOf'
			// } else if currentKeyword == "anyOf" {
			//  // Parse the nested conditions inside 'anyOf'
			//  when.AnyOfConditions = parseConditions(Tokens[i+1:]) // You'll need a helper function to parse nested conditions
			//  i++ // Skip to next token after the 'anyOf'
			// }
		}

		if currentBlock == "not" {
			if currentKeyword == "branch" {
				i++
				condition := Condition{Branch: Tokens[i].Value}
				when.NotCondition = append(when.NotCondition, condition)
				currentKeyword = ""
			} else if currentKeyword == "tag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.NotCondition = append(when.NotCondition, condition)
			} else if currentKeyword == "buildingTag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.NotCondition = append(when.NotCondition, condition)
			}

		}

		if currentBlock == "allOf" {
			if currentKeyword == "branch" {
				i++
				condition := Condition{Branch: Tokens[i].Value}
				when.AllOfConditions = append(when.AllOfConditions, condition)
				currentKeyword = ""
			} else if currentKeyword == "tag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.AllOfConditions = append(when.AllOfConditions, condition)
			} else if currentKeyword == "buildingTag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.AllOfConditions = append(when.AllOfConditions, condition)
			}

		}

		if currentBlock == "anyOf" {
			if currentKeyword == "branch" {
				i++
				condition := Condition{Branch: Tokens[i].Value}
				when.AnyOfConditions = append(when.AnyOfConditions, condition)
				currentKeyword = ""
			} else if currentKeyword == "tag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.AnyOfConditions = append(when.AnyOfConditions, condition)
			} else if currentKeyword == "buildingTag" {
				condition := Condition{Tag: Tokens[i].Value}
				when.AnyOfConditions = append(when.AnyOfConditions, condition)
			}

		}

		/*------------------- Getting Values For Triggers Block -------------------*/

		//-------------------- Getting Values For Tools Block -------------------*/

		if currentBlock == "tools" {
			if currentKeyword == "jdk" {
				i++
				pipeline.Tools.Jdk = Tokens[i].Value
			}
			if currentKeyword == "maven" {
				i++
				pipeline.Tools.Maven = Tokens[i].Value
			}
		}

		//------------------- Getting Values For Always Block -------------------

		if currentBlock == "always" {
			if currentKeyword == "echo" {
				i++
				post.Always.Echo = Tokens[i].Value
			}
		}

		//------------------- Getting Values For Success Block -------------------

		if currentBlock == "success" {
			if currentKeyword == "echo" {
				i++
				post.Success.Echo = Tokens[i].Value
			}
		}

		//------------------- Getting Values For Failure Block -------------------

		if currentBlock == "failure" {
			if currentKeyword == "echo" {
				post.Failure.Echo = Tokens[i].Value
			}
		}

		if currentBlock == "options" {
			if currentKeyword == "retry" {
				i++
				if Tokens[i].Value == "(" {
					i++
				}
				if len(parenthesisKeyword) > 2 && parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
					stage.Options.Retry = Tokens[i].Value

				} else {
					pipeline.Options.Retry = Tokens[i].Value
				}
				i++
			} else if currentKeyword == "timeout" {
				if Tokens[i].Value == "(" {
					i++
				}

				for i < len(Tokens) && Tokens[i].Value != ")" {
					// Handle "time" token
					fmt.Print(Tokens[i].Value)
					if Tokens[i].Value == "time:" {
						i++ // Skip "time"
						if i < len(Tokens) && Tokens[i].Value == ":" {
							i++ // Skip ":"
						}
						if len(parenthesisKeyword) > 2 && parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
							stage.Options.TimeOut.Time = Tokens[i].Value
						} else {
							pipeline.Options.TimeOut.Time = Tokens[i].Value
						}
					}

					if Tokens[i].Value == "unit:" {
						i++ // Skip "unit"
						if i < len(Tokens) && Tokens[i].Value == ":" {
							i++ // Skip ":"
						}
						// Ensure we capture the unit value correctly
						if i < len(Tokens) && (Tokens[i].Value == "MINUTES" || Tokens[i].Value == "SECONDS") {
							if parenthesisKeyword[len(parenthesisKeyword)-2] == "stage" {
								stage.Options.TimeOut.Unit = Tokens[i].Value
							} else {
								pipeline.Options.TimeOut.Unit = Tokens[i].Value
							}
						}
					}
					// if Tokens[i].Value == "," || Tokens[i].Value == ":" {
					// 	i++ // Skip commas and colons
					// 	continue
					// }
					if Tokens[i].Value == ")" {
						break
					}
					i++

				}
			} else if currentKeyword == "upstream" {
				i++
				pipeline.Options.Upstream = Tokens[i].Value
			}
		}

		if currentBlock == "parameters" {
			if currentKeyword == "parameters" {
				i++ // Skip the "parameters" keyword
				if Tokens[i].Value == "(" {
					i++ // Skip the opening "(" token
				}

				// Iterate through the tokens within the parentheses
				for Tokens[i].Value != "}" {
					param := Parameter{}

					// Check for each parameter type (string, text, booleanParam, choice, password)
					if Tokens[i].Value == "string" || Tokens[i].Value == "text" ||
						Tokens[i].Value == "booleanParam" || Tokens[i].Value == "choice" || Tokens[i].Value == "password" {

						param.Type = Tokens[i].Value
						i++
						if Tokens[i].Value == "(" {
							i++
						}

						// Handling Boolean Value For Boolean Param
						if param.Type == "booleanParam" {
							param.BooleanValue = true
						}

						for Tokens[i].Value != ")" {
							if Tokens[i].Value == "name:" {
								i++ // Skip "name"
								if Tokens[i].Value == ":" {
									i++ // Skip ":"
								}
								param.Name = Tokens[i].Value
							}

							// Parse the defaultValue parameter
							if Tokens[i].Value == "defaultValue:" {
								i++ // Skip "defaultValue"
								if Tokens[i].Value == ":" {
									i++ // Skip ":"
								}
								param.DefaultValue = Tokens[i].Value
							}

							// Parse the description parameter
							if Tokens[i].Value == "description:" {
								i++
								if Tokens[i].Value == ":" {
									i++
								}
								param.Description = Tokens[i].Value
							}
							if Tokens[i].Value == "choices:" && param.Type == "choice" {
								i++ // Skip "choices"
								if Tokens[i].Value == ":" {
									i++ // Skip ":"
								}
								if Tokens[i].Value == "[" {
									i++ // Skip "[" token
									for Tokens[i].Value != "]" {
										param.Choices = append(param.Choices, Tokens[i].Value)
										i++
										if Tokens[i].Value == "," {
											i++ // Skip "," separator
										}
									}
								}
							}

							i++
						}
						i++
						pipeline.Parameters.Params = append(pipeline.Parameters.Params, param)
					}

				}

			}
			parenthesisKeyword = parenthesisKeyword[:len(parenthesisKeyword)-1]
			parenthesis = parenthesis[:len(parenthesis)-1]
			if len(parenthesisKeyword) > 0 {
				currentBlock = parenthesisKeyword[len(parenthesisKeyword)-1]
			}
			continue
		}
	}

	return pipeline
}
