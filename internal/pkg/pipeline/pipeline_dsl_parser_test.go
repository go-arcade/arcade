// Copyright 2025 Arcade Team
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

package pipeline

import (
	"testing"

	"github.com/go-arcade/arcade/pkg/log"
)

func TestDSLParser_Parse(t *testing.T) {
	logger := log.Logger{Log: log.GetLogger()}
	parser := NewDSLParser(logger)

	tests := []struct {
		name    string
		dslJSON string
		wantErr bool
	}{
		{
			name: "valid pipeline",
			dslJSON: `{
				"namespace": "test",
				"version": "1.0.0",
				"variables": {
					"ENV": "prod"
				},
				"jobs": [
					{
						"name": "build",
						"steps": [
							{
								"name": "checkout",
								"uses": "git@1.0.0",
								"action": "clone"
							}
						]
					}
				]
			}`,
			wantErr: false,
		},
		{
			name: "missing namespace",
			dslJSON: `{
				"version": "1.0.0",
				"jobs": []
			}`,
			wantErr: true,
		},
		{
			name: "missing jobs",
			dslJSON: `{
				"namespace": "test"
			}`,
			wantErr: true,
		},
		{
			name: "job without steps",
			dslJSON: `{
				"namespace": "test",
				"jobs": [
					{
						"name": "build"
					}
				]
			}`,
			wantErr: true,
		},
		{
			name: "step without uses",
			dslJSON: `{
				"namespace": "test",
				"jobs": [
					{
						"name": "build",
						"steps": [
							{
								"name": "checkout"
							}
						]
					}
				]
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.dslJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDSLParser_ToJSON(t *testing.T) {
	logger := log.Logger{Log: log.GetLogger()}
	parser := NewDSLParser(logger)

	pipeline := &Pipeline{
		Namespace: "test",
		Version:   "1.0.0",
		Jobs: []Job{
			{
				Name: "build",
				Steps: []Step{
					{
						Name: "checkout",
						Uses: "git@1.0.0",
					},
				},
			},
		},
	}

	jsonStr, err := parser.ToJSON(pipeline)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if jsonStr == "" {
		t.Error("ToJSON() returned empty string")
	}

	// Parse back to verify round-trip
	parsed, err := parser.Parse(jsonStr)
	if err != nil {
		t.Fatalf("Parse() after ToJSON() error = %v", err)
	}

	if parsed.Namespace != pipeline.Namespace {
		t.Errorf("Namespace = %v, want %v", parsed.Namespace, pipeline.Namespace)
	}
}

func TestValidator_Validate(t *testing.T) {
	logger := log.Logger{Log: log.GetLogger()}
	parser := NewDSLParser(logger)
	validator := NewValidator(parser)

	tests := []struct {
		name     string
		pipeline *Pipeline
		wantErr  bool
	}{
		{
			name: "valid pipeline",
			pipeline: &Pipeline{
				Namespace: "test",
				Version:   "1.0.0",
				Jobs: []Job{
					{
						Name: "build-job",
						Steps: []Step{
							{
								Name: "checkout",
								Uses: "git@1.0.0",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid namespace",
			pipeline: &Pipeline{
				Namespace: "test@invalid",
				Jobs: []Job{
					{
						Name: "build",
						Steps: []Step{
							{
								Name: "checkout",
								Uses: "git",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate job names",
			pipeline: &Pipeline{
				Namespace: "test",
				Jobs: []Job{
					{
						Name: "build",
						Steps: []Step{
							{
								Name: "checkout",
								Uses: "git",
							},
						},
					},
					{
						Name: "build",
						Steps: []Step{
							{
								Name: "checkout",
								Uses: "git",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timeout format",
			pipeline: &Pipeline{
				Namespace: "test",
				Jobs: []Job{
					{
						Name:    "build",
						Timeout: "invalid",
						Steps: []Step{
							{
								Name: "checkout",
								Uses: "git",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.pipeline)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
