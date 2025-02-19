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
	"github.com/spf13/cast"
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/jenkins/models"
)

var ConvertBuildsToCicdTasksMeta = plugin.SubTaskMeta{
	Name:             "convertBuildsToCICD",
	EntryPoint:       ConvertBuildsToCicdTasks,
	EnabledByDefault: true,
	Description:      "convert builds to cicd",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD},
}

func ConvertBuildsToCicdTasks(taskCtx plugin.SubTaskContext) (err errors.Error) {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*JenkinsTaskData)
	clauses := []dal.Clause{
		dal.From("_tool_jenkins_builds"),
		dal.Where(`_tool_jenkins_builds.connection_id = ?
						and _tool_jenkins_builds.job_path = ?
						and _tool_jenkins_builds.job_name = ?`,
			data.Options.ConnectionId, data.Options.JobPath, data.Options.JobName),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	buildIdGen := didgen.NewDomainIdGenerator(&models.JenkinsBuild{})
	jobIdGen := didgen.NewDomainIdGenerator(&models.JenkinsJob{})

	converter, err := api.NewDataConverter(api.DataConverterArgs{
		InputRowType: reflect.TypeOf(models.JenkinsBuild{}),
		Input:        cursor,
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Params: JenkinsApiParams{
				ConnectionId: data.Options.ConnectionId,
				FullName:     data.Options.JobFullName,
			},
			Ctx:   taskCtx,
			Table: RAW_BUILD_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			jenkinsBuild := inputRow.(*models.JenkinsBuild)
			var durationMillis float64
			if jenkinsBuild.Duration > 0 {
				durationMillis = jenkinsBuild.Duration
			} else {
				durationMillis = 0
			}
			durationSec := durationMillis / 1e3

			jenkinsPipelineStatus := devops.GetStatusCommon(&devops.StatusRuleCommon[bool]{
				InProgress: []bool{true},
				Done:       []bool{false},
				Default:    devops.STATUS_OTHER,
			}, jenkinsBuild.Building)
			jenkinsPipelineResult := devops.RESULT_DEFAULT
			if !jenkinsBuild.Building {
				jenkinsPipelineResult = devops.GetResult(&devops.ResultRule{
					Success: []string{SUCCESS},
					Failure: []string{FAILURE, ABORTED},
					Default: devops.RESULT_DEFAULT,
				}, jenkinsBuild.Result)
			}
			var jenkinsPipelineFinishedDate *time.Time
			results := make([]interface{}, 0)

			if jenkinsPipelineStatus == devops.STATUS_DONE {
				finishTime := jenkinsBuild.StartTime.Add(time.Duration(int64(durationMillis) * int64(time.Millisecond)))
				jenkinsPipelineFinishedDate = &finishTime
			}
			jenkinsPipeline := &devops.CICDPipeline{
				DomainEntity: domainlayer.DomainEntity{
					Id: buildIdGen.Generate(jenkinsBuild.ConnectionId, jenkinsBuild.FullName),
				},
				Name:           jenkinsBuild.FullName,
				Result:         jenkinsPipelineResult,
				OriginalResult: jenkinsBuild.Result,
				Status:         jenkinsPipelineStatus,
				OriginalStatus: cast.ToString(jenkinsBuild.Building),
				FinishedDate:   jenkinsPipelineFinishedDate,
				DurationSec:    durationSec,
				CreatedDate:    jenkinsBuild.StartTime,
				CicdScopeId:    jobIdGen.Generate(jenkinsBuild.ConnectionId, data.Options.JobFullName),
				Type:           data.RegexEnricher.ReturnNameIfMatched(devops.DEPLOYMENT, jenkinsBuild.FullName),
				Environment:    data.RegexEnricher.ReturnNameIfOmittedOrMatched(devops.PRODUCTION, jenkinsBuild.FullName),
			}
			jenkinsPipeline.RawDataOrigin = jenkinsBuild.RawDataOrigin
			results = append(results, jenkinsPipeline)

			if !jenkinsBuild.HasStages {
				jenkinsTask := &devops.CICDTask{
					DomainEntity: domainlayer.DomainEntity{
						Id: buildIdGen.Generate(jenkinsBuild.ConnectionId, jenkinsBuild.FullName),
					},
					Name:           data.Options.JobFullName,
					Result:         jenkinsPipelineResult,
					Status:         jenkinsPipelineStatus,
					OriginalResult: jenkinsBuild.Result,
					OriginalStatus: cast.ToString(jenkinsBuild.Building),
					DurationSec:    durationSec,
					StartedDate:    jenkinsBuild.StartTime,
					FinishedDate:   jenkinsPipelineFinishedDate,
					CicdScopeId:    jobIdGen.Generate(jenkinsBuild.ConnectionId, data.Options.JobFullName),
					Type:           data.RegexEnricher.ReturnNameIfMatched(devops.DEPLOYMENT, jenkinsBuild.FullName),
					Environment:    data.RegexEnricher.ReturnNameIfOmittedOrMatched(devops.PRODUCTION, jenkinsBuild.FullName),
					PipelineId:     buildIdGen.Generate(jenkinsBuild.ConnectionId, jenkinsBuild.FullName),
				}
				results = append(results, jenkinsTask)

			}

			return results, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
