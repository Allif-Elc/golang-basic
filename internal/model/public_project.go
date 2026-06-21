package model

// PaginatedAPIResponse represents a paginated response for API endpoints
type PaginatedAPIResponse[T any] struct {
	Data  []T `json:"data"`
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
	Total int  `json:"total"`
}

// PublicProjectDocumentationResponse represents the full documentation response for a public project
type PublicProjectDocumentationResponse struct {
	Project Project                              `json:"project"`
	Rest    PaginatedAPIResponse[RestAPI]    `json:"rest"`
	GraphQL PaginatedAPIResponse[GraphQLAPI] `json:"graphql"`
	Grpc    PaginatedAPIResponse[GrpcAPI]    `json:"grpc"`
}
