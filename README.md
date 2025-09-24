# Task Management API

---
Task Management API is a GraphQL-based backend service built with Golang using gqlgen. It follows a structured architecture comprising repository, usecase, graph, and model layers. PostgreSQL is used as the database.

## 🚀 Features

- GraphQL API powered by gqlgen
- Repository-Usecase-Graph architecture
- PostgreSQL for persistent storage
- Generated models and resolvers

## 🛠 Pre-requisites

Make sure you have the following installed:

- Go 1.23.5
- PostgreSQL (latest version recommended)
- Gqlgen (GraphQL generator for Golang)

```aiexclude
/go-task-management
│── /_scripts        # Contains all SQL migrations scripts
│── /cmd             # Contains main.go (entry point of the application)
│── /internal        # Contains core business logic
│   │── /graph       # Contains GraphQL schema and generated resolvers
│   │── /model       # Contains generated models and database models
│   │── /repository  # Handles database interactions
│   │── /usecase     # Business logic layer
│   │── /db          # Database migrations and setup
│── /config          # Configuration files for reading environment variables
│── go.mod
│── go.sum
│── gqlgen.yml       # Gqlgen configuration
```

## 🛠 Installation & Setup

#### 1. Clone the repository
```aiexclude
$ git clone https://bitbucket.org/edts/go-task-management.git
$ cd go-task-management
```
#### 2. Install dependencies
```aiexclude
$ go mod tidy
```
#### 3. Set up environment variables
Update `config.yml` by replacing database url (will be given personally)
#### 4. Generate GraphQL code
```aiexclude
$ go run github.com/99designs/gqlgen generate
```
or using the given `Makefile`
```aiexclude
$ make generate
```
#### 5. Run the server
```aiexclude
$ go run cmd/main.go
```
or using the given `Makefile`
```aiexclude
$ make run_server
```

## 📜 Usage
You can test your GraphQL queries and mutations using the playground by accessing `http://localhost:8080`

Example query to fetch tasks:
```aiexclude
query {
  tasks {
    id
    title
    description
    status
  }
}
```
## 🤝 Contributing

Feel free to submit issues and pull requests to improve this API.