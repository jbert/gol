package typ

import "fmt"

type Env []Frame

type Frame map[string]Type

func NewEnv() Env {
	return Env(make([]Frame, 0))
}

func (e Env) WithFrame(f Frame) Env {
	newEnv := []Frame{f}
	newEnv = append(newEnv, e...)
	return newEnv
}

func (e Env) Lookup(s string) (Type, error) {
	for _, f := range []Frame(e) {
		t, ok := f[s]
		if ok {
			return t, nil
		}
	}
	return nil, fmt.Errorf("No type found for identifier [%s]", s)
}

func (e *Env) AddTopLevel(k string, t Type) {
	f0 := (*e)[0]
	f0[k] = t
}
