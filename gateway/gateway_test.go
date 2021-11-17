// Copyright 2021 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/cerbos/cerbos-aws-lambda/test"
)

// implement process manager interface.
type testProcessManager struct{}

func (testProcessManager) StartProcess(_ context.Context, _, _, _ string) error { return nil }

func (testProcessManager) Started() bool { return true }

func (testProcessManager) StopProcess() error { return nil }

const localCerbosURL = "http://127.0.0.1" + httpListenAddr

var remoteCerbosURL = os.Getenv("REMOTE_CERBOS_URL")

func TestGateway_Invoke(t *testing.T) {
	ctx := context.Background()
	is := require.New(t)
	tests := loadTests(t, "checks/check_resource_set", "checks/check_resource_batch")
	is.NotEmpty(tests)

	cerbosURL := localCerbosURL
	if remoteCerbosURL != "" {
		cerbosURL = remoteCerbosURL
	}
	gw, err := NewGateway(cerbosURL)
	if remoteCerbosURL != "" {
		gw.processManager = testProcessManager{}
	}
	is.NoError(err)
	err = gw.StartProcess(ctx, pathToCerbos(t), test.PathToDir(t, ""), "conf.yml")
	is.NoError(err)
	defer func() {
		is.NoError(gw.StopProcess())
	}()
	for file, tt := range tests {
		batch := false
		resources := tt.CheckResourceSet
		if strings.HasSuffix(filepath.Dir(file), "batch") {
			batch = true
			resources = tt.CheckResourceBatch
		}
		t.Run(filepath.Base(file)+":"+tt.Description, func(t *testing.T) {
			is := require.New(t)
			input, err := json.Marshal(resources.Input)
			is.NoError(err)

			endpoint := "/api/check"
			if batch {
				endpoint = "/api/check_resource_batch"
			}
			payload, err := mkPayload(input, endpoint)
			is.NoError(err)

			got, err := gw.Invoke(ctx, payload)
			is.NoError(err)

			var response events.APIGatewayV2HTTPResponse
			err = json.Unmarshal(got, &response)
			is.NoError(err)
			if tt.WantError {
				is.NotEqual(response.StatusCode, 200)
			}
			is.Equal(tt.WantStatus.HTTPStatusCode, response.StatusCode, struct {
				request, response string
			}{request: string(input), response: response.Body})
			if resources.WantResponse != nil {
				body := make(map[string]interface{})
				err = json.Unmarshal([]byte(response.Body), &body)
				is.NoError(err)
				diff := cmp.Diff(resources.WantResponse, body, cmpopts.EquateEmpty(), cmpopts.SortSlices(func(a, b interface{}) bool {
					if a, ok := a.(string); ok {
						if b, ok := b.(string); ok {
							return a < b
						}
					}
					return false
				}))
				is.Empty(diff, "mismatch: -want +got")
			}
		})
	}
}

func pathToCerbos(t *testing.T) string {
	t.Helper()
	goOS := os.Getenv("GOOS")
	if goOS == "" {
		goOS = runtime.GOOS
	}
	goARCH := os.Getenv("GOARCH")
	if goARCH == "" {
		goARCH = runtime.GOARCH
	}
	arch := goARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	path := filepath.Join(test.PathToDir(t, ""),
		"../../.cerbos",
		fmt.Sprintf("%s_%s", strings.Title(goOS), arch),
		"cerbos")

	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.True(t, !stat.IsDir())

	return path
}

func mkPayload(input []byte, endpoint string) ([]byte, error) {
	var request events.APIGatewayV2HTTPRequest
	request.RawPath = "http://example.com" + endpoint
	request.RequestContext.HTTP.Method = "POST"
	request.Headers = map[string]string{
		"content-type": "application/json",
	}
	request.Body = string(input)
	return json.Marshal(request)
}

func loadTests(t *testing.T, dirs ...string) map[string]checkResources {
	t.Helper()
	tests := make(map[string]checkResources)

	for _, dir := range dirs {
		err := filepath.WalkDir(test.PathToDir(t, dir), func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".yaml" {
				return nil
			}
			// read contents of file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			test := checkResources{}
			err = yaml.Unmarshal(content, &test)
			if err != nil {
				return err
			}
			tests[path] = test
			return nil
		})
		require.NoError(t, err)
	}
	return tests
}

type checkResources struct {
	Description string `json:"description"`
	WantError   bool   `json:"wantError"`
	WantStatus  struct {
		HTTPStatusCode int `json:"httpStatusCode"`
	} `json:"wantStatus"`
	CheckResourceSet struct {
		Input        map[string]interface{} `json:"input"`
		WantResponse map[string]interface{} `json:"wantResponse"`
	} `json:"checkResourceSet"`
	CheckResourceBatch struct {
		Input        map[string]interface{} `json:"input"`
		WantResponse map[string]interface{} `json:"wantResponse"`
	} `json:"checkResourceBatch"`
}
