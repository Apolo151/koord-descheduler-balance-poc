package model

import "sort"

type ResourceVector map[string]float64

func (rv ResourceVector) Clone() ResourceVector {
	if rv == nil {
		return ResourceVector{}
	}
	copy := make(ResourceVector, len(rv))
	for key, value := range rv {
		copy[key] = value
	}
	return copy
}

func (rv ResourceVector) Get(key string) float64 {
	if rv == nil {
		return 0
	}
	return rv[key]
}

func (rv ResourceVector) Set(key string, value float64) {
	if rv == nil {
		return
	}
	rv[key] = value
}

func (rv ResourceVector) Add(other ResourceVector) ResourceVector {
	result := rv.Clone()
	for key, value := range other {
		result[key] += value
	}
	return result
}

func (rv ResourceVector) Sub(other ResourceVector) ResourceVector {
	result := rv.Clone()
	for key, value := range other {
		result[key] -= value
	}
	return result
}

func (rv ResourceVector) Keys() []string {
	if rv == nil {
		return nil
	}
	keys := make([]string, 0, len(rv))
	for key := range rv {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
