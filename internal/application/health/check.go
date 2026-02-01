package health

type Checker interface {
	Name() string
	Check() error
}
