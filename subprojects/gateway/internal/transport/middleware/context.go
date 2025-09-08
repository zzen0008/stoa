package middleware

// contextKey is a custom type to avoid key collisions in context.
type contextKey string

// Defines the key used to store user claims in the request context.
const UserGroupsKey contextKey = "user_groups"
