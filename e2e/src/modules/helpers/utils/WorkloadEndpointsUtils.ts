/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { V1Deployment } from '@kubernetes/client-node';
import { logger } from '../logger/logger';
import { kubeCli } from '../../cli';
import fs from 'fs';
import path from 'path';

export function createDeployment(
    name: string,
    namespace: string,
    project: string,
    app: string,
    replicas = 2
): V1Deployment {
    const deployment: V1Deployment = {
        apiVersion: 'apps/v1',
        kind: 'Deployment',
        metadata: {
            name,
            namespace,
            labels: {
                app: 'my-app',
                project
            }
        },
        spec: {
            replicas,
            selector: {
                matchLabels: {
                    app: app
                }
            },
            template: {
                metadata: {
                    labels: {
                        app: app,
                        project
                    }
                },
                spec: {
                    containers: [
                        {
                            name: 'ubuntu',
                            image: 'ubuntu',
                            ports: [
                                {
                                    containerPort: 8000
                                }
                            ],
                            command: ['sleep'],
                            args: ['infinity'],
                            imagePullPolicy: 'Always',
                            resources: {
                                limits: {
                                    'nvidia.com/gpu': '0'
                                }
                            }
                        }
                    ],
                    schedulerName: 'runai-scheduler'
                }
            }
        }
    };
    logger.debug(`Deployment structure: ${JSON.stringify(deployment)}`);
    return deployment;
}

export function createNodePortEndpointYaml(app: string, namespace: string): string {
    return `
    apiVersion: v1
    kind: Service
    metadata:
      name: ${app}-svc
      namespace: ${namespace}
    spec:
      type: NodePort
      selector:
        app: ${app} 
      ports:
        - protocol: TCP
          port: 8888    
          targetPort: 8888  
          nodePort: 32057 
    `;
}

export function createLoadBalancerEndpointYaml(app: string, namespace: string): string {
    return `
    apiVersion: v1
    kind: Service
    metadata:
      name: ${app}-lb-service
      namespace: ${namespace}
    spec:
      selector:
        app: ${app}
      ports:
        - protocol: TCP
          port: 443
          targetPort: 443
      type: LoadBalancer
    `;
}

export function createClusterIPServiceYaml(app: string, namespace: string, serviceName?: string): string {
    const svcName = serviceName || `${app}-clusterip-svc`;
    return `
apiVersion: v1
kind: Service
metadata:
  name: ${svcName}
  namespace: ${namespace}
spec:
  type: ClusterIP
  selector:
    app: ${app}
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8000
`;
}

export function createIngressYaml(
    namespace: string,
    serviceName: string,
    host: string,
    ingressName?: string
): string {
    const name = ingressName || `${serviceName}-ingress`;
    return `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${name}
  namespace: ${namespace}
spec:
  rules:
    - host: ${host}
      http:
        paths:
          - path: /predict
            pathType: Prefix
            backend:
              service:
                name: ${serviceName}
                port:
                  number: 8000
`;
}

export function createGatewayYaml(
    namespace: string,
    gatewayName: string,
    listenerPort: number,
    protocol: string,
    listenerName?: string
): string {
    const lName = listenerName || 'default';
    return `
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ${gatewayName}
  namespace: ${namespace}
spec:
  gatewayClassName: kgateway
  listeners:
    - name: ${lName}
      port: ${listenerPort}
      protocol: ${protocol}
      allowedRoutes:
        namespaces:
          from: Same
`;
}

export function createHTTPRouteYaml(
    namespace: string,
    routeName: string,
    gatewayName: string,
    backendName: string,
    hostname: string,
    backendPort?: number,
    path?: string,
    backendKind?: string,
    backendGroup?: string
): string {
    const port = backendPort || 8000;
    const routePath = path || '/';
    const kind = backendKind ? `          kind: ${backendKind}` : '';
    const groupLine = backendGroup ? `          group: ${backendGroup}` : '';
    return `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ${routeName}
  namespace: ${namespace}
spec:
  parentRefs:
    - name: ${gatewayName}
  hostnames:
    - ${hostname}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: ${routePath}
      backendRefs:
        - name: ${backendName}
${kind}
${groupLine}
          port: ${port}
`;
}

export function createTLSRouteYaml(
    namespace: string,
    routeName: string,
    gatewayName: string,
    serviceName: string,
    hostname: string,
    servicePort?: number
): string {
    const port = servicePort || 8000;
    return `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSRoute
metadata:
  name: ${routeName}
  namespace: ${namespace}
spec:
  parentRefs:
    - name: ${gatewayName}
  hostnames:
    - ${hostname}
  rules:
    - backendRefs:
        - name: ${serviceName}
          port: ${port}
`;
}

