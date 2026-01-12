package model

type PageRequest struct {
	Page int
	Size int
}

type PageResult[T any] struct {
	Data       []T
	Page       int
	Size       int
	StartRow   int
	EndRow     int
	NextCursor int64
}
