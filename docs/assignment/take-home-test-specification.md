# Take Home Test Fullstack Engineer

## Related Documents

- [README.md](../../README.md) — how Savio satisfies this specification.
- [Business Requirements](../product/business-requirements.md) — domain chosen for this take-home.
- [Implementation Plan](../engineering/implementation-plan.md) — mapping of requirements to milestones.
- [Testing Strategy](../engineering/testing-strategy.md) — how testing requirements are covered.

## PT. Mumtaz Teknologi Indonesia

## Overview Project

Membuat aplikasi **Fullstack Web Application** yang mencakup:

* Backend REST API
* Frontend Dashboard
* Database
* Authentication & Authorization

Domain/jenis aplikasi bebas dipilih kandidat, misalnya:

* Business management
* Finance
* Education
* Healthcare
* Productivity
* Marketplace
* Operations
* HR
* Project management
* Inventory
* Customer management
* Ide lain

Aplikasi **tidak boleh sekadar CRUD sederhana**. Harus ada:

* Problem yang jelas
* Target user yang jelas
* Business flow
* Business rules
* Data relationship
* UI/UX yang baik

---

## 1. Backend — Wajib

* **Language:** Go
* **Framework:** Gin
* **ORM:** GORM
* **Database:** PostgreSQL
* Arsitektur bebas:

    * Layered
    * Clean
    * Hexagonal
    * Modular
    * atau pendekatan lain
* Yang penting arsitektur jelas dan maintainable.

---

## 2. Frontend — Wajib

* **Language:** JavaScript / TypeScript
* **Framework:** React atau Vue
* **CSS:** Tailwind CSS
* **HTTP Client:** Axios
* State management, routing, form library, component architecture, dan UI component library bebas ditentukan.

---

## 3. Authentication & Security

* Wajib menggunakan **cookie-based authentication**.
* Tidak boleh menggunakan `localStorage` atau `sessionStorage` untuk authentication.
* Minimal mencakup:

    * Login
    * Logout
    * Authentication state
    * Protected API
    * Protected frontend route
    * Session/token expiration
    * Current-user endpoint
* Cookie wajib memperhatikan security attributes:

    * `HttpOnly`
    * `Secure`
    * `SameSite`
    * `Path`
    * `Max-Age` / `Expires`
* **CSRF protection wajib**.
* Kandidat harus mampu menjelaskan alasan pemilihan mekanisme CSRF.

---

## 4. Authorization

* Wajib ada **RBAC / Permission System**.
* Minimal memiliki lebih dari satu level akses.
* Contoh:

    * Administrator
    * Manager
    * Staff
    * User
* Role dapat disesuaikan dengan domain aplikasi.
* Authorization **wajib diterapkan di backend**.
* Frontend hanya berfungsi sebagai UX layer untuk hide/show elemen.

---

## 5. Database & Migration

* Gunakan **PostgreSQL + GORM**.
* Migration wajib mendukung:

    * Create schema
    * Alter schema
    * Rollback
    * Dijalankan dari kondisi database kosong secara reproducible
* Contoh command:

    * `make migrate-up`
    * `make migrate-down`
* Perhatikan desain database:

    * Primary key
    * Foreign key
    * Index
    * Unique constraint
    * Nullable field
    * Relationship
    * Data integrity

---

## 6. Business Logic

* Wajib memiliki **business process / workflow** yang jelas.
* Contoh:

    * `Draft → Submitted → Approved → Completed`
    * `Pending → Processing → Completed`
* Business rules wajib diterapkan di backend, bukan hanya mengandalkan frontend.

---

## 7. Input Validation

Semua input wajib divalidasi, termasuk:

* Request body
* Query parameter
* Path parameter
* Form
* File upload
* Authentication input

Frontend memberikan feedback cepat kepada user, tetapi backend tetap menjadi **source of truth**.

Minimal mencakup validasi:

* Required field
* String length
* Format
* Enum
* Numeric range
* Date
* UUID
* Email
* File type
* File size

---

## 8. REST API

* Gunakan API versioning, misalnya:

```text
/api/v1/...
```

* Gunakan HTTP status code secara tepat:

    * `200`
    * `201`
    * `204`
    * `400`
    * `401`
    * `403`
    * `404`
    * `409`
    * `422`
    * `429`
    * `500`

* Response API harus konsisten.

Contoh success response:

```json
{
  "success": true,
  "data": {},
  "message": "Success"
}
```

Error response juga wajib memiliki format yang konsisten.

---

## 9. Axios Interceptor

Wajib implementasikan interceptor minimal untuk:

* `401` — expired authentication
* `403` — permission error
* `422` — validation error
* `429` — rate limit feedback
* `500` — generic server error

Jika menggunakan refresh mechanism, implementasi harus menghindari:

* Infinite retry
* Duplicate refresh request
* Race condition saat multiple concurrent request
* Refresh failure yang tidak berujung logout

---

## 10. UI/UX

Tidak disediakan Figma atau design system.

Kandidat menentukan sendiri:

* Information architecture
* Navigation
* Layout
* Typography
* Color system
* Component system
* User flow
* Responsive behaviour

Wajib tersedia:

* Loading state

    * Skeleton / spinner
* Disabled state
* Empty state
* Error state
* Form UX:

    * Inline validation
    * Error message
    * Disabled/loading submit
    * Success feedback
* Confirmation dialog untuk destructive action
* Toast / notification
* Responsive:

    * Desktop
    * Tablet
    * Mobile

---

## 11. Reusable Components

Minimal siapkan reusable component sesuai kebutuhan:

