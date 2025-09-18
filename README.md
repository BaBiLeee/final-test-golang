# Blog API

The Blog API is a service written in Golang that utilizes:

- **PostgreSQL** for storing blog data.
- **Redis** for caching.
- **Elasticsearch** for full-text search.
- **Docker Compose** for easy system setup.

## Project Structure
```
.
├── cmd/api/main.go          # API entry point
├── internal/db/db.go        # PostgreSQL logic
├── internal/redis/cache.go  # Redis caching logic
├── internal/es/es.go        # Elasticsearch service
├── internal/handlers/posts.go # HTTP handlers
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Running the Project

### 1. Prerequisites
- **Docker**
- **Docker Compose**

### 2. Start All Services
Run the following command to build and start the services:
```bash
docker-compose up -d --build
```

Services will be available at:
- **API**: [http://localhost:8080](http://localhost:8080)
- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`
- **Elasticsearch**: [http://localhost:9200](http://localhost:9200)

### 3. Create Tables in PostgreSQL
Once the PostgreSQL container is running, connect to it:
```bash
docker exec -it blog-postgres psql -U new_user -d golang
```

Run the following SQL commands:
```sql
CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  tags TEXT[] DEFAULT ARRAY[]::TEXT[],
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS activity_logs (
  id SERIAL PRIMARY KEY,
  action VARCHAR(100) NOT NULL,
  post_id INTEGER,
  logged_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_posts_tags_gin ON posts USING GIN (tags);
```

## 📡 API Endpoints

| Method | Endpoint              | Description               |
|--------|-----------------------|---------------------------|
| POST   | `/posts`              | Create a new post         |
| GET    | `/posts/{id}`         | Get post details          |
| PUT    | `/posts/{id}`         | Update a post             |
| GET    | `/posts/search`       | Search posts in Elasticsearch |
| GET    | `/posts/search-by-tag`| Search posts by tag in PostgreSQL |

### Example API Calls

#### Create a New Post
```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello2",
    "content": "World2",
    "tags": ["a", "b"]
  }'
```

#### Get Post by ID
```bash
curl http://localhost:8080/posts/1
```

## 🛠️ Development Tips

### View API Logs
```bash
docker logs blog-api -f
```

### Connect to PostgreSQL via DBeaver
- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `golang`
- **User**: `new_user`
- **Password**: `abcd1234`
