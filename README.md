# Todo API

A RESTful API for managing tasks with authentication, filtering, sorting, and pagination capabilities.

> **Note**: This is a simplified proof-of-concept implementation focusing on core functionality with minimal complexity.

## Features

- **Authentication**: JWT-based authentication with mock users
- **Task Management**: Full CRUD operations for tasks
- **Filtering**: Filter tasks by status and search terms
- **Sorting**: Sort tasks by various fields (created_at, updated_at, title, status)
- **Pagination**: Paginated task listing with metadata
- **Real API Responses**: Proper HTTP status codes and error handling

## Mock Users

The API comes with pre-configured mock users for testing:

| Email | Password |
|-------|----------|
| john.doe@example.com | password123 |
| jane.smith@example.com | password123 |
| mike.wilson@example.com | password123 |

## Simplified Data Models

This implementation uses simplified data models for proof-of-concept purposes:

### User Model
```json
{
  "id": "uuid",
  "email": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Task Model
```json
{
  "id": "uuid",
  "title": "string",
  "status": "pending|in_progress|completed|cancelled",
  "user_id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## API Endpoints

### Authentication

#### POST /api/v1/auth/login
Login with email and password to get access token.

**Request Body:**
```json
{
  "email": "john.doe@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "error": false,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Tasks

All task endpoints require authentication. Include the access token in the Authorization header:
```
Authorization: Bearer <access_token>
```

#### GET /api/v1/tasks
Get list of tasks with filtering, sorting, and pagination.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)
- `status` (optional): Filter by status (pending, in_progress, completed, cancelled)
- `search` (optional): Search in title
- `sort_field` (optional): Sort field (created_at, updated_at, title, status)
- `sort_order` (optional): Sort order (asc, desc)

**Example:**
```
GET /api/v1/tasks?page=1&limit=5&status=in_progress&sort_field=created_at&sort_order=desc
```

**Response:**
```json
{
  "error": false,
  "message": "Tasks retrieved successfully",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "title": "Complete project documentation",
      "status": "in_progress",
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T14:20:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 5,
      "total": 1,
      "total_pages": 1
    },
    "sort": "created_at:desc",
    "filter": "status:in_progress"
  }
}
```

#### POST /api/v1/tasks
Create a new task.

**Request Body:**
```json
{
  "title": "Review code changes"
}
```

**Response:**
```json
{
  "error": false,
  "message": "Task created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "title": "Review code changes",
    "status": "pending",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T15:30:00Z",
    "updated_at": "2024-01-15T15:30:00Z"
  }
}
```

#### GET /api/v1/tasks/:id
Get a specific task by ID.

**Response:**
```json
{
  "error": false,
  "message": "Task retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Complete project documentation",
    "status": "in_progress",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T14:20:00Z"
  }
}
```

#### PUT /api/v1/tasks/:id
Update a specific task.

**Request Body:**
```json
{
  "title": "Updated task title",
  "status": "completed"
}
```

**Response:**
```json
{
  "error": false,
  "message": "Task updated successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Updated task title",
    "status": "completed",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T16:45:00Z"
  }
}
```

#### DELETE /api/v1/tasks/:id
Delete a specific task.

**Response:**
```json
{
  "error": false,
  "message": "Task deleted successfully"
}
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": true,
  "message": "Error description"
}
```

Common HTTP status codes:
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Running the API

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the server:**
   ```bash
   go run cmd/main.go
   ```

3. **The API will be available at:**
   ```
   http://localhost:3000
   ```

## Environment Variables

You can customize the API behavior using environment variables:

- `SERVER_PORT`: Server port (default: 3000)
- `SERVER_HOST`: Server host (default: 0.0.0.0)
- `JWT_SECRET_KEY`: JWT secret key (default: todo-api-secret-key-change-in-production)
- `JWT_ACCESS_TOKEN_TTL`: Access token TTL (default: 15m)
- `JWT_REFRESH_TOKEN_TTL`: Refresh token TTL (default: 168h)
- `APP_ENV`: Application environment (default: development)

## Project Structure

```
todo-api/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── domain/
│   │   ├── auth/              # Authentication domain models
│   │   └── task/              # Task domain models
│   ├── handler/
│   │   ├── auth/              # Authentication handlers
│   │   └── task/              # Task handlers
│   ├── middleware/
│   │   └── auth_middleware.go # Authentication middleware
│   └── service/
│       ├── auth/              # Authentication service
│       └── task/              # Task service
└── pkg/
    ├── config/                # Configuration management
    ├── types/                 # Common types
    └── utils/                 # Utility functions
```

## Testing the API

You can test the API using curl or any HTTP client. Here are some example requests:

1. **Login:**
   ```bash
   curl -X POST http://localhost:3000/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"john.doe@example.com","password":"password123"}'
   ```

2. **Get tasks:**
   ```bash
   curl -X GET http://localhost:3000/api/v1/tasks \
     -H "Authorization: Bearer <access_token>"
   ```

3. **Create task:**
   ```bash
   curl -X POST http://localhost:3000/api/v1/tasks \
     -H "Authorization: Bearer <access_token>" \
     -H "Content-Type: application/json" \
     -d '{"title":"New task"}'
   ```
