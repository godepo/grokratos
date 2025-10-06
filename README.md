# grokratos

[![codecov](https://codecov.io/gh/godepo/grokratos/graph/badge.svg?token=nDHx5tOFz6)](https://codecov.io/gh/godepo/grokratos)
[![Go Report Card](https://goreportcard.com/badge/godepo/grokratos)](https://goreportcard.com/report/godepo/grokratos)
[![License](https://img.shields.io/badge/License-MIT%202.0-blue.svg)](https://github.com/godepo/grokratos/blob/main/LICENSE)

A Go library for integration testing with [Ory Kratos](https://www.ory.sh/kratos/) using testcontainers.

## Overview

`grokratos` provides a simple and flexible way to spin up Ory Kratos containers for integration testing in Go applications. It integrates seamlessly with the `groat` testing framework and uses testcontainers-go under the hood.

## Features

- 🚀 Easy setup of Ory Kratos containers for testing
- 🔧 Configurable container images and settings
- 🏷️ Dependency injection support with custom labels
- 🔄 Automatic container lifecycle management
- 📝 Custom identity schema support
- ⚙️ Custom Kratos configuration support

## Installation
```bash 
go get github.com/godepo/grokratos
```



