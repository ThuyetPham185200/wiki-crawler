package safemap

import (
	"fmt"
	"sync"
)

type SafeSetMap struct {
	mtx  sync.RWMutex
	data map[string]map[string]struct{}
}

func NewSafeSetMap() *SafeSetMap {
	s := &SafeSetMap{}
	s.data = make(map[string]map[string]struct{})
	return s
}

func (s *SafeSetMap) Make(k string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.data[k]; !ok {
		s.data[k] = make(map[string]struct{})
	}
}

func (s *SafeSetMap) Set(k, v string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.data[k]; !ok {
		s.data[k] = make(map[string]struct{})
	}
	if _, ok := s.data[k][v]; !ok {
		s.data[k][v] = struct{}{}
	}
}

func (s *SafeSetMap) GetValue(k string) []string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if vals, ok := s.data[k]; ok {
		res := make([]string, 0)
		for key := range vals {
			res = append(res, key)
		}
		return res
	}
	return nil
}

// PrintAll in toàn bộ nội dung của SafeSetMap
func (s *SafeSetMap) PrintAll() {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	fmt.Println("SafeSetMap contents:===================================================================")
	for k, vSet := range s.data {
		fmt.Printf("Key: %s -> Values: [", k)
		first := true
		for v := range vSet {
			if !first {
				fmt.Print(", ")
			}
			fmt.Print(v)
			first = false
		}
		fmt.Println("]")
	}
	fmt.Println("==========================================================================================")
}
