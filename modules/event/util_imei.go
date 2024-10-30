package event

type WithIMEI interface {
	SetIMEI(string)
	GetIMEI() string
}

type IMEI struct {
	Imei string
}

func (i *IMEI) SetIMEI(v string) { i.Imei = v }

func (i *IMEI) GetIMEI() string { return i.Imei }
