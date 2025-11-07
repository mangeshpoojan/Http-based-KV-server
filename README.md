# 🗄️ Go HTTP CRUD Server with MySQL & phpMyAdmin

This project implements a Go-based HTTP server supporting full CRUD operations (Create, Read, Update, Delete) across multiple route groups.
The server connects to a MySQL database managed via phpMyAdmin, both running locally through Docker Compose.

A Bash script (crud.sh) is included for quick command-line interaction with the server’s endpoints.

---

## 🚀 Setup Instructions

### 1. Clone the Repository
```bash
git clone https://github.com/<your-username>/<your-repo>.git
cd <your-repo>
```

### 2. Start MySQL and phpMyAdmin with Docker
This project uses Docker Compose to spin up a MySQL server and phpMyAdmin on your localhost.

Run the following command:
```bash
docker compose up -d
```

**Services started:**
- MySQL → localhost:3306
- phpMyAdmin → localhost:8080

Access phpMyAdmin in your browser at:
[http://localhost:8080](http://localhost:8080)

### 3. Start the Go HTTP Server
Run the Go server on port 6000 with cache:
```bash
go run server.go :6000 C
```

Run the Go server on port 6000 without cache:
```bash
go run server.go :6000 D
```

---

## ⚙️ Usage via Bash Script

The repository includes a helper script named `crud.sh` for easily testing the CRUD endpoints.

### Command format:
```bash
./crud.sh <port> <memory> <CRUD> <key> <value>
```

<memory>    -> this has 2 options
                - C (with cache)
                - D (without cache)

### Example:
```bash
./crud.sh 6000 C create 10 new
./crud.sh 6000 D read 10
```

---

## 🌐 HTTP Routes

This HTTP server exposes 8 routes, divided between two route groups — C and D.

| Route | Method | Description |
|--------|---------|-------------|
| /C/create | GET | Create (hits the server cache as well) |
| /C/read | GET | Read (hits the server cache as well) |
| /C/update | POST | Update (hits the server cache as well) |
| /C/delete | POST | Delete (hits the server cache as well) |
| /D/create | GET | Create (hits the DB directly) |
| /D/read | GET | Read (hits the DB directly) |
| /D/update | POST | Update (hits the DB directly) |
| /D/delete | POST | Delete (hits the DB directly) |

> ⚠️ **Notes:**
> - `create` and `read` routes only support GET requests.
> - `update` and `delete` routes only support POST requests.

---

## 🧾 JSON Request Format

For POST requests (`update`, `delete`), send data in the following JSON format:

```json
{
  "key": 10,
  "value": "apple"
}
```

---

## 📦 Example Commands

### Create (POST)
```bash
curl -X POST "http://localhost:6000/C/delete" \
  -H "Content-Type: application/json" \
  -d '{"key": 1, "value": "orange"}'
```

### Read (GET)
```bash
curl "http://localhost:6000/D/read?key=1"
```

### Update (POST)
```bash
curl -X POST "http://localhost:6000/C/update" \
  -H "Content-Type: application/json" \
  -d '{"key": 1, "value": "mango"}'
```

### Delete (GET)
```bash
curl "http://localhost:6000/D/delete?key=1"
```

---

## 🧱 Project Structure

```
├── server.go               # Go HTTP server (routes + logic)
├── crud.sh                 # Bash for CRUD operations
├── docker-compose.yml      # MySQL + phpMyAdmin setup
├── go.mod                  # Go module file
├── go.sum                  # Go dependency
├── architecture.drawio     # architecture
├── assets/
│   └── architecture.png    # png file of architecture
└── README.md               # Project documentation
```

---

## 🧰 Tech Stack

- Go (Golang) — HTTP server and JSON handling
- MySQL — Persistent storage backend
- phpMyAdmin — Database management interface
- Docker Compose — Container orchestration
- Bash — CLI automation for CRUD testing

---

## 🚧 Future Improvements

- abhi sochne do pls

---

## 👤 Author

- Developed by Mangesh Poojan
- GitHub: https://github.com/mangeshpoojan
