package blockchain

import "github.com/pkg/errors"

type stacks struct {
	errors []error
}

func (s *stacks) Append(err error, message string, args ...any) {
	s.errors = append(s.errors, errors.Wrapf(err, message, args...))
}

func (s *stacks) TrimLast() {
	if len(s.errors) > 0 {
		s.errors = s.errors[0 : len(s.errors)-1]
	}
}

func (s *stacks) Final() error {
	var final error
	for _, err := range s.errors {
		if err != nil {
			if final == nil {
				final = err
			} else {
				final = errors.Wrap(err, final.Error())
			}
		}
	}
	return final
}
