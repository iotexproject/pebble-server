package event

type WithIMEI interface {
	SetIMEI(string)
	GetIMEI() string
}

type IMEI struct {
	imei string
}

func (i *IMEI) SetIMEI(v string) { i.imei = v }

func (i *IMEI) GetIMEI() string { return i.imei }
