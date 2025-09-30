# Example Configurations

This document provides example mvx configurations for common project types.

## Table of Contents

- [Go Projects](#go-projects)
- [Maven/Java Projects](#mavenjava-projects)
- [Node.js Projects](#nodejs-projects)
- [Multi-Language Projects](#multi-language-projects)
- [Monorepo Projects](#monorepo-projects)

## Go Projects

### Basic Go Project

```json5
{
  project: {
    name: "my-go-app",
    description: "A Go application"
  },

  tools: {
    go: {
      version: "1.24.2"
    }
  },

  commands: {
    build: {
      description: "Build the application",
      script: "go build -o bin/app ."
    },

    test: {
      description: "Run all tests",
      script: "go test -v ./..."
    },

    "test-coverage": {
      description: "Run tests with coverage",
      script: "go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out"
    },

    lint: {
      description: "Run linter",
      script: "golangci-lint run"
    },

    fmt: {
      description: "Format code",
      script: "go fmt ./..."
    },

    clean: {
      description: "Clean build artifacts",
      script: "rm -rf bin/ coverage.out"
    }
  }
}
```

### Go Project with Multiple Binaries

```json5
{
  tools: {
    go: { version: "1.24.2" }
  },

  commands: {
    build: {
      description: "Build all binaries",
      script: [
        "go build -o bin/server ./cmd/server",
        "go build -o bin/client ./cmd/client",
        "go build -o bin/worker ./cmd/worker"
      ]
    },

    "build-server": {
      description: "Build server binary",
      script: "go build -o bin/server ./cmd/server"
    },

    "build-client": {
      description: "Build client binary",
      script: "go build -o bin/client ./cmd/client"
    },

    test: {
      description: "Run all tests",
      script: "go test -v ./..."
    }
  }
}
```

## Maven/Java Projects

### Basic Maven Project

```json5
{
  project: {
    name: "my-java-app",
    description: "A Java application"
  },

  tools: {
    java: {
      version: "17",
      distribution: "temurin"
    },
    maven: {
      version: "4.0.0-rc-4"
    }
  },

  commands: {
    build: {
      description: "Build the project",
      script: "mvn clean install"
    },

    "build-fast": {
      description: "Build without tests",
      script: "mvn clean install -DskipTests"
    },

    test: {
      description: "Run tests",
      script: "mvn test"
    },

    "test-integration": {
      description: "Run integration tests",
      script: "mvn verify -Pintegration-tests"
    },

    run: {
      description: "Run the application",
      script: "mvn exec:java"
    },

    clean: {
      description: "Clean build artifacts",
      script: "mvn clean"
    },

    package: {
      description: "Package the application",
      script: "mvn package"
    }
  }
}
```

### Spring Boot Project

```json5
{
  tools: {
    java: { version: "21", distribution: "temurin" },
    maven: { version: "4.0.0-rc-4" }
  },

  commands: {
    build: {
      description: "Build Spring Boot application",
      script: "mvn clean package -DskipTests"
    },

    run: {
      description: "Run Spring Boot application",
      script: "mvn spring-boot:run"
    },

    "run-dev": {
      description: "Run with dev profile",
      script: "mvn spring-boot:run -Dspring-boot.run.profiles=dev"
    },

    test: {
      description: "Run tests",
      script: "mvn test"
    },

    docker: {
      description: "Build Docker image",
      script: "mvn spring-boot:build-image"
    }
  }
}
```

## Node.js Projects

### Basic Node.js Project

```json5
{
  project: {
    name: "my-node-app",
    description: "A Node.js application"
  },

  tools: {
    node: {
      version: "22.20.0"
    }
  },

  commands: {
    install: {
      description: "Install dependencies",
      script: "npm install"
    },

    build: {
      description: "Build the project",
      script: "npm run build"
    },

    test: {
      description: "Run tests",
      script: "npm test"
    },

    "test-watch": {
      description: "Run tests in watch mode",
      script: "npm run test:watch"
    },

    lint: {
      description: "Lint code",
      script: "npm run lint"
    },

    dev: {
      description: "Start development server",
      script: "npm run dev"
    },

    start: {
      description: "Start production server",
      script: "npm start"
    },

    clean: {
      description: "Clean build artifacts",
      script: "rm -rf dist/ node_modules/"
    }
  }
}
```

### TypeScript Project

```json5
{
  tools: {
    node: { version: "22.20.0" }
  },

  commands: {
    build: {
      description: "Compile TypeScript",
      script: "npm run build"
    },

    "build-watch": {
      description: "Compile TypeScript in watch mode",
      script: "npm run build:watch"
    },

    test: {
      description: "Run tests",
      script: "npm test"
    },

    typecheck: {
      description: "Type check without emitting",
      script: "tsc --noEmit"
    },

    lint: {
      description: "Lint and format",
      script: "npm run lint && npm run format"
    }
  }
}
```

## Multi-Language Projects

### Full-Stack Application (Java Backend + Node Frontend)

```json5
{
  project: {
    name: "fullstack-app",
    description: "Full-stack application with Java backend and React frontend"
  },

  tools: {
    java: { version: "17", distribution: "temurin" },
    maven: { version: "4.0.0-rc-4" },
    node: { version: "22.20.0" }
  },

  commands: {
    build: {
      description: "Build both backend and frontend",
      script: [
        "cd backend && mvn clean package -DskipTests",
        "cd frontend && npm install && npm run build"
      ]
    },

    "build-backend": {
      description: "Build backend only",
      script: "cd backend && mvn clean package"
    },

    "build-frontend": {
      description: "Build frontend only",
      script: "cd frontend && npm install && npm run build"
    },

    test: {
      description: "Run all tests",
      script: [
        "cd backend && mvn test",
        "cd frontend && npm test"
      ]
    },

    "dev-backend": {
      description: "Start backend dev server",
      script: "cd backend && mvn spring-boot:run"
    },

    "dev-frontend": {
      description: "Start frontend dev server",
      script: "cd frontend && npm run dev"
    },

    clean: {
      description: "Clean all build artifacts",
      script: [
        "cd backend && mvn clean",
        "cd frontend && rm -rf dist/ node_modules/"
      ]
    }
  }
}
```

## Monorepo Projects

### Go Monorepo with Multiple Services

```json5
{
  project: {
    name: "microservices",
    description: "Microservices monorepo"
  },

  tools: {
    go: { version: "1.24.2" }
  },

  commands: {
    build: {
      description: "Build all services",
      script: [
        "cd services/auth && go build -o ../../bin/auth .",
        "cd services/api && go build -o ../../bin/api .",
        "cd services/worker && go build -o ../../bin/worker ."
      ]
    },

    test: {
      description: "Run all tests",
      script: "go test -v ./..."
    },

    "test-auth": {
      description: "Test auth service",
      script: "cd services/auth && go test -v ./..."
    },

    "test-api": {
      description: "Test API service",
      script: "cd services/api && go test -v ./..."
    },

    lint: {
      description: "Lint all services",
      script: "golangci-lint run ./..."
    },

    clean: {
      description: "Clean all build artifacts",
      script: "rm -rf bin/"
    }
  }
}
```

## Tips

### Using Command Arrays

Commands can use arrays for multiple steps:

```json5
commands: {
  build: {
    description: "Build with multiple steps",
    script: [
      "echo 'Step 1: Clean'",
      "rm -rf dist/",
      "echo 'Step 2: Build'",
      "go build -o dist/app ."
    ]
  }
}
```

### Platform-Specific Commands

Use platform-specific scripts:

```json5
commands: {
  build: {
    description: "Build for current platform",
    script: {
      windows: "go build -o bin\\app.exe .",
      unix: "go build -o bin/app ."
    }
  }
}
```

### Using mvx-shell Interpreter

For advanced shell features:

```json5
commands: {
  deploy: {
    description: "Deploy application",
    interpreter: "mvx-shell",
    script: [
      "export VERSION=$(git describe --tags)",
      "echo Deploying version $VERSION",
      "docker build -t myapp:$VERSION .",
      "docker push myapp:$VERSION"
    ]
  }
}
```

### Command Dependencies

Use the `requires` field to specify tool dependencies:

```json5
commands: {
  build: {
    description: "Build the project",
    requires: ["go", "node"],
    script: [
      "cd frontend && npm run build",
      "go build -o bin/app ."
    ]
  }
}
```

