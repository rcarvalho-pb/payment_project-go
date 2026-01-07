package logging

type Logger interface {
	Info(string, map[string]any)
	Error(string, map[string]any)
}
