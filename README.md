# Vessel Monitoring System (VMS)

A comprehensive monitoring system for aquaculture operations, providing real-time data tracking and analysis for fish farming activities.

## Overview

The Vessel Monitoring System (VMS) integrates Golang-based APIs, Python algorithms, and a MariaDB database to monitor and manage aquaculture data including user information, fish farming metrics, weather conditions, and feed data. The system connects to external services via OpenClaw CLI and provides a Vue.js frontend for data visualization.

## Features

- **Real-time Data Monitoring**: Track fish farming parameters, weather conditions, and feed usage
- **API Integration**: RESTful APIs built with Golang for seamless data access
- **Data Analytics**: Python-powered algorithms for data search, sorting, and analysis
- **Database Management**: MariaDB for reliable data storage and retrieval
- **Web Interface**: Vue.js frontend for intuitive data visualization and management
- **Containerized Deployment**: Docker-based deployment for easy scaling and portability

## Technology Stack

### Backend
- **Golang**: API development (v1.26.2)
- **MariaDB**: Primary database for data persistence
- **OpenClaw CLI**: External service integration

### Frontend
- **Vue.js**: User interface framework
- **Node.js**: JavaScript runtime for development

### Data Processing
- **Python**: Algorithm implementation for data processing and analysis

### Infrastructure
- **Docker**: Containerization platform
- **Docker Compose**: Multi-container orchestration
- **Nginx**: Web server and reverse proxy
- **Ansible**: Configuration management and deployment automation

## Prerequisites

- Go 1.26.2 or later
- Python 3.x
- Node.js 16.x or later
- Docker and Docker Compose
- MariaDB (if running locally)

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd project_f
   ```

2. **Install backend dependencies**
   ```bash
   cd backend
   go mod download
   ```

3. **Install frontend dependencies**
   ```bash
   cd ../frontend
   npm install
   ```

4. **Install Python dependencies**
   ```bash
   cd ../python-ai
   pip install -r requirements.txt
   ```

## Configuration

1. **Database Setup**
   - Configure MariaDB connection in `backend/config/database.go`
   - Run database migrations (if applicable)

2. **Environment Variables**
   Create a `.env` file in the project root:
   ```
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=your_user
   DB_PASSWORD=your_password
   DB_NAME=vms_db
   API_PORT=8080
   ```

3. **Docker Configuration**
   Update `docker-compose.yml` with your environment settings

## Running the Application

### Development Mode

1. **Start Database**
   ```bash
   docker-compose up mariadb -d
   ```

2. **Run Backend**
   ```bash
   cd backend
   go run main.go
   ```

3. **Run Frontend**
   ```bash
   cd frontend
   npm run dev
   ```

4. **Run Python Services** (if applicable)
   ```bash
   cd python-ai
   python main.py
   ```

### Production Deployment

```bash
# Build and start all services
docker-compose up --build -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

## API Documentation

The API endpoints are documented in the backend code. Key endpoints include:

- `GET /api/users` - Retrieve user information
- `GET /api/fish-data` - Get fish farming data
- `POST /api/weather` - Submit weather data
- `GET /api/analytics` - Access processed analytics data

## Testing

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## Project Structure

```
project_f/
├── backend/          # Golang API server
├── frontend/         # Vue.js web application
├── python-ai/        # Python algorithms and data processing
├── docs/             # Documentation
├── docker-compose.yml # Docker orchestration
├── nginx/            # Nginx configuration
└── scripts/          # Deployment and utility scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

For common issues and solutions, refer to [docs/troubleshooting.md](docs/troubleshooting.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue in the GitHub repository
- Check the documentation in the `docs/` directory
- Review the troubleshooting guide for common problems


