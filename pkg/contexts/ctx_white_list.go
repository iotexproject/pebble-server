package contexts

import "slices"

type WhiteList []string

func (wl WhiteList) NeedHandle(imei string) bool {
	last := imei[len(imei)-1]
	return len(wl) == 0 || len(wl) > 0 && slices.Contains(wl, imei) || last == '0'
}
