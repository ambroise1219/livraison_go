# ILEX Backend - Go + SurrealDB

Backend API for ILEX delivery platform built with Go and SurrealDB, featuring complex business logic for delivery management, user authentication, and promotional systems.

## 🏗️ Architecture

This backend focuses on **complex business logic** while leveraging SurrealDB's native CRUD and real-time capabilities:

- ✅ **Complex Business Logic**: Delivery assignment, price calculation, role-based validation
- ✅ **Smart Delivery Assignment**: Automatic driver assignment based on location and vehicle compatibility
- ✅ **Dynamic Pricing**: Multi-factor price calculation (distance, vehicle type, delivery type, promos)
- ✅ **OTP Authentication**: Secure phone-based authentication with JWT tokens
- ✅ **Role-Based Access**: CLIENT, LIVREUR (Driver), ADMIN, GESTIONNAIRE, MARKETING
- ✅ **Promotional System**: Discount codes, referral system with validation
- ✅ **Real-time Ready**: WebSocket endpoints for delivery tracking

## 🚀 Quick Start

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [SurrealDB](https://surrealdb.com/) running on your server

### Installation

1. **Clone the repository**
```bash
git clone <your-repo>
cd ilex-backend
```

2. **Install dependencies**
```bash
go mod tidy
```

3. **Configure environment**
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your actual values
# IMPORTANT: Never commit .env to version control
```

4. **Set up your .env file**
```env
# SurrealDB Configuration
SURREALDB_URL=ws://your-surrealdb-host:8000/rpc
SURREALDB_USERNAME=root
SURREALDB_PASSWORD=your-password
SURREALDB_NS=ilex
SURREALDB_DB=production

# JWT Secret (use a strong 32+ character string)
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters-long

# SMS API for OTP (optional for development)
SMS_API_KEY=your-sms-api-key
SMS_API_SECRET=your-sms-secret
```

5. **Run the server**
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### Health Check
```http
GET /health
```

### Authentication
```http
POST /api/v1/auth/otp/send     # Send OTP to phone
POST /api/v1/auth/login        # Verify OTP & login
POST /api/v1/auth/refresh      # Refresh JWT token
POST /api/v1/auth/logout       # Logout (revoke refresh token)
GET  /api/v1/auth/profile      # Get user profile
```

### Deliveries
```http
POST /api/v1/deliveries                    # Create delivery (CLIENT only)
GET  /api/v1/deliveries                    # List deliveries
GET  /api/v1/deliveries/{id}               # Get delivery details
PUT  /api/v1/deliveries/{id}/status        # Update delivery status
GET  /api/v1/deliveries/calculate-price    # Calculate delivery price
POST /api/v1/deliveries/assign             # Assign delivery (ADMIN only)
```

### Admin Panel
```http
GET /api/v1/admin/dashboard        # Admin dashboard
GET /api/v1/admin/users           # User management
GET /api/v1/admin/deliveries/stats # Delivery statistics
```

## 🔐 Authentication Flow

1. **Send OTP**: `POST /auth/otp/send` with phone number
2. **Verify OTP**: `POST /auth/login` with phone + OTP code
3. **Get JWT Token**: Use token for subsequent API calls
4. **Authorization Header**: `Bearer your-jwt-token`

### Example Authentication
```bash
# 1. Send OTP
curl -X POST http://localhost:8080/api/v1/auth/otp/send \
  -H "Content-Type: application/json" \
  -d '{"phone":"+221771234567"}'

# 2. Login with OTP
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"+221771234567","code":"123456"}'

# 3. Use JWT token
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/auth/profile
```

## 🏪 Business Logic Examples

### Smart Delivery Assignment
The system automatically assigns deliveries to the best available driver based on:
- Driver location (closest first using Haversine distance)
- Vehicle compatibility with delivery type
- Driver status (ONLINE/AVAILABLE)
- Driver profile completion (documents, vehicle setup)

### Dynamic Price Calculation
Prices are calculated considering:
- **Base price** by vehicle type (MOTO: 1000 FCFA, VOITURE: 2000 FCFA, CAMIONNETTE: 5000 FCFA)
- **Distance surcharge** after included kilometers
- **Delivery type multipliers** (EXPRESS: +50%, GROUPEE: -30%, DEMENAGEMENT: +100%)
- **Waiting time charges** after free minutes
- **Promotional discounts** (percentage, fixed amount, or free delivery)

### Role-Based Permissions
- **CLIENT**: Create deliveries, view own deliveries, apply promos
- **LIVREUR**: Accept deliveries, update status, view assigned deliveries
- **ADMIN/GESTIONNAIRE**: Full access, user management, delivery assignment
- **MARKETING**: Promo management, referral system

## 🗂️ Project Structure

```
ilex-backend/
├── config/           # Environment configuration
├── db/               # SurrealDB connection & helpers
├── models/           # Data models matching SurrealDB schema
│   ├── user.go       # User, roles, driver status
│   ├── auth.go       # OTP, JWT, refresh tokens
│   ├── delivery.go   # Delivery, package, location
│   ├── vehicle.go    # Vehicle, driver location
│   └── promo.go      # Promotions, referrals, pricing
├── services/         # Business logic layer
│   ├── auth_service.go     # OTP, JWT management
│   ├── delivery_service.go # Assignment, pricing
│   └── promo_service.go    # Promotions, referrals
├── handlers/         # HTTP request handlers
├── routes/           # Route configuration & middleware
├── tests/            # Unit & integration tests
└── main.go           # Application entry point
```

## 🔧 Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run only unit tests (skip integration)
go test -short ./...
```

### Environment Variables
All sensitive configuration is handled through environment variables. Never commit actual credentials to version control.

Key variables:
- `SURREALDB_URL`: Your SurrealDB connection URL
- `JWT_SECRET`: Secure random string for JWT signing
- `SMS_API_KEY`: SMS provider credentials for OTP
- `SMTP_*`: Email configuration for notifications

## 📦 SurrealDB Schema

The backend works with a comprehensive SurrealDB schema including:
- **Users** with role management and driver profiles
- **Deliveries** with status tracking and special types (grouped, moving)
- **Vehicles** with document management
- **Promotions** and referral system
- **Payments** and wallet management
- **Real-time tracking** and notifications

## 🔒 Security Features

- **JWT Authentication** with refresh tokens
- **Role-based access control** on all endpoints
- **OTP verification** for phone-based authentication
- **Input validation** on all API endpoints
- **Secure environment** configuration
- **CORS protection** and request rate limiting

## 🚦 Production Deployment

Before deploying:
1. Set `ENVIRONMENT=production`
2. Use strong JWT secrets (32+ characters)
3. Configure SMS/Email providers
4. Set up proper logging and monitoring
5. Use HTTPS in production
6. Configure firewall rules for SurrealDB

## 🤝 Contributing

1. Follow Go best practices
2. Add tests for new features
3. Update documentation
4. Never commit sensitive data

## 📄 License

[Your License Here]