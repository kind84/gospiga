package stemmer

import (
	"github.com/tebeka/snowball"
)

// Stem returns the stemming of the provided word.
func Stem(w, lang string) (string, error) {
	s, err := snowball.New(lang)
	if err != nil {
		return "", err
	}
	return s.Stem(w), nil
}
