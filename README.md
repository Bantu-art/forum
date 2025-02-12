# Forum Web Application

A modern web forum built with Go that enables user communication through posts, comments, and reactions.

## Features

- **User Authentication**
  - Secure registration and login system
  - Email verification
  - Session management with cookies
  - Password encryption using bcrypt

- **Posts & Comments**
  - Create, read, and delete posts
  - Comment on posts
  - Associate categories with posts
  - File attachments for posts

- **Interactive Features**
  - Like/dislike posts and comments
  - Real-time updates for reactions
  - User profile management
  - Avatar/profile picture support

- **Content Organization**
  - Category-based post filtering
  - View created posts
  - View liked posts

## Technology Stack

- **Backend**
  - Go (Standard library)
  - SQLite3 for database
  - bcrypt for password encryption

- **Frontend**
  - Pure HTML/CSS/JavaScript
  - No external frameworks
  - Responsive design
  - Modern UI/UX

## Prerequisites

- Go 1.19 or higher
- Docker
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/Bantu-art/forum.git
cd forum
```

2. Build and run with Docker:
```bash
docker build -t forum .
docker run -p 8080:8080 forum
```

3. Or run locally without Docker:
```bash
go mod init forum
go run .
```

The application will be available at `http://localhost:8080`

## Database Schema

The application uses SQLite with the following main tables:
- Users
- Posts
- Comments
- Categories
- Reactions
- Sessions

## API Endpoints

### Authentication
- `POST /signup` - Register new user
- `POST /signin` - User login
- `POST /signout` - User logout

### Posts
- `GET /` - Get all posts
- `GET /post/{id}` - Get single post
- `POST /post` - Create new post
- `POST /post/delete` - Delete post

### Comments
- `POST /comment` - Add comment
- `POST /comment/delete` - Delete comment
- `POST /comment/edit` - Edit comment

### Reactions
- `POST /react` - Like/dislike post
- `POST /commentreact` - Like/dislike comment

### Filters
- `GET /category/{id}` - Filter posts by category
- `GET /created` - View created posts
- `GET /liked` - View liked posts

## Security Features

- Password encryption using bcrypt
- Session management with UUID
- CSRF protection
- Input validation and sanitization
- Secure cookie handling

## Development

### Project Structure
```
forum/
├── main.go
├──authentication/
├── controllers/
├── go.mod
├── go.sum
├── utils/
├── static/
├── templates/
├── .gitignore
├── forum.db
└── README.md
```

### Running Tests
```bash
go test ./...
```

## Docker Support

The application is containerized using Docker for consistent deployment. The Dockerfile includes:
- Multi-stage build
- Minimal base image
- Security best practices
- Environment configuration

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
