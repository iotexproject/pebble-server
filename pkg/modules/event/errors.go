package event

type UnmarshalError struct{}

func (e *UnmarshalError) Error() string {
	return ""
}

type UnmarshalTopicError struct{}

func (e *UnmarshalTopicError) Error() string {
	return ""
}

type ValidateError struct {
}

func (e *ValidateError) Error() string {
	return ""
}

type HandleError struct {
}

func (e *HandleError) Error() string {
	return ""
}

type RegistryError struct {
	v Event
}
