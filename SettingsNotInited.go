package main

type SettingsNotInitedError struct {
	msg string ""
}

func (e *SettingsNotInitedError) Error() string {
	return e.msg
}

func NewSettingsNotInitedError() error {
	return &SettingsNotInitedError{"Settings was not inited"}
}
