# Expense Tracking Application

Application for tracking expenses

## System Overview

Expense Tracking Application is designed for using the microservice architecture.

The system consists of two components:

1. UserAPI: This component handles all user related activities, i.e., login and registration.

2. ExpenseAPI: This component handles all expense related activities, i.e., getting, adding, updating and removing expenses.

- The communication between these APIs is empowered by RabbitMQ, a message broker. RabbitMQ queues are used to facilitate communication between APIs, ensuring decoupling of services and improving the system's scalability and maintainability.

- As the persistent storage, MongoDB is used to store "User" and "Expense" collections.

## Setup and Execution

1. Set variables in a copy of ".env.example" file and rename the new file as ".env":
```sh
MONGO_URI        = "<...>"
DATABASE_NAME    = "<...>"
RABBITMQ_URI     = "<...>"
JWT_SECRET       = "<...>"
TOKEN_EXPIRATION = "<...>h<...>m<...>s"
```

2. Run with "docker compose":
```sh
docker compose up --build
```

## API Endpoints

### User Endpoints

#### `POST /user/login`
- **Description**: Log in an existing user.
- **Request Body**: 
  - `email` (string) – The user's email.
  - `password` (string) – The user's password.
- **Response**: 
  - Returns a token or session data if login is successful.

#### `POST /user/register`
- **Description**: Register a new user.
- **Request Body**: 
  - `email` (string) – The user's email.
  - `password` (string) – The user's password.
  - `name` (string) – The user's full name.
- **Response**: 
  - Returns confirmation of successful registration.

### Expense Endpoints

#### `GET /expense?expenseId={expenseId}`
- **Description**: Retrieve an expense by its `expenseId`.
- **Query Parameters**: 
  - `expenseId` (string) – The unique ID of the expense.
- **Response**: 
  - Returns a list of single expense the specified id.

#### `GET /expense?category={category}`
- **Description**: Retrieve expenses filtered by the specified category.
- **Query Parameters**: 
  - `category` (string) – The category to filter expenses by.
- **Response**: 
  - Returns a list of expenses within the specified category.

#### `GET /expense`
- **Description**: Retrieve all expenses owned by the user.
- **Request Parameters**: 
  - None.
- **Response**: 
  - Returns a list of all expenses.

#### `POST /expense`
- **Description**: Add a new expense.
- **Request Body**: 
  - `description` (string) – A description of the expense.
  - `amount` (number) – The amount of the expense.
  - `category` (string) – The category of the expense.
- **Response**: 
  - Returns the confirmation of the successful operation.

#### `PUT /expense`
- **Description**: Update an existing expense.
- **Request Body**: 
  - `expenseId` (string) – The ID of the expense to be updated.
  - `description` (string) – The updated description.
  - `amount` (number) – The updated amount.
  - `category` (string) – The updated category.
- **Response**: 
  - Returns the confirmation of the successful operation.

#### `DELETE /expense`
- **Description**: Remove an existing expense.
- **Request Body**: 
  - `expenseId` (string) – The ID of the expense to be deleted.
- **Response**: 
  - Returns the confirmation of the successful operation.

## Dependencies

- github.com/dgrijalva/jwt-go v3.2.0+incompatible
- github.com/google/uuid v1.6.0
- github.com/gorilla/mux v1.8.1
- github.com/streadway/amqp v1.1.0
- go.mongodb.org/mongo-driver v1.17.1
- golang.org/x/crypto v0.29.0
- github.com/golang/snappy v0.0.4 // indirect
- github.com/klauspost/compress v1.13.6 // indirect
- github.com/montanaflynn/stats v0.7.1 // indirect
- github.com/xdg-go/pbkdf2 v1.0.0 // indirect
- github.com/xdg-go/scram v1.1.2 // indirect
- github.com/xdg-go/stringprep v1.0.4 // indirect
- github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
- golang.org/x/crypto v0.26.0 // indirect
- golang.org/x/sync v0.9.0 // indirect
- golang.org/x/text v0.17.0 // indirect
