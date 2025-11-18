package result // import "github.com/LarsArtmann/clean-wizard/internal/result"

type Result[T any] struct{ ... }
    func Err[T any](err error) Result[T]
    func Map[T, U any](r Result[T], fn func(T) U) Result[U]
    func MockSuccess[T any](value T, message string) Result[T]
    func Ok[T any](value T) Result[T]
