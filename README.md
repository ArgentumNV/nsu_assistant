# 🎓 NSU Assistant

A Telegram bot for Novosibirsk State University (NSU) students that provides class schedules, reminders, and notifications about upcoming academic events.

The project is built using **Go** and **Python**, with communication between services through a REST API.

> 🚧 **Project Status:** Early development

## ✨ Features

### 📚 Class Schedule

- View today's schedule.
- View tomorrow's schedule.
- View the weekly schedule.
- Select your academic group.
- Receive automatic daily schedule notifications.

### 🔔 Reminders

- Create personal reminders.
- Track upcoming exams, tests, and assignment deadlines.
- Receive notifications before important events.
- Manage existing reminders.

### 📢 Schedule Updates

- Automatically check for schedule changes.
- Notify students about canceled or rescheduled classes.
- Track classroom and timetable changes.

> Note: These features are planned and will be implemented incrementally.

## 🛠 Tech Stack

### Backend (Go)

| Technology    | Purpose             |
| ------------- | ------------------- |
| Go            | Backend application |
| net/http      | REST API            |
| PostgreSQL    | Data storage        |
| pgx           | PostgreSQL driver   |
| encoding/json | JSON serialization  |
| Docker        | Containerization    |

The Go backend is responsible for:

- Fetching schedules from NSU resources.
- Parsing and processing schedule data.
- Providing a REST API.
- Managing users and reminders.
- Scheduling notifications.
- Interacting with PostgreSQL.

### Telegram Bot (Python)

| Technology    | Purpose                   |
| ------------- | ------------------------- |
| Python 3      | Bot application           |
| aiogram       | Telegram Bot API          |
| aiohttp       | Async HTTP requests       |
| python-dotenv | Environment configuration |

The Python service is responsible for:

- Handling Telegram commands.
- Managing user interactions.
- Displaying schedules and reminders.
- Sending notifications to users.
- Communicating with the Go backend.

## 🏗 Architecture

The project consists of two independent services:

1. **Go Backend** — business logic, data processing and storage.
2. **Python Telegram Bot** — user interaction through Telegram.

Services communicate using HTTP requests and JSON.

```text
                  User
                   |
                   v
                Telegram
                   |
                   v
           +----------------+
           |   Python Bot   |
           |    (aiogram)   |
           +--------+-------+
                    |
                    | REST API (HTTP/JSON)
                    v
           +----------------+
           |   Go Backend   |
           |   (net/http)   |
           +--------+-------+
                    |
          +---------+---------+
          |                   |
          v                   v
    +------------+     +--------------+
    | PostgreSQL |     | NSU Schedule |
    |            |     |    Source    |
    +------------+     +--------------+
```

## 📁 Project Structure

```text
nsu-assistant/
│
├── backend/                  # Go backend service
│   ├── cmd/
│   │   └── server/
│   │       └── main.go       # Application entry point
│   │
│   ├── internal/
│   │   ├── api/              # HTTP handlers and routing
│   │   ├── schedule/         # Schedule fetching and parsing
│   │   ├── reminder/         # Reminder management
│   │   ├── storage/          # Database operations
│   │   └── models/           # Domain models
│   │
│   ├── migrations/           # Database migrations
│   ├── go.mod
│   └── Dockerfile
│
├── bot/                      # Python Telegram bot
│   ├── handlers/             # Telegram command handlers
│   ├── services/             # Backend API client
│   ├── keyboards/            # Telegram keyboards
│   ├── main.py               # Bot entry point
│   ├── requirements.txt
│   └── Dockerfile
│
├── docs/
│   └── api.md                # REST API contracts
│
├── .gitignore
├── .env.example              # Example environment variables
├── docker-compose.yml        # Multi-container configuration
├── README.md
└── LICENSE
```

> This is the planned project structure. Some directories and files will be added during development.

## 🔌 REST API

The Go backend exposes an HTTP API used by the Python Telegram bot.

Base URL (local development):

```text
http://localhost:8080
```

### Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

### Get Schedule

```http
GET /api/v1/schedule?group=24101&date=2026-10-19
```

Response:

```json
{
  "group": "24101",
  "date": "2026-10-19",
  "lessons": [
    {
      "subject": "Mathematical Analysis",
      "start_time": "09:00",
      "end_time": "10:35",
      "room": "312"
    }
  ]
}
```

For complete API documentation, see [API Contracts](docs/api.md).

## 🚀 Getting Started

The project is currently under development.

### Prerequisites

- Go
- Python 3
- Git
- PostgreSQL (required in later development stages)
- Docker (optional during early development)

### Clone Repository

```bash
git clone https://github.com/YOUR_USERNAME/nsu-assistant.git
cd nsu-assistant
```

Replace `YOUR_USERNAME` with the actual GitHub username or organization.

### Run Go Backend

Once the backend entry point is implemented:

```bash
cd backend
go run ./cmd/server
```

The backend will be available at:

```text
http://localhost:8080
```

### Run Python Bot

Once the bot is implemented:

```bash
cd bot

python -m venv .venv
source .venv/bin/activate

pip install -r requirements.txt
python main.py
```

Before running the bot, configure your Telegram bot token and backend URL.

## ⚙️ Configuration

Configuration will be managed through environment variables.

Example `.env`:

```env
# Telegram
BOT_TOKEN=your_telegram_bot_token

# Go Backend
SERVER_PORT=8080
BACKEND_URL=http://localhost:8080

# PostgreSQL
DATABASE_URL=postgres://user:password@localhost:5432/nsu_assistant?sslmode=disable
```

The `.env` file must not be committed to Git.

Use `.env.example` to document required configuration values.

## 🗺 Development Roadmap

### Phase 1 — MVP

- [ ] Initialize Go backend.
- [ ] Initialize Python Telegram bot.
- [ ] Implement `/health` endpoint.
- [ ] Implement schedule API with mock data.
- [ ] Implement `/start` Telegram command.
- [ ] Implement `/today` Telegram command.
- [ ] Connect Python bot to Go backend.

### Phase 2 — Schedule Integration

- [ ] Investigate NSU schedule data source.
- [ ] Implement schedule fetching.
- [ ] Implement schedule parsing.
- [ ] Support different academic groups.
- [ ] Add `/tomorrow` and `/week` commands.

### Phase 3 — Database & Reminders

- [ ] Integrate PostgreSQL.
- [ ] Implement user storage.
- [ ] Implement reminder CRUD operations.
- [ ] Implement notification scheduler.
- [ ] Add personal reminders.

### Phase 4 — Notifications & Deployment

- [ ] Implement automatic daily notifications.
- [ ] Detect schedule changes.
- [ ] Add integration tests.
- [ ] Configure Docker Compose.
- [ ] Set up GitHub Actions.
- [ ] Deploy the application.

## 👥 Development Team

The project is developed by two contributors.

**Go Backend Developer**

- Backend architecture
- REST API
- Schedule processing
- Database integration
- Reminder scheduling

**Python Bot Developer**

- Telegram bot architecture
- Command handlers
- User interaction
- Backend API integration
- Notification delivery

## 🤝 Contributing

Development takes place in separate Git branches.

Recommended workflow:

1. Create a new branch from `main`.
2. Implement a feature or bug fix.
3. Commit your changes.
4. Push your branch to GitHub.
5. Open a Pull Request.
6. Review and merge changes.

Example:

```bash
git checkout -b feature/schedule-api

git add .
git commit -m "feat: implement schedule API"

git push origin feature/schedule-api
```

API changes must be documented in `docs/api.md`.

## 📄 License

This project is licensed under the MIT License.

See [LICENSE](LICENSE) for details.
