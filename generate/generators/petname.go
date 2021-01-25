package generators

import (
	"fmt"

	petname "github.com/dustinkirkland/golang-petname"
)

func PetNameGenerator(prefix string, words int, separator string) StringGenerator {
	return func() (string, error) {
		return fmt.Sprintf("%s%s", prefix, petname.Generate(words, separator)), nil
	}
}
