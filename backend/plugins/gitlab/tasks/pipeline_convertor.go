/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/core/errors"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
)

func init() {
	RegisterSubtaskMeta(&ConvertPipelineMeta)
}

var ConvertPipelineMeta = plugin.SubTaskMeta{
	Name:             "convertPipelines",
	EntryPoint:       ConvertPipelines,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitlab_pipeline into domain layer table pipeline",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD},
	Dependencies:     []*plugin.SubTaskMeta{&ConvertCommitsMeta},
}

func ConvertPipelines(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*GitlabTaskData)

	cursor, err := db.Cursor(dal.From(models.GitlabPipeline{}),
		dal.Where("project_id = ? and connection_id = ?", data.Options.ProjectId, data.Options.ConnectionId))
	if err != nil {
		return err
	}
	defer cursor.Close()

	pipelineIdGen := didgen.NewDomainIdGenerator(&models.GitlabPipeline{})
	projectIdGen := didgen.NewDomainIdGenerator(&models.GitlabProject{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(models.GitlabPipeline{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: models.GitlabApiParams{
				ConnectionId: data.Options.ConnectionId,
				ProjectId:    data.Options.ProjectId,
			},
			Table: RAW_PIPELINE_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			gitlabPipeline := inputRow.(*models.GitlabPipeline)

			startedAt := time.Now()
			if gitlabPipeline.StartedAt != nil {
				startedAt = *gitlabPipeline.StartedAt
			} else if gitlabPipeline.GitlabCreatedAt != nil {
				startedAt = *gitlabPipeline.GitlabCreatedAt
			}

			domainPipeline := &devops.CICDPipeline{
				DomainEntity: domainlayer.DomainEntity{
					Id: pipelineIdGen.Generate(data.Options.ConnectionId, gitlabPipeline.GitlabId),
				},
				Name: pipelineIdGen.Generate(data.Options.ConnectionId, gitlabPipeline.GitlabId),
				Result: devops.GetResult(&devops.ResultRule{
					Success: []string{StatusSuccess, StatusCompleted},
					Failure: []string{StatusFailed, StatusCanceled},
					Default: devops.RESULT_DEFAULT,
				}, gitlabPipeline.Status),
				Status: devops.GetStatus(&devops.StatusRule{
					Done:       []string{StatusSuccess, StatusCompleted, StatusFailed, StatusCanceled},
					InProgress: []string{StatusRunning, StatusWaitingForResource, StatusPending, StatusPreparing},
					Default:    devops.STATUS_OTHER,
				}, gitlabPipeline.Status),
				OriginalStatus: gitlabPipeline.Status,
				CreatedDate:    startedAt,
				FinishedDate:   gitlabPipeline.GitlabUpdatedAt,
				CicdScopeId:    projectIdGen.Generate(data.Options.ConnectionId, gitlabPipeline.ProjectId),
				Environment:    gitlabPipeline.Environment,
				Type:           gitlabPipeline.Type,
			}

			// rebuild the FinishedDate and DurationSec by Status
			if domainPipeline.Status != devops.STATUS_DONE {
				domainPipeline.FinishedDate = nil
				domainPipeline.DurationSec = 0
			} else if domainPipeline.FinishedDate != nil {
				durationTime := domainPipeline.FinishedDate.Sub(startedAt)
				domainPipeline.DurationSec = durationTime.Seconds()
			}

			return []interface{}{
				domainPipeline,
			}, nil
		},
	})

	if err != nil {
		return err
	}

	return converter.Execute()
}
