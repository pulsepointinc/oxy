package roundrobin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUANewChromeUnicode(t *testing.T) {
	assert.False(t, IsUANewChrome("Mozilla/5.0 (Linux; U; Android 2.3.6; es-co; XT320 Build/GRK39F) AppleWebKit/533.1 (KHTML, like Gecko) Versión/4.0 Mobile Safari/533.1"))
}

func TestIsUANewChromeEmptyString(t *testing.T) {
	assert.False(t, IsUANewChrome(""))
}

func TestIsUANewChromeImpostors(t *testing.T) {
	assert.False(t, IsUANewChrome("Chrome/NotANumber"))
	assert.False(t, IsUANewChrome("Chrome/NaN.1"))
	assert.False(t, IsUANewChrome("Chrome/∞.0"))
	assert.False(t, IsUANewChrome("Chrome/-1.0"))
}

func TestIsUANewChrome77Plus(t *testing.T) {
	assert.True(t, IsUANewChrome("Chrome/78.0"))
	assert.True(t, IsUANewChrome("Mozilla/5.0 AppleWebKit/537.36 (KHTML like Gecko) Chrome/79.0.3883.121"))
}
