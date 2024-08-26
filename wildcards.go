package stdchi

import (
	"context"
	"strings"
)

type wildcardCtx struct{}
type wildcardValues map[string]string

func withWildcards(ctx context.Context, wcs wildcardValues) context.Context {
	return context.WithValue(ctx, wildcardCtx{}, wcs)
}

func wildcardsFromContext(ctx context.Context) wildcardValues {
	wcs, ok := ctx.Value(wildcardCtx{}).(wildcardValues)
	if !ok {
		wcs = wildcardValues{}
	}
	return wcs
}

func wildcards(s string) []string {
	var wilds []string

	for len(s) > 0 {
		idx := strings.IndexRune(s, '/')
		if idx < 0 {
			if ws := toWildcard(s); ws != "" {
				wilds = append(wilds, ws)
			}
			break
		}
		wilds = append(wilds, toWildcard(s[:idx]))
		s = s[idx+1:]
	}

	return wilds
}

func uniWildcards(s string) []string {
	wilds := wildcards(s)
	uniWilds := wilds[:0]
	for _, ws := range wilds {
		if ws == "" {
			continue
		}
		uniWilds = append(uniWilds, ws)
	}
	return uniWilds
}

func toWildcard(s string) string {
	if !(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) {
		return ""
	}
	if s == "{$}" {
		return ""
	}
	return strings.TrimSuffix(s[1:len(s)-1], "...")
}

func stripToLastSlash(s string, cnt int) string {
	pos := 0
	for i, r := range s {
		if r == '/' {
			pos = i
			cnt--
			if cnt <= 0 {
				break
			}
		}
	}
	return s[pos:]
}
