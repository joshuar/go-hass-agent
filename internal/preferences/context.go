// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import "context"

type contextKey string

const (
	appIDContextKey      contextKey = "appID"
	restAPIURLContextKey contextKey = "restAPIURL"
)

func AppIDToContext(ctx context.Context, appID string) context.Context {
	newCtx := context.WithValue(ctx, appIDContextKey, appID)

	return newCtx
}

func AppIDFromContext(ctx context.Context) string {
	appID, ok := ctx.Value(appIDContextKey).(string)
	if !ok {
		return defaultAppID
	}

	return appID
}

func RestAPIURLToContext(ctx context.Context, url string) context.Context {
	newCtx := context.WithValue(ctx, restAPIURLContextKey, url)

	return newCtx
}

func RestAPIURLFromContext(ctx context.Context) string {
	url, ok := ctx.Value(appIDContextKey).(string)
	if !ok {
		return ""
	}

	return url
}
