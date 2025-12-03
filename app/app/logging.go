package app

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)