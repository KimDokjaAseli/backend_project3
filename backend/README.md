# Wallet Point Backend - Golang

Backend implementation for Platform Wallet Point Gamifikasi Kampus.

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- MySQL 8.0 or higher

### Installation

1. **Clone the repository**
```bash
cd backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Set up database**
```bash
# Run SQL files in order
mysql -u root -p < ../database/01_tables.sql
mysql -u root -p < ../database/02_triggers_procedures.sql
mysql -u root -p < ../database/03_seed_data.sql
```

5. **Run the application**
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## 📁 Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   ├── config.go                # Configuration loader
│   └── database.go              # Database connection
├── internal/
│   ├── auth/                    # Authentication module
│   ├── user/                    # User management (Admin)
│   ├── wallet/                  # Wallet & transactions (Admin)
│   └── marketplace/             # Product management (Admin)
├── middleware/
│   ├── auth.go                  # JWT authentication
│   ├── role.go                  # Role-based access control
│   ├── logger.go                # Request logging
│   └── cors.go                  # CORS handling
├── routes/
│   └── routes.go                # API route definitions
├── utils/
│   ├── jwt.go                   # JWT utilities
│   ├── password.go              # Password hashing
│   └── response.go              # Response formatting
├── .env.example                 # Environment variables template
├── go.mod                       # Go module file
└── README.md                    # This file
```

## 🔑 Environment Variables

Key environment variables (see `.env.example` for all):

```env
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=walletpoint_db
JWT_SECRET=your-secret-key-minimum-32-characters
```

## 📚 API Endpoints

### Public Endpoints
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/health` - Health check

### Admin Endpoints (Protected)

**User Management**
- `POST /api/v1/admin/users` - Create new user
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/users/:id` - Get user details
- `PUT /api/v1/admin/users/:id` - Update user
- `DELETE /api/v1/admin/users/:id` - Deactivate user
- `PUT /api/v1/admin/users/:id/password` - Change password

**Wallet Management**
- `GET /api/v1/admin/wallets` - List all wallets
- `GET /api/v1/admin/wallets/:id` - Get wallet details
- `GET /api/v1/admin/wallets/:id/transactions` - Get wallet transactions
- `POST /api/v1/admin/wallet/adjustment` - Adjust points manually
- `POST /api/v1/admin/wallet/reset` - Reset wallet balance

**Transaction Monitoring**
- `GET /api/v1/admin/transactions` - List all transactions

**Marketplace Management**
- `GET /api/v1/admin/products` - List all products
- `POST /api/v1/admin/products` - Create product
- `GET /api/v1/admin/products/:id` - Get product details
- `PUT /api/v1/admin/products/:id` - Update product
- `DELETE /api/v1/admin/products/:id` - Delete product

## 🧪 Testing

### Login Test
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@campus.edu","password":"Password123!"}'
```

### Get Users (Admin)
```bash
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ Architecture

### Handler-Service-Repository Pattern

```
Client → Handler → Service → Repository → Database
```

- **Handler**: HTTP request/response handling
- **Service**: Business logic
- **Repository**: Database operations

### Example Flow
```go
// 1. Handler receives request
func (h *UserHandler) GetAll(c *gin.Context) {
    // 2. Call service
    response, err := h.service.GetAllUsers(params)
    
    // 3. Return response
    utils.SuccessResponse(c, 200, "Success", response)
}

// Service processes business logic
func (s *UserService) GetAllUsers(params) {
    // 4. Call repository
    return s.repo.GetAllWithWallets(params)
}

// Repository queries database
func (r *UserRepository) GetAllWithWallets(params) {
    // 5. Execute query
    return r.db.Table("users").Joins("wallets").Find(&users)
}
```

## 🔐 Security

- **Password Hashing**: Bcrypt with cost factor 10
- **JWT Authentication**: Stateless token-based auth
- **Role-Based Access**: Admin, Dosen, Mahasiswa
- **SQL Injection Prevention**: Parameterized queries (GORM)
- **CORS**: Configurable allowed origins

## 📊 Default Test Accounts

After running seed data:

| Role | Email | Password | NIM/NIP |
|------|-------|----------|---------|
| Admin | admin@campus.edu | Password123! | ADM001 |
| Dosen | dosen1@campus.edu | Password123! | NIP001 |
| Mahasiswa | mahasiswa1@campus.edu | Password123! | 2023001 |

⚠️ **Change these passwords in production!**

## 🛠️ Development

### Build

```bash
go build -o wallet-point cmd/server/main.go
```

### Run

```bash
./wallet-point
```

### Install new dependency

```bash
go get github.com/package/name
go mod tidy
```

## 📝 Next Steps

### To be Implemented
- [ ] Dosen module (Mission, Task, Validation)
- [ ] Mahasiswa module (Wallet view, Transfer, Marketplace purchase)
- [ ] External integration module
- [ ] Audit logs module
- [ ] File upload handling
- [ ] Swagger documentation

## 🐛 Troubleshooting

### Database connection failed
```bash
# Check MySQL status
# Windows
net start MySQL80

# Check credentials in .env
```

### Port already in use
```bash
# Change SERVER_PORT in .env
SERVER_PORT=8081
```

### Go module errors
```bash
go mod tidy
go mod download
```

## 📄 License

MIT License - see LICENSE file

---

**Part of**: Platform Wallet Point Gamifikasi Kampus  
**Version**: 1.0.0 (Admin Features)  
**Last Updated**: 2026-01-13
