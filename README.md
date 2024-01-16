# Nami - Orchestrate AWS ECS with kubectl-like Commands


Nami is a command-line tool that simplifies the orchestration of AWS ECS (Elastic Container Service) using commands similar to kubectl, making it easier for developers and DevOps teams to manage containerized applications on AWS. With Nami, you can seamlessly deploy, scale, and manage containerized workloads on ECS, leveraging familiar commands and workflows.

## Features
kubectl-like Commands: Nami adopts a syntax similar to kubectl, providing a familiar experience for users already familiar with Kubernetes.

AWS ECS Integration: Interact with AWS ECS using Nami commands, simplifying the deployment and management of containerized applications.

Scale Applications: Scale your ECS services up or down with a simple command, adjusting resources based on demand.

## In progress

AWS Proton Integration: Easily deploy and manage containerized workloads on ECS, including tasks, services, and clusters.

Configuration Management: support for configuration files and templates to deploy ECS services as Cloudformation stack.

## Installing


```
git clone https://github.com/chnacib/nami.git
cd nami
go build
mv nami /usr/bin
nami version
```

## Command Line Reference

### List available commands.

```
nami -h
```
```
nami --help
```

#### List clusters

```
nami get clusters
```

#### List services

```
nami get service -c [cluster] 
```

#### List Tasks

```
nami get task [service] -c [cluster]
```


#### List Task definition

```
nami get taskdefinition
```

#### List Task definition revisions

```
nami get revision [taskdefinition] 
```

#### List autoscaling

```
nami get autoscaling -c [cluster]
```

#### Describe cluster

```
nami describe cluster [cluster]
```

#### Describe service

```
nami describe service [service] -c [cluster]
```

#### Describe task

```
nami describe task [task] -c [cluster]
```

#### Describe task definition

```
nami describe taskdefinition [taskdefinition]
```


#### Set autoscaling configuration

```
nami set autoscale [service] --cpu 40 --mem 30 --request 500 --min 1 --max 10 -c [cluster]
```

#### Set service desired count

```
nami set replicas [service] -d 5 -c [cluster]
```

#### Update service revision

```
nami set revision [service] -r 78 -c [cluster]
```

#### Container exec interactive command

```
nami exec [task] -c [cluster] [command]
```

#### Retrieve container logs

```
nami logs task [task] -c [cluster] --limit 500
```

```
nami logs service [service] -c [cluster] --limit 500
```

## Getting Started










