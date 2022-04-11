package util

import "strings"

type StrSet struct {
	m map[string]struct{}
}

func NewStrSet(size int) *StrSet {
	return &StrSet{
		m: make(map[string]struct{}, size),
	}
}

func (set *StrSet) String() string {
	return strings.Join(set.List(), ", ")
}

func (set *StrSet) Len() int {
	return len(set.m)
}

func (set *StrSet) Merge(s *StrSet) {
	set.AppendMap(s.m)
}

func (set *StrSet) Add(s string) {
	set.m[s] = struct{}{}
}

func (set *StrSet) Has(s string) bool {
	_, ok := set.m[s]
	return ok
}

func (set *StrSet) Append(list ...string) {
	for _, s := range list {
		set.Add(s)
	}
}

func (set *StrSet) AppendMap(m map[string]struct{}) {
	for k := range m {
		set.Add(k)
	}
}

func (set *StrSet) Map() map[string]struct{} {
	return set.m
}

func (set *StrSet) List() []string {
	arr := make([]string, 0, len(set.m))
	for k := range set.m {
		arr = append(arr, k)
	}
	return arr
}
