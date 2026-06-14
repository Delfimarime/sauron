package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeBanner(t *testing.T) {
	in := "  sauron v0.0.0-SNAPSHOT  \n\n   Hash abc \nHome: /tmp/h\n"
	assert.Equal(t, "sauron v0.0.0-SNAPSHOT\nHash abc\nHome: /tmp/h", normalizeBanner(in))
}

func TestParseBannerVersion(t *testing.T) {
	v, err := parseBannerVersion("sauron v0.0.0-SNAPSHOT\nHash abc\nHome: /tmp/h\n")
	require.NoError(t, err)
	assert.Equal(t, "0.0.0-SNAPSHOT", v)
}

func TestParseBannerVersionErrors(t *testing.T) {
	_, err := parseBannerVersion("garbage-without-the-marker\n")
	assert.Error(t, err)
	_, err = parseBannerVersion("")
	assert.Error(t, err)
}
