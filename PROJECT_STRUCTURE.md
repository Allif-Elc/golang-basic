# API Documentation Platform - Project Structure

Complete project structure for the API Documentation Platform with separated backend and frontend.

## Directory Structure

```
golang-basic/
├── api/                               # Backend (Go)
│   ├── internal/                      # Internal Go packages
│   │   ├── config/                    # Configuration
│   │   │   └── database.go            # PostgreSQL connection
│   │   │
│   │   ├── model/                     # Data models
│   │   │   ├── user.go                # ✅ Existing
│   │   │   ├── profile.go             # ✅ Existing
│   │   │   ├── authorization.go       # ✅ Existing - ABAC models
│   │   │   ├── permission.go          # ✅ Existing
│   │   │   ├── pagination.go           # ✅ Existing
│   │   │   ├── project.go             # 🆕 NEW - Project model
│   │   │   ├── rest_api.go            # 🆕 NEW - REST API model
│   │   │   ├── graphql_api.go         # 🆕 NEW - GraphQL API model
│   │   │   ├── grpc_api.go            # 🆕 NEW - gRPC API model
│   │   │   └── tag.go                 # 🆕 NEW - Tag model
│   │   │
│   │   ├── repository/                # Data access layer
│   │   │   ├── user_repository.go     # ✅ Existing
│   │   │   ├── profile_repository.go  # ✅ Existing
│   │   │   ├── authorization_repository.go # ✅ Existing - ABAC
│   │   │   ├── permission_repository.go   # ✅ Existing
│   │   │   ├── project_repository.go  # 🆕 NEW
│   │   │   ├── rest_api_repository.go # 🆕 NEW
│   │   │   ├── graphql_api_repository.go # 🆕 NEW
│   │   │   ├── grpc_api_repository.go   # 🆕 NEW
│   │   │   └── tag_repository.go        # 🆕 NEW
│   │   │
│   │   ├── service/                   # Business logic
│   │   │   ├── user_service.go        # ✅ Existing
│   │   │   ├── profile_service.go     # ✅ Existing
│   │   │   ├── authorization_service.go # ✅ Existing - ABAC
│   │   │   ├── permission_service.go    # ✅ Existing
│   │   │   ├── project_service.go     # 🆕 NEW
│   │   │   ├── rest_api_service.go    # 🆕 NEW
│   │   │   ├── graphql_api_service.go # 🆕 NEW
│   │   │   ├── grpc_api_service.go    # 🆕 NEW
│   │   │   └── tag_service.go         # 🆕 NEW
│   │   │
│   │   ├── controller/                # HTTP handlers
│   │   │   ├── user_controller.go     # ✅ Existing
│   │   │   ├── profile_controller.go  # ✅ Existing
│   │   │   ├── permission_controller.go # ✅ Existing
│   │   │   ├── project_controller.go  # 🆕 NEW
│   │   │   ├── rest_api_controller.go # 🆕 NEW
│   │   │   ├── graphql_api_controller.go # 🆕 NEW
│   │   │   ├── grpc_api_controller.go   # 🆕 NEW
│   │   │   ├── tag_controller.go        # 🆕 NEW
│   │   │   └── public_controller.go     # 🆕 NEW - Public docs viewer
│   │   │
│   │   ├── middleware/                # Middleware
│   │   │   └── authorization.go       # ✅ Existing - ABAC middleware (reuse)
│   │   │
│   │   ├── routes/                    # Route definitions
│   │   │   └── routes.go              # ✅ Existing (add new routes here)
│   │   │
│   │   └── utility/                   # Helper functions
│   │       └── response.go            # ✅ Existing - Standardized responses
│   │
│   ├── sql/                           # Database schemas
│   │   ├── abac_schema.sql            # ✅ Existing - ABAC schema
│   │   └── api_docs_schema.sql        # 🆕 NEW - API docs tables
│   │
│   ├── main.go                        # ✅ Entry point
│   ├── go.mod                         # ✅ Go module definition
│   └── go.sum                         # ✅ Go dependencies lock
│
├── web/                               # Frontend (React + Vite + TypeScript)
│   ├── src/
│   │   ├── components/                # React components
│   │   │   ├── ui/                    # Reusable UI components
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   ├── Select.tsx
│   │   │   │   ├── Textarea.tsx
│   │   │   │   └── ...
│   │   │   ├── forms/                 # Form components
│   │   │   │   ├── ProjectForm.tsx
│   │   │   │   ├── RESTAPIForm.tsx
│   │   │   │   ├── GraphQLAPIForm.tsx
│   │   │   │   ├── GrpcAPIForm.tsx
│   │   │   │   └── ...
│   │   │   ├── editors/               # API-specific editors
│   │   │   │   ├── RESTEditor.tsx
│   │   │   │   ├── GraphQLEditor.tsx
│   │   │   │   ├── GrpcEditor.tsx
│   │   │   │   └── ParameterField.tsx
│   │   │   ├── viewers/               # Documentation viewers
│   │   │   │   ├── DocViewer.tsx
│   │   │   │   ├── APIDisplay.tsx
│   │   │   │   ├── CodeHighlighter.tsx
│   │   │   │   └── Sidebar.tsx
│   │   │   └── layout/                # Layout components
│   │   │       ├── Header.tsx
│   │   │       ├── Sidebar.tsx
│   │   │       └── Layout.tsx
│   │   │
│   │   ├── pages/                     # Page components
│   │   │   ├── Dashboard.tsx          # 🆕 Project list
│   │   │   ├── CreateProject.tsx      # 🆕 Create project
│   │   │   ├── ProjectDetail.tsx      # 🆕 Project details
│   │   │   ├── EditorREST.tsx         # 🆕 REST API editor
│   │   │   ├── EditorGraphQL.tsx      # 🆕 GraphQL API editor
│   │   │   ├── EditorGRPC.tsx         # 🆕 gRPC API editor
│   │   │   ├── Viewer.tsx             # 🆕 Public viewer
│   │   │   └── Login.tsx              # 🆕 Authentication
│   │   │
│   │   ├── hooks/                     # Custom React hooks
│   │   │   ├── useAuth.ts
│   │   │   ├── useProjects.ts
│   │   │   ├── useAPI.ts
│   │   │   └── useToast.ts
│   │   │
│   │   ├── services/                  # API service layer
│   │   │   ├── api.ts                 # Axios configuration
│   │   │   ├── auth.ts                # Authentication service
│   │   │   ├── projectService.ts      # Project API calls
│   │   │   ├── restApiService.ts      # REST API calls
│   │   │   ├── graphqlApiService.ts   # GraphQL API calls
│   │   │   ├── grpcApiService.ts      # gRPC API calls
│   │   │   └── tagService.ts          # Tag API calls
│   │   │
│   │   ├── stores/                    # Zustand state management
│   │   │   ├── authStore.ts
│   │   │   ├── projectStore.ts
│   │   │   └── notificationStore.ts
│   │   │
│   │   ├── utils/                     # Utility functions
│   │   │   ├── cn.ts                  # Class name utility (clsx)
│   │   │   ├── validation.ts          # Zod schemas
│   │   │   └── format.ts              # Formatting utilities
│   │   │
│   │   ├── types/                     # TypeScript type definitions
│   │   │   └── api.ts                 # 🆕 API response types
│   │   │
│   │   ├── App.tsx                    # Root component with routes
│   │   ├── main.tsx                   # Entry point
│   │   └── index.css                  # Global styles
│   │
│   ├── public/                        # Static assets
│   │   └── vite.svg
│   │
│   ├── package.json                   # Dependencies
│   ├── pnpm-lock.yaml                 # Lock file
│   ├── tsconfig.json                  # TypeScript config
│   ├── vite.config.ts                 # Vite config
│   ├── tailwind.config.js             # Tailwind config
│   ├── postcss.config.js              # PostCSS config
│   ├── .eslintrc.json                 # ESLint config
│   ├── .prettierrc.json               # Prettier config
│   ├── .env.example                   # Environment variables template
│   ├── .gitignore                     # Git ignore
│   ├── index.html                     # HTML entry point
│   └── README.md                      # Frontend documentation
│
├── .claude/                           # Claude Code
│   ├── design-document.md            # 🆕 NEW - High-level design
│   └── implementation-plan.md         # Implementation tasks
│
├── ABAC_README.md                    # ✅ Existing - ABAC documentation
├── PROJECT_STRUCTURE.md              # 🆕 NEW - This file
└── README.md                         # Project overview
```

