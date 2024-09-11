package framework

type Component interface {
	Render() string
	Update(data string)
}
