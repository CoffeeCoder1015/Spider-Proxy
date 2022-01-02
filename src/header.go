package main

type responseHeader struct {
	hMap map[string]string
}

func (s *responseHeader) add(header string, value string) {
	s.hMap[header] = header + ": " + value + "\r\n"
}

func (s *responseHeader) remove(header string) {
	delete(s.hMap, header)
}

func (s *responseHeader) String() string {
	rt := ""
	for _, v := range s.hMap {
		rt += v
	}
	return rt
}