## Module Information

### Go Module
- **Module Path**: `golang-basic/api`
- **Entry Point**: `api/main.go`
- **Working Directory**: Run commands from `/api` folder

### Frontend
- **Package Manager**: pnpm
- **Entry Point**: `web/src/main.tsx`
- **Dev Server**: Runs on `http://localhost:5173`

## Development Commands

### Backend (from `/api` folder)
```bash
# Run backend server
cd api
go run main.go

# Run tests
go test ./...

# Build
go build -o ../bin/api-docs
```

### Frontend (from `/web` folder)
```bash
# Install dependencies
cd web
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Import Path Updates

All Go imports have been updated from:
```go
import "golang-basic/internal/model"
```

To:
```go
import "golang-basic/api/internal/model"
```

## Implementation Status

### ✅ Completed

- Project structure created with `/api` and `/web` separation
- Backend models created (project, rest_api, graphql_api, grpc_api, tag)
- SQL schema created in `api/sql/`
- Frontend configuration created in `web/`
- All Go import paths updated to `golang-basic/api`
- Documentation files created

### 🚧 To Be Implemented

#### Backend (`/api`)
- [ ] Repository layer (project, rest_api, graphql_api, grpc_api, tag)
- [ ] Service layer (project, rest_api, graphql_api, grpc_api, tag)
- [ ] Controller layer (project, rest_api, graphql_api, grpc_api, tag, public)
- [ ] Routes setup (add to `api/internal/routes/routes.go`)
- [ ] MinIO integration

#### Frontend (`/web`)
- [ ] UI components (Button, Input, Modal, etc.)
- [ ] Form components (with React Hook Form + Zod)
- [ ] Editor components (REST, GraphQL, gRPC)
- [ ] Viewer components
- [ ] API services
- [ ] State management (Zustand stores)
- [ ] Authentication integration

## Running the Application

### Backend
```bash
cd api
go run main.go
# Server runs on http://localhost:8080
```

### Frontend
```bash
cd web
pnpm dev
# Dev server runs on http://localhost:5173
```

The frontend is configured to proxy API requests to the backend automatically via Vite proxy configuration.