* Button
* Input
* Select
* Modal
* Dialog
* Drawer
* Table
* Pagination
* Dropdown
* Badge
* Card
* Toast
* Form
* Loading
* EmptyState

Hindari duplikasi component dengan fungsi yang sama.

---

## 12. Search, Filter & Pagination

Minimal beberapa halaman listing wajib memiliki:

* Search
* Filter
* Sorting
* Pagination

Contoh query:

```text
?page=1&limit=20&search=keyword&sort=created_at&order=desc&status=active
```

---

# 13–16. Fitur Bonus

## File Storage

Wajib menggunakan cloud/object storage seperti:

* MinIO
* AWS S3
* GCS
* Azure Storage
* Firebase Storage

Bukan local filesystem.

Perhatikan:

* File validation
* File size
* MIME type
* Access control
* Filename strategy

## Queue Worker

Gunakan:

* Redis
* RabbitMQ

Untuk asynchronous/background processing, misalnya:

* File processing
* Email
* Notification
* Report generation
* Image processing
* Data import
* Scheduled job

## Docker

Direkomendasikan aplikasi dapat dijalankan dengan:

```bash
docker compose up
```

Mencakup sesuai kebutuhan:

* Frontend
* Backend
* PostgreSQL
* Redis
* MinIO
* Worker

## Testing

Backend:

* Minimal unit test
* Lebih baik ditambah integration test

Frontend:

* Minimal component test
* Lebih baik ditambah integration test

Advanced:

* E2E testing untuk critical flow

Contoh:

```text
login
→ create data
→ business process
→ verify result
```

---

## 17. Error Handling

Implementasikan **centralized error handling**.

Backend harus mampu membedakan:

* Validation error
* Authentication error
* Authorization error
* Not found
* Conflict
* Business error
* Internal server error

Jangan expose kepada user:

* Database error
* Stack trace
* Secret
* Internal implementation detail

---

## 18. API Documentation

Dokumentasikan seluruh API menggunakan salah satu:

* Swagger / OpenAPI
* Postman
* Insomnia

Minimal dokumentasi mencakup:

* Endpoint
* Method
* Authentication
* Request
* Query parameter
* Response
* Error response
* Example

Dokumentasi harus dapat digunakan developer lain untuk mencoba API.

---

## 19. GitHub & README

Repository wajib memiliki:

* `README.md`
* `.env.example`
* `Dockerfile` / `docker-compose.yml`
* Migration
* API documentation
* Source code

Jangan commit:

* `.env`
* Credential

README wajib menjelaskan:

* Application overview
* Problem yang diselesaikan
* Target user
* Feature
* Tech stack
* Architecture
* Database design
* Installation
* Environment configuration
* Migration
* Running application
* Testing
* API documentation
* Technical decision
* Trade-off
* Future improvement

---

## 20. Creativity Challenge

Bebas menambahkan fitur value-add, misalnya:

* Dashboard analytics
* Notification
* Approval workflow
* Audit trail
* Import/export
* Reporting
* Realtime update
* Scheduling
* Automation
* Activity timeline
* Third-party integration
* AI feature
* Recommendation
* Fitur lain yang relevan

Yang dinilai **bukan jumlah fitur**, tetapi:

* Alasan pemilihan fitur
* Kualitas implementasi

---

## 21. Bonus Engineering

### Performance

* Redis caching
* Database indexing
* Query optimization
* Rate limiting
* Optimistic locking

### Security

* CSRF protection
* Refresh token rotation
* Session management
* Brute-force protection
* Security headers

### Observability

* Structured logging
* Request ID
* Health check
* Metrics
* OpenTelemetry

### DevOps

* CI/CD
* Automated migration
* Automated testing pipeline
* Production Docker setup

---

## 22. Deliverables

GitHub Repository berisi:

* Backend
* Frontend
* Database migration
* API documentation
* README
* `.env.example`

Bonus:

* File storage
* Queue worker
* Docker
* Automated testing
* E2E testing
* CI/CD
* Caching
* Observability

---

## 23. Bobot Penilaian

Total: **100%**

| Area                                | Bobot |
| ----------------------------------- | ----: |
| Backend Architecture & Code Quality |   15% |
| API & Business Logic                |   15% |
| Authentication & Security           |   15% |
| UI/UX                               |   15% |
| Database & Migration                |   10% |
| Frontend Architecture               |   10% |
| Testing                             |    5% |
| Documentation                       |    5% |
| Git & Engineering Practice          |    5% |

Bonus dapat digunakan sebagai nilai pembeda, bukan sebagai kewajiban.

---

## 24. Interview Discussion

Setelah project dikumpulkan, kandidat akan diminta menjelaskan keputusan teknis yang dibuat.

Contoh topik pembahasan:

* Kenapa memilih domain tersebut
* Siapa target user
* Apa business problem yang diselesaikan
* Kenapa memilih architecture tersebut
* Kenapa authentication menggunakan cookie
* Fungsi:

    * `HttpOnly`
    * `Secure`
    * `SameSite`
* Cara mengatasi CSRF
* Cara refresh authentication
* Cara authorization diterapkan
* Kenapa menggunakan Composite API
* Cara menangani concurrent request
* Desain database jika data menjadi sangat besar
* Kenapa menggunakan queue
* Cara menangani failed job
* Cara memastikan file upload aman
* Cara meningkatkan performance aplikasi
* Trade-off dari architecture yang dipilih

Kandidat diharapkan mampu menjelaskan **alasan di balik setiap keputusan teknis**, bukan hanya menunjukkan bahwa fitur tersebut berjalan.

