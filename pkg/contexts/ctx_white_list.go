package contexts

import "slices"

type WhiteList []string

func (wl WhiteList) NeedHandle(imei string) bool {
	return len(wl) == 0 || len(wl) > 0 && slices.Contains(wl, imei)
}
