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

import { createBrowserRouter, Navigate, json } from 'react-router-dom';

import {
  ConnectionDetailPage,
  ProjectHomePage,
  ProjectDetailPage,
  BlueprintHomePage,
  BlueprintDetailPage,
  BlueprintConnectionDetailPage,
} from '@/pages';
import { Layout, loader as layoutLoader } from '@/routes/layout';
import { Error, ErrorEnum } from '@/routes/error';
import { Connections } from '@/routes/connection';
import { Pipelines, Pipeline } from '@/routes/pipeline';
import { ApiKeys } from '@/routes/api-keys';

export const router = createBrowserRouter([
  {
    path: 'db-migrate',
    element: <></>,
    loader: () => {
      throw json({ error: ErrorEnum.NEEDS_DB_MIRGATE }, { status: 428 });
    },
    errorElement: <Error />,
  },
  {
    path: '/',
    element: <Layout />,
    loader: layoutLoader,
    errorElement: <Error />,
    children: [
      {
        index: true,
        element: <Navigate to="connections" />,
      },
      {
        path: 'connections',
        element: <Connections />,
      },
      {
        path: 'connections/:plugin/:id',
        element: <ConnectionDetailPage />,
      },
      {
        path: 'projects',
        element: <ProjectHomePage />,
      },
      {
        path: 'projects/:pname',
        element: <ProjectDetailPage />,
      },
      {
        path: 'projects/:pname/:unique',
        element: <BlueprintConnectionDetailPage />,
      },
      {
        path: 'advanced',
        children: [
          {
            path: 'blueprints',
            element: <BlueprintHomePage />,
          },
          {
            path: 'blueprints/:id',
            element: <BlueprintDetailPage />,
          },
          {
            path: 'blueprints/:bid/:unique',
            element: <BlueprintConnectionDetailPage />,
          },
          {
            path: 'pipelines',
            element: <Pipelines />,
          },
          {
            path: 'pipeline/:id',
            element: <Pipeline />,
          },
        ],
      },
      {
        path: 'keys',
        element: <ApiKeys />,
      },
    ],
  },
]);
