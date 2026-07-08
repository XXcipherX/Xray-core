package splithttp_test

import (
	"testing"

	. "github.com/xtls/xray-core/transport/internet/splithttp"
)

func Test_GetNormalizedPath(t *testing.T) {
	tests := []struct {
		name string
		c    Config
		want string
	}{
		{
			name: "root with query",
			c: Config{
				Path: "/?world",
			},
			want: "/",
		},
		{
			name: "legacy default keeps trailing slash",
			c: Config{
				Path: "/stream",
			},
			want: "/stream/",
		},
		{
			name: "auto keeps trailing slash when metadata is in path",
			c: Config{
				Path:               "/stream",
				PathTrailingSlash:  PathTrailingSlashAuto,
				SessionIDPlacement: PlacementCookie,
				SeqPlacement:       PlacementPath,
			},
			want: "/stream/",
		},
		{
			name: "auto preserves slashless path when metadata is elsewhere",
			c: Config{
				Path:               "/assets/app.js",
				PathTrailingSlash:  PathTrailingSlashAuto,
				SessionIDPlacement: PlacementCookie,
				SeqPlacement:       PlacementCookie,
			},
			want: "/assets/app.js",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.c.GetNormalizedPath()
			if got != test.want {
				t.Fatalf("unexpected path: got %q, want %q", got, test.want)
			}
		})
	}
}
