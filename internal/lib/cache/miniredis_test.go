package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func newMiniredis(t *testing.T) string {
	t.Helper()
	s := miniredis.RunT(t)
	return s.Addr()
}
