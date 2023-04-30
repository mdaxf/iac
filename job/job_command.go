// Copyright 2020
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import (
	"context"
	"fmt"
	"html/template"

	"github.com/pkg/errors"

	admin "github.com/mdaxf/iac/framework/job/admin"
)

type listJobCommand struct{}

func (l *listJobCommand) Execute(params ...interface{}) *admin.Result {
	resultList := make([][]string, 0, len(globalJobManager.adminJobList))
	for tname, tk := range globalJobManager.adminJobList {
		result := []string{
			template.HTMLEscapeString(tname),
			template.HTMLEscapeString(tk.GetSpec(nil)),
			template.HTMLEscapeString(tk.GetStatus(nil)),
			template.HTMLEscapeString(tk.GetPrev(context.Background()).String()),
		}
		resultList = append(resultList, result)
	}

	return &admin.Result{
		Status:  200,
		Content: resultList,
	}
}

type runJobCommand struct{}

func (r *runJobCommand) Execute(params ...interface{}) *admin.Result {
	if len(params) == 0 {
		return &admin.Result{
			Status: 400,
			Error:  errors.New("job name not passed"),
		}
	}

	tn, ok := params[0].(string)

	if !ok {
		return &admin.Result{
			Status: 400,
			Error:  errors.New("parameter is invalid"),
		}
	}

	if t, ok := globalJobManager.adminJobList[tn]; ok {
		err := t.Run(context.Background())
		if err != nil {
			return &admin.Result{
				Status: 500,
				Error:  err,
			}
		}
		return &admin.Result{
			Status:  200,
			Content: t.GetStatus(context.Background()),
		}
	} else {
		return &admin.Result{
			Status: 400,
			Error:  errors.New(fmt.Sprintf("job with name %s not found", tn)),
		}
	}
}

func registerCommands() {
	admin.RegisterCommand("job", "list", &listJobCommand{})
	admin.RegisterCommand("job", "run", &runJobCommand{})
}
