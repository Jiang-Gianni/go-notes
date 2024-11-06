package dt

type Type struct {
	name string
}

func Hello(name string) Type {
	return Type{name: name}
}
