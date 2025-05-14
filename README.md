# nami

> Simple CLI to interact with AWS ECS in a fast and scriptable way.

![Go Version](https://img.shields.io/badge/go-1.22+-blue)
![Build](https://github.com/chnacib/nami/actions/workflows/ci.yml/badge.svg)

---

## ✨ Overview

`nami` is a lightweight command-line tool that simplifies interaction with Amazon ECS. It allows developers and platform teams to quickly inspect ECS services and tasks, fetch logs, execute commands in running containers, and much more.

If you're tired of navigating through the AWS Console or struggling with the AWS CLI verbosity — `nami` is for you.

---

## ⚡ Quickstart

### Prerequisites

- Go 1.22 or newer
- AWS CLI configured (`aws configure`)
- Valid IAM permissions for ECS and CloudWatch logs

### Installation

```bash
go install github.com/chnacib/nami@latest
```

---

# 📘 Command Line Reference

This document provides examples of common `nami` commands.

---

## 🔎 Help

List available commands:

```bash
nami -h
nami --help
```

---

## 📋 List Resources

### List ECS Clusters

```bash
nami get clusters
```

### List Services in a Cluster

```bash
nami get service -c [cluster]
```

### List Tasks in a Service

```bash
nami get task [service] -c [cluster]
```

### List Task Definitions

```bash
nami get taskdefinition
```

### List Task Definition Revisions

```bash
nami get revision [taskdefinition]
```

### List Auto Scaling Configurations

```bash
nami get autoscaling -c [cluster]
```

---

## 📝 Describe Resources

### Describe a Cluster

```bash
nami describe cluster [cluster]
```

### Describe a Service

```bash
nami describe service [service] -c [cluster]
```

### Describe a Task

```bash
nami describe task [task] -c [cluster]
```

### Describe a Task Definition

```bash
nami describe taskdefinition [taskdefinition]
```

---

## ⚙️ Set Configurations

### Set Auto Scaling for a Service

```bash
nami set autoscale [service] --cpu 40 --mem 30 --request 500 --min 1 --max 10 -c [cluster]
```

### Set Desired Task Count

```bash
nami set replicas [service] -d 5 -c [cluster]
```

### Update Service Revision

```bash
nami set revision [service] -r 78 -c [cluster]
```

---

## 🧪 Execute and Monitor

### Execute Interactive Command in a Container

```bash
nami exec [task] -c [cluster] [command]
```

### Retrieve Container Logs

```bash
nami logs task [task] -c [cluster] --limit 500

nami logs service [service] -c [cluster] --limit 500
```

---

## 🧪 Development

Clone the repo:

```bash
git clone https://github.com/chnacib/nami.git
cd nami
make build
```

Run tests:

```bash
go test ./...
```

Format and lint:

```bash
golangci-lint run
```

---

## 🤝 Contributing

Pull requests are welcome! Feel free to open issues to report bugs, request features or propose ideas.

To contribute:

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push and open a pull request

---

## 📄 License

[MIT](LICENSE)

---

## 📣 Author

Made with ☕ by [@chnacib](https://github.com/chnacib)

