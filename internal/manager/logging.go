package manager

import "fmt"

func (s *Service) appendLogf(format string, args ...any) {
	s.appendLog(fmt.Sprintf(format, args...))
}
