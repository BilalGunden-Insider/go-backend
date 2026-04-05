# Go Backend - Financial Transaction API

Go ile yazılmış, PostgreSQL destekli finansal işlem backend'i.

## Teknolojiler

- **Go 1.25** - Backend
- **PostgreSQL 16** - Veritabanı
- **Redis 7** - Cache katmanı
- **Docker Compose** - Orkestrasyon
- **Prometheus + Grafana** - Monitoring

## Özellikler

- JWT tabanlı authentication (register/login)
- Role-based access control (user/admin)
- Credit, debit, transfer işlemleri
- Redis cache ile balance sorgulama
- Circuit breaker (DB fallback mekanizması)
- İşlem başına ve günlük transaction limitleri
- Zamanlanmış (scheduled) transaction desteği
- Batch transaction processing (async worker pool)
- Audit logging (tüm state change'ler)
- Rate limiting (IP bazlı)
- Prometheus metrikleri + Grafana dashboard

## Başlangıç

```bash
docker-compose up --build -d
```

Servisler:
- API: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## API Endpoint'leri

### Auth (Public)
```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Users
```
GET    /api/v1/users              (admin)
GET    /api/v1/users/{id}         (authenticated)
PUT    /api/v1/users/{id}         (authenticated)
DELETE /api/v1/users/{id}         (admin)
```

### Transactions
```
POST /api/v1/transactions/credit     (admin)
POST /api/v1/transactions/debit      (admin)
POST /api/v1/transactions/transfer   (authenticated)
GET  /api/v1/transactions            (authenticated)
GET  /api/v1/transactions/{id}       (authenticated)
POST /api/v1/transactions/{id}/rollback (admin)
POST /api/v1/transactions/batch      (admin)
```

### Balances
```
GET /api/v1/balances/{user_id}       (authenticated)
GET /api/v1/balances/{user_id}/at    (authenticated)
```

### Scheduled Transactions
```
POST /api/v1/scheduled-transactions           (authenticated)
GET  /api/v1/scheduled-transactions           (authenticated)
GET  /api/v1/scheduled-transactions/{id}      (authenticated)
POST /api/v1/scheduled-transactions/{id}/cancel (authenticated)
```

### Worker Stats
```
GET /api/v1/workers/stats            (admin)
```

## Test

```bash
docker-compose up --build -d
sleep 5
./test_requests.sh
```

## Environment Variables

| Değişken | Default | Açıklama |
|----------|---------|----------|
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/backend_path?sslmode=disable | PostgreSQL bağlantı string'i |
| REDIS_URL | redis://localhost:6379 | Redis bağlantı string'i |
| PORT | 8080 | HTTP server portu |
| JWT_SECRET | change-me-in-production | JWT imzalama anahtarı |
| LOG_LEVEL | info | Log seviyesi (debug/info/warn/error) |
| ENVIRONMENT | development | Ortam (development/production) |

## Proje Yapısı

```
cmd/server/          - Giriş noktası
internal/
  api/handler/       - HTTP handler'lar
  api/middleware/     - Auth, CORS, rate limit, logging
  api/response/      - JSON response helper'ları
  auth/              - JWT token işlemleri
  cache/             - Redis cache katmanı
  circuitbreaker/    - Circuit breaker implementasyonu
  config/            - Environment config
  database/          - PostgreSQL bağlantı ve migration
  logger/            - Structured logging
  metrics/           - Prometheus metrikleri
  models/            - Veri modelleri
  repository/        - Data access layer (interface + postgres)
  service/           - İş mantığı katmanı
  worker/            - Async worker pool
migrations/          - SQL migration dosyaları
monitoring/          - Prometheus & Grafana config
```
