/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { useState, useEffect } from 'react';
import { Tag, Intent, Checkbox, InputGroup } from '@blueprintjs/core';

import { ExternalLink, HelpTooltip } from '@/components';
import { DOC_URL } from '@/release';

import * as S from './styled';

interface Props {
  entities: string[];
  transformation: any;
  setTransformation: React.Dispatch<React.SetStateAction<any>>;
  setHasError: React.Dispatch<React.SetStateAction<boolean>>;
}

export const GitLabTransformation = ({ entities, transformation, setTransformation, setHasError }: Props) => {
  const [useCustom, setUseCustom] = useState(false);

  useEffect(() => {
    if (transformation.deploymentPattern || transformation.productionPattern) {
      setUseCustom(true);
    } else {
      setUseCustom(false);
    }
  }, [transformation]);

  useEffect(() => {
    setHasError(useCustom && !transformation.deploymentPattern);
  }, [useCustom, transformation]);

  const handleChangeUseCustom = (e: React.FormEvent<HTMLInputElement>) => {
    const checked = (e.target as HTMLInputElement).checked;

    if (!checked) {
      setTransformation({
        ...transformation,
        deploymentPattern: '',
        productionPattern: '',
      });
    }

    setUseCustom(checked);
  };

  return (
    <S.Transformation>
      {entities.includes('CICD') && (
        <S.CICD>
          <h2>CI/CD</h2>
          <h3>
            <span>Deployment</span>
            <Tag minimal intent={Intent.PRIMARY} style={{ marginLeft: 8 }}>
              DORA
            </Tag>
          </h3>
          <p style={{ marginBottom: 16 }}>
            Use Regular Expression to define Deployments in DevLake in order to measure DORA metrics.{' '}
            <ExternalLink link={DOC_URL.PLUGIN.GITLAB.TRANSFORMATION}>Learn more</ExternalLink>
          </p>
          <div className="text">
            <Checkbox disabled checked />
            <span>Convert a GitLab Deployment to a DevLake Deployment </span>
          </div>
          <div className="sub-text">
            <span>If its environment name matches</span>
            <InputGroup
              style={{ width: 180, margin: '0 8px' }}
              placeholder="(?i)prod(.*)"
              value={transformation.envNamePattern}
              onChange={(e) =>
                setTransformation({
                  ...transformation,
                  envNamePattern: e.target.value,
                })
              }
            />
            <span>, this deployment is a ‘Production Deployment’</span>
          </div>
          <div className="text">
            <Checkbox checked={useCustom} onChange={handleChangeUseCustom} />
            <span>Convert a GitLab Pipeline as a DevLake Deployment when: </span>
          </div>
          <div className="sub-text">
            <span>
              Its branch/tag name or <strong>one of its jobs</strong> matches
            </span>
            <InputGroup
              style={{ width: 180, margin: '0 8px' }}
              placeholder="(deploy|push-image)"
              value={transformation.deploymentPattern ?? ''}
              onChange={(e) =>
                setTransformation({
                  ...transformation,
                  deploymentPattern: e.target.value,
                  productionPattern: !e.target.value ? '' : transformation.productionPattern,
                })
              }
            />
            <i style={{ color: '#E34040' }}>*</i>
            <HelpTooltip content="GitLab Pipelines: https://docs.gitlab.com/ee/ci/pipelines/" />
          </div>
          <div className="sub-text">
            <span>If the name also matches</span>
            <InputGroup
              style={{ width: 180, margin: '0 8px' }}
              placeholder="prod(.*)"
              value={transformation.productionPattern ?? ''}
              onChange={(e) =>
                setTransformation({
                  ...transformation,
                  productionPattern: e.target.value,
                })
              }
            />
            <span>, this deployment is a ‘Production Deployment’</span>
            <HelpTooltip content="If you leave this field empty, all Deployments will be tagged as in the Production environment. " />
          </div>
        </S.CICD>
      )}
    </S.Transformation>
  );
};
