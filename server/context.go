package server

import (
	"context"
)

type contextKey string

const routeVarsKey contextKey = "routeVars"

func withRouteVars(ctx context.Context, vars map[string]string) context.Context {
	return context.WithValue(ctx, routeVarsKey, vars)
}

func GetRouteVars(ctx context.Context) map[string]string {
	val := ctx.Value(routeVarsKey)
	if v, ok := val.(map[string]string); ok {
		return v
	}
	return map[string]string{}
}

