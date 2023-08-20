# Todo API

This is a simple RESTful API that allows you to manage your personal todos. You can create, read, update, and delete todos using HTTP requests.

## Getting Started

To use this API, you need to have the following:

- A web server that can run Go code
- A Postgres database that stores your todos
- A tool to send HTTP requests, such as Postman or curl

## Installation

To install this API, follow these steps:

- Clone this repository to your local machine
- Navigate to the project folder and run `go mod download` to download the dependencies
- Run `go run main.go` to start the web server
- You will need to set various environment variables if you want to run it with your own postgres database
  - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, SSL_MODE
- The API will be available at `http://localhost:8080`

## Usage

The API exposes the following endpoints:

- `GET /todos`: Get all todos
- `POST /todos`: Create a new todo
- `GET /todos/:id`: Get a single todo by ID
- `PUT /todos/:id`: Update a todo by ID
- `DELETE /todos/:id`: Delete a todo by ID
- `DELETE /todos`: Delete multiple todos by ids

Each todo has the following fields:

- `id`: A unique identifier for the todo
- `title`: The message of the todo
- `completed`: A boolean value that indicates whether the todo is completed or not
- `last_updated`: The date and time when the todo was last updated

### Examples

Here are some examples of how to use the API with curl:

#### Get all todos

```bash
curl -X GET http://localhost:8080/todos
```

#### Get a single todo by ID

```bash
curl -X GET http://localhost:8080/todos/1
```

#### Update a todo by ID

```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Check for food", "completed": true}'
```

#### Delete a todo by ID

```bash
curl -X DELETE http://localhost:8080/todos/1
```

#### Delete multiple todos by ids

```bash
curl -X DELETE http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '[1, 2]'
```
