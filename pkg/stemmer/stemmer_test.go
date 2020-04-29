package stemmer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStem(t *testing.T) {
	tests := []struct {
		name     string
		toStem   string
		lang     string
		expected string
	}{
		{
			name:     "single word ok",
			toStem:   "tavolo",
			lang:     "italian",
			expected: "tavol",
		},
		{
			name:     "multiple words ok",
			toStem:   "tavolo di legno",
			lang:     "italian",
			expected: "tavolo di legn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			s, err := Stem(tt.toStem, tt.lang)

			require.NoError(err)
			require.Equal(tt.expected, s)
		})
	}
}