export function createGRPCRouteYaml(
    namespace: string,
    routeName: string,
    gatewayName: string,
    serviceName: string,
    hostname: string,
    servicePort?: number
): string {
    const port = servicePort || 8000;
    return `
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: ${routeName}
  namespace: ${namespace}
spec:
  parentRefs:
    - name: ${gatewayName}
  hostnames:
    - ${hostname}
  rules:
    - backendRefs:
        - name: ${serviceName}
          port: ${port}
`;
}

export function createInferencePoolYaml(
    namespace: string,
    poolName: string,
    app: string,
    targetPort?: number
): string {
    const port = targetPort || 8000;
    return `
apiVersion: inference.networking.k8s.io/v1
kind: InferencePool
metadata:
  name: ${poolName}
  namespace: ${namespace}
spec:
  targetPorts:
    - number: ${port}
  selector:
    matchLabels:
      app: ${app}
  endpointPickerRef:
    group: inference.networking.k8s.io
    kind: EndpointPicker
    name: default
`;
}

export function createOpenshiftRouteYaml(
    namespace: string,
    routeName: string,
    serviceName: string,
    host: string,
    path: string
): string {
    return `
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: ${routeName}
  namespace: ${namespace}
spec:
  host: ${host}
  path: ${path}
  to:
    kind: Service
    name: ${serviceName}
  port:
    targetPort: 8000
  tls:
    termination: edge
`;
}

export function createKnativeServiceYaml(
    namespace: string,
    serviceName: string,
    app: string,
    jobUrl: string
): string {
    return `
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: ${serviceName}
  namespace: ${namespace}
  labels:
    app: ${app}
  annotations:
    JobUrl: ${jobUrl}
spec:
  template:
    metadata:
      labels:
        app: ${app}
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
              nvidia.com/gpu: "0"
            limits:
              cpu: "500m"
              memory: "256Mi"
              nvidia.com/gpu: "0"
`;
}

export function createDgdYaml(namespace: string, name: string, app: string, ingressHost: string): string {
    return `
apiVersion: nvidia.com/v1alpha1
kind: DynamoGraphDeployment
metadata:
  name: ${name}
  namespace: ${namespace}
  labels:
    app: ${app}
spec:
  services:
    Frontend:
      componentType: frontend
      dynamoNamespace: ${name}
      replicas: 1
      ingress:
        enabled: true
        host: api
        hostPrefix: "${app}-"
        hostSuffix: "${ingressHost}"
        ingressControllerClassName: nginx
        tls:
          secretName: dgd-tls-secret
        annotations:
          nginx.ingress.kubernetes.io/ssl-redirect: "true"
      extraPodSpec:
        mainContainer:
          image: gcr.io/knative-samples/helloworld-go
          resources:
            limits:
              cpu: "500m"
              memory: "256Mi"
              nvidia.com/gpu: "0"
            requests:
              cpu: "100m"
              memory: "128Mi"
              nvidia.com/gpu: "0"
    Worker:
      componentType: worker
      dynamoNamespace: ${name}
      replicas: 1
      extraPodSpec:
        mainContainer:
          image: gcr.io/knative-samples/helloworld-go
          command: ["/bin/sh", "-c", "sleep infinity"]
          resources:
            limits:
              cpu: "500m"
              memory: "256Mi"
              nvidia.com/gpu: "0"
            requests:
              cpu: "100m"
              memory: "128Mi"
              nvidia.com/gpu: "0"
`;
}

export function createNimServiceYaml(
    namespace: string,
    name: string,
    app: string,
    ingressHost: string
): string {
    return `
apiVersion: apps.nvidia.com/v1alpha1
kind: NIMService
metadata:
  name: ${name}
  namespace: ${namespace}
  labels:
    app: ${app}
spec:
  image:
    repository: gcr.io/knative-samples/helloworld-go
    tag: latest
  replicas: 1
  authSecret: nvidia-ngc-secret
  resources:
    limits:
      cpu: "500m"
      memory: "256Mi"
      nvidia.com/gpu: "0"
    requests:
      cpu: "100m"
      memory: "128Mi"
  expose:
    service:
      type: ClusterIP
      port: 8000
    ingress:
      enabled: true
      spec:
        ingressClassName: nginx
        tls:
          - hosts:
              - ${name}.${ingressHost}
            secretName: nim-tls-secret
        rules:
          - host: ${name}.${ingressHost}
            http:
              paths:
                - path: /
                  pathType: Prefix
                  backend:
                    service:
                      name: ${name}
                      port:
                        number: 8000
`;
}

