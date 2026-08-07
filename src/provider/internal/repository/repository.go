// Package repository provides data access abstractions.
//
// Repository Layer: abstracts storage backends (filesystem, GitHub, Feishu).
// Currently a placeholder for future implementation.
package repository

// SourceAdapter defines the interface for multi-source asset discovery.
type SourceAdapter interface {
	// Discover scans the source and returns discovered assets.
	// TODO: implement for filesystem, GitHub, Feishu adapters.
	Discover() ([]byte, error)
}
