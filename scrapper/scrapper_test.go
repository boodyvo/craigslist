package scrapper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testProxy = "http://165.225.84.143:8800"

func TestGetLast(t *testing.T) {
	r := require.New(t)
	scrapper := New(testProxy)
	last, err := scrapper.GetLastIndex()
	r.NoError(err)
	r.NotEqual(0, last)
}
