package gol

type Frame map[string]Node

type Environment []Frame

func makeDefaultEnvironment() Environment {
	defEnv := []Frame{
	//		Frame{"+": addNum},
	}
	return defEnv
}