export function createAiAppHelmChart(releaseName: string, ingressHost: string, projectName: string): string {
    const chartDir = path.join('/tmp', `ai-app-chart-${releaseName}`);
    const templatesDir = path.join(chartDir, 'templates');
    fs.mkdirSync(templatesDir, { recursive: true });

    fs.writeFileSync(
        path.join(chartDir, 'Chart.yaml'),
        `apiVersion: v2
name: ${releaseName}
version: 0.1.0
type: application
`
    );

    fs.writeFileSync(
        path.join(chartDir, 'values.yaml'),
        `image: gcr.io/knative-samples/helloworld-go
`
    );

    // Deployment 1: nodeport workload
    fs.writeFileSync(
        path.join(templatesDir, 'deployment-nodeport.yaml'),
        `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-nodeport
  namespace: {{ .Release.Namespace }}
  labels:
    app: {{ .Release.Name }}-nodeport
    app.kubernetes.io/instance: {{ .Release.Name }}
    project: ${projectName}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-nodeport
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}-nodeport
        project: ${projectName}
    spec:
      containers:
        - name: app
          image: {{ .Values.image }}
          ports:
            - containerPort: 8080
          resources:
            limits:
              nvidia.com/gpu: "0"
`
    );

    // NodePort service for deployment 1
    fs.writeFileSync(
        path.join(templatesDir, 'service-nodeport.yaml'),
        `apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-nodeport-svc
  namespace: {{ .Release.Namespace }}
spec:
  type: NodePort
  selector:
    app: {{ .Release.Name }}-nodeport
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
`
    );

    // Deployment 2: ingress workload
    fs.writeFileSync(
        path.join(templatesDir, 'deployment-ingress.yaml'),
        `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-ingress
  namespace: {{ .Release.Namespace }}
  labels:
    app: {{ .Release.Name }}-ingress
    app.kubernetes.io/instance: {{ .Release.Name }}
    project: ${projectName}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Release.Name }}-ingress
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}-ingress
        project: ${projectName}
    spec:
      containers:
        - name: app
          image: {{ .Values.image }}
          ports:
            - containerPort: 8080
          resources:
            limits:
              nvidia.com/gpu: "0"
`
    );

    // ClusterIP service for deployment 2
    fs.writeFileSync(
        path.join(templatesDir, 'service-clusterip.yaml'),
        `apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-ingress-svc
  namespace: {{ .Release.Namespace }}
spec:
  type: ClusterIP
  selector:
    app: {{ .Release.Name }}-ingress
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
`
    );

    // Ingress for deployment 2
    fs.writeFileSync(
        path.join(templatesDir, 'ingress.yaml'),
        `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Release.Name }}-ingress
  namespace: {{ .Release.Namespace }}
spec:
  ingressClassName: nginx
  rules:
    - host: {{ .Release.Name }}.${ingressHost}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ .Release.Name }}-ingress-svc
                port:
                  number: 8080
`
    );

    logger.info(`Created AI application helm chart at ${chartDir}`);
    return chartDir;
}

export async function waitForLoadBalancerIngress(
    namespace: string,
    serviceName: string,
    timeoutMs = 15000,
    intervalMs = 2000
): Promise<string | undefined> {
    const start = Date.now();
    while (Date.now() - start < timeoutMs) {
        const svc = await kubeCli.getServiceByName(namespace, serviceName);
        const ingress = svc?.status?.loadBalancer?.ingress;
        if (ingress?.length > 0) {
            return ingress[0].ip || ingress[0].hostname;
        }
        await new Promise((resolve) => setTimeout(resolve, intervalMs));
    }
    return undefined;
}

export async function getLoadBalancerIngress(
    namespace: string,
    serviceName: string,
    fallbackIp = '10.0.0.100'
): Promise<string> {
    // Wait for a LB controller to assign an IP, or patch manually if none exists
    const ip = await waitForLoadBalancerIngress(namespace, serviceName);
    if (ip) {
        return ip;
    }
    logger.info(
        `No LB controller assigned IP for ${serviceName}, patching status with fallback IP ${fallbackIp}`
    );
    await kubeCli.patchServiceStatus(namespace, serviceName, [
        { op: 'add', path: '/status/loadBalancer', value: { ingress: [{ ip: fallbackIp }] } }
    ]);
    return fallbackIp;
}
