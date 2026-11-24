# Threat Alert API

A simple REST API built using Go to manage threat alerts reported by mobile applications.

Each company has its own users and applications. Admins can manage users and applications, while admins and viewers can access alerts belonging to their company. Mobile SDKs can report threats using an application key.

The API uses Go's standard library and stores all data in memory, without requiring an external database.

## How to run

**Requirements:**
- Go 1.22 or higher.

**Steps:**

1. Clone the repository:

   ```bash
   git clone https://github.com/Mitarthsharma/threat-alert-api.git
   ```

2. Navigate to the project directory:

   ```bash
   cd threat-alert-api
   ```

3. Start the server:

   ```bash
   go run .
   ```

4. The server will be available at:

   ```text
   http://localhost:8080
   ```

   To use a different port, set the `PORT` environment variable before starting the server.

**Run automated tests:**

```bash
go test ./...
```

Since the project uses in-memory storage, all data is lost when the server restarts.

## Project structure

| File | Description |
|---|---|
| `main.go` | Application entry point and HTTP route registration. |
| `models.go` | Defines the Company, User, App and Alert structs. |
| `store.go` | Handles in-memory data storage and synchronization using a mutex. |
| `helpers.go` | Contains JSON response, validation and random string helper functions. |
| `auth.go` | Handles user authentication using Bearer tokens. |
| `handlers.go` | Contains handlers for registration, user and app management, and alert operations. |
| `handlers_test.go` | Contains automated tests for authorization, company isolation and concurrent alert reporting. |
| `go.mod` | Defines the Go module and Go version. |
| `README.md` | Project documentation and setup instructions. |

## What works and what does not

**Implemented:**

- Company registration with an initial admin account.
- Creating additional admin and viewer users.
- Creating Android and iOS applications with randomly generated app keys.
- Reporting threat alerts using application keys.
- Updating existing open alerts when the same threat is reported again.
- Listing alerts with optional status filtering.
- Retrieving individual alerts.
- Resolving alerts with admin-only access.
- Token-based authentication and role-based authorization.
- Company-level data isolation.
- Thread-safe alert reporting for concurrent requests.
- Automated tests using Go's `net/http/httptest` package.

**Limitations:**

- Data is stored in memory and is lost when the server stops.
- Database integration and persistent storage are not implemented.
- I could not run the race detector locally because my Windows environment has a 32-bit GCC compiler. The regular automated tests pass using `go test ./...`. The race detector can be run in an environment with a compatible 64-bit C compiler.

## How AI helped

- **Tools I used:** ChatGPT for implementation guidance, debugging and understanding Go concepts; VS Code for development; Postman for API testing.

- **Two prompts I used:**
  1. "Can you help me create a company registration API in Go? It should create the company and its first admin, generate random IDs and tokens, validate the input, and use only the standard library."
  2. "I've completed the backend API and implemented all the required routes. Can you please review my entire project and check whether all the routes are implemented correctly according to the assignment requirements?"

- **One thing the AI got wrong, and how I found and fixed it:** During implementation, an unused variable caused a compilation error (`app declared and not used`). I checked the compiler output and corrected the declaration.

- **One part I wrote or changed myself:** I wrote the Go struct definitions in `models.go` myself, using the suggested structure as a reference. I implemented the fields for companies, users, applications and alerts according to the project requirements.

## Questions

**1. Two identical alerts arrive at the same moment. What makes sure only one alert is created?**

The alert reporting handler uses `sync.Mutex` to protect the check-and-update operation.

It locks the shared store before checking whether an open alert already exists for the same app, device ID and threat type.

If an alert exists, its count is increased. Otherwise, a new alert is created. This ensures concurrent requests are processed safely without creating duplicate alerts.

**2. Why does a user from another company get 404 and not 403?**

A 403 response indicates that a resource exists but the user is not allowed to access it. This could reveal information about another company's data.

A 404 response hides whether the resource exists at all. Therefore, the API returns 404 when a user tries to access an alert belonging to another company.