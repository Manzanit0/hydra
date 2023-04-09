package owner

import (
	"fmt"
	"os"

	"github.com/hairyhenderson/go-codeowners"
)

type Owners struct {
	codeowners.Codeowners
}

func GetCodeOwners() (*Owners, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	c, err := codeowners.FromFile(wd)
	if err != nil {
		return nil, fmt.Errorf("find CODEOWNERS file: %w", err)
	}

	return &Owners{*c}, nil
}

func (c *Owners) Owns(thing, who string) (bool, error) {
	owners := c.Owners(thing)
	for _, o := range owners {
		if o == who {
			return true, nil
		}
	}

	return false, nil
}
