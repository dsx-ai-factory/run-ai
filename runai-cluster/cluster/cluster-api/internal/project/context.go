package project

import "context"

type contextKey int

const (
	projectContextKey contextKey = iota
	namespaceContextKey
)

func WithProject(ctx context.Context, projectName string) context.Context {
	return context.WithValue(ctx, projectContextKey, projectName)
}

func FromContext(ctx context.Context) string {
	projectName, _ := ctx.Value(projectContextKey).(string)
	return projectName
}

func WithNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, namespaceContextKey, namespace)
}

func NamespaceFromContext(ctx context.Context) string {
	namespace, _ := ctx.Value(namespaceContextKey).(string)
	return namespace
}
