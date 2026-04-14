# UNIVERSITY OF TRANSPORT AND COMMUNICATIONS
## FACULTY OF INFORMATION TECHNOLOGY

**SOCIALIST REPUBLIC OF VIETNAM**  
*Independence – Freedom – Happiness*

**Hanoi, January 2026**

# BACHELOR GRADUATION PROJECT ASSIGNMENT

## 1. Student and supervisor information

**Student full name:** Vu Nhat Minh  
**Student ID:** 222631127  
**Class:** CNTTVA1  
**Course:** 63  
**Phone number:** 0969122004  
**Email:** minhkr2302@gmail.com  
**Major:** High-quality program of English – Vietnamese Information Technology  
**Program:** Full-time

**Supervisor:** Dr. Dao Thi Le Thuy  
**Affiliation:** Faculty of Information Technology, University of Transport and Communications  
**Phone number:** 0946921976  
**Email:** thuydtl@utc.edu.vn

## 2. Project title

**DEVELOPING A SUBSCRIPTION-BASED MOVIE STREAMING PLATFORM WITH AI-POWERED RECOMMENDATION SYSTEM**

## 3. Content and scope of the thesis

### 3.1 Research content

- **User Management & Authentication**  
  The system provides secure user registration, login, and role-based access control using JWT authentication. Users are categorized into different roles such as administrators and subscribers, enabling appropriate authorization and content access.

- **Subscription & Payment Management**  
  The platform supports paid subscription models (monthly/annual plans), allowing users to access premium movie content after successful payment. The system records subscription history, payment status, and expiration time to control viewing permissions.

- **Movie Streaming & Content Delivery**  
  The system allows users to browse, search, and stream movies online. Movie metadata (title, genre, description, duration, rating) is managed by administrators. Video streaming is optimized to ensure smooth playback with minimal latency.

- **AI-Based Movie Recommendation System**  
  An AI-powered recommendation module is developed to suggest movies based on user viewing history, preferred genres, and interaction patterns. The system applies machine learning techniques such as content-based filtering and similarity measurement to personalize user experience.

- **Backend API & Business Logic**  
  A RESTful API is developed using Golang to handle business logic such as user authentication, movie catalog management, subscription validation, and viewing history tracking. PostgreSQL is used to store structured data securely and efficiently.

- **Frontend Web Application**  
  A modern single-page application (SPA) is built using Vue.js to deliver a user-friendly interface similar to Netflix. The UI supports responsive design, dynamic movie listings, user profiles, and subscription status visualization.

- **System Integration & Security**  
  The project focuses on integrating frontend and backend through secure APIs, implementing input validation, access control, and data protection to ensure system reliability and security.

### 3.2 Scope of the topic

- The project focuses on building a paid movie streaming platform integrated with basic AI capabilities.
- AI is applied primarily for personalized recommendation and intelligent search, rather than complex video content analysis.
- The system is designed for academic purposes and small-scale deployment, serving as a foundation for future commercial expansion.

## 4. Technologies, tools, and programming languages

### 4.1 Programming languages

- **Backend:** Golang
- **AI Service:** Python 3.10+
- **Frontend:** Vue 3 (JavaScript / TypeScript)

### 4.2 AI & data processing

- **Libraries:** Scikit-learn, NumPy, Pandas
- **Techniques:** Content-based filtering, cosine similarity

### 4.3 Backend technologies

- **Framework:** Gin / Fiber
- **Database:** PostgreSQL
- **Authentication:** JWT

### 4.4 Frontend technologies

- Vue 3
- Pinia
- Vue Router
- Tailwind CSS

### 4.5 Development tools

- Visual Studio Code
- Docker & Dock

## 5. Expected main outcomes

### 5.1 Technical performance

- Personalized movie recommendations with measurable relevance improvement
- Secure and scalable backend architecture
- Responsive frontend with Netflix-like UI/UX
- Stable integration between Golang backend and Python AI microservice

### 5.2 Deliverables

- A complete subscription-based movie streaming website
- An AI-powered recommendation system
- RESTful backend APIs and AI microservice APIs
- Source code repository with deployment instructions
- Technical report covering system architecture, AI model design, and experimental results

## 6. Project implementation plan

| No. | Task description | Implementation period | Notes |
|---|---|---|---|
| 1 | Requirement Analysis & System Design | 01/03/2026 – 07/03/2026 | Define functional requirements, system architecture, database schema. |
| 2 | Backend Core Development (User & Auth Service) | 08/03/2026 – 21/03/2026 | JWT Authentication, Role-based Access Control. |
| 3 | Movie & Subscription Management Service | 22/03/2026 – 28/03/2026 | KPI: CRUD latency < 200ms per request |
| 4 | Payment & Subscription Validation Integration | 29/03/2026 – 04/04/2026 | KPI: Payment success handling accuracy > 99%. |
| 5 | AI Recommendation Model Design | 05/04/2026 – 11/04/2026 | Content-based filtering using cosine similarity |
| 6 | AI Recommendation Service Implementation | 12/04/2026 – 18/04/2026 | KPI: Recommendation relevance improvement ≥ 30% compared to random. |
| 7 | Frontend Development (Netflix-like UI) | 19/04/2026 – 02/05/2026 | Responsive UI, personalized recommendation display |
| 8 | Backend–AI–Frontend Integration | 03/05/2026 – 09/05/2026 | End-to-end API integration, KPI: Response time < 500ms |
| 9 | System Testing & Optimization | 10/05/2026 – 20/05/2026 | Load testing, bug fixing, performance tuning |
| 10 | Validation & Final Deployment | 21/05/2026 – 31/05/2026 | Deliverable: Final system & technical report. |

## 7. Signatures

**Dean**  
(Signature and printed name)  

**TS. Hoàng Văn Thông**

**Head of Department**  
(Signature and printed name)  

**TS. Cao Thị Luyên**

**Supervisor**  
(Signature and printed name)  

**TS. Đào Lệ Thu Thủy**

**Student**  
(Signature and printed name)  

**Vũ Nhật Minh**
