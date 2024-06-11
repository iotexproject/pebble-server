package models

const (
	CREATED DeviceStatus = iota
	PROPOSAL
	CONFIRM
)

type DeviceStatus int32

type Devices []*Device

type Device struct {
	ID                       string `json:"id"`                          // id varchar(64)
	Name                     string `json:"name"`                        // name varchar(64)
	Owner                    string `json:"owner" index:""`              // owener varchar(64)
	Address                  string `json:"address" uinqueIdex:""`       // address varchar(64)
	Avatar                   string `json:"avatar"`                      // avatar varchar(128)
	Status                   int32  `json:"status"`                      // status int(11)
	Proposer                 string `json:"proposer"`                    // proposer varchar(64)
	Firmware                 string `json:"firmware"`                    // firmware varchar(32)
	Config                   string `json:"config"`                      // config varchar(32)
	TotalGas                 int32  `json:"total_gas"`                   // total_gas int(11)
	BulkUpload               int32  `json:"bulk_upload"`                 // bulk_upload int(11)
	DataChannel              int32  `json:"data_channel"`                // data_channel int(11)
	UploadPeriod             int32  `json:"upload_period"`               // upload_period int(11)
	BulkUploadSamplingCnt    int32  `json:"bulk_upload_sampling_cnt"`    // bulk_upload_sampling_cnt int(11)
	BulkUploadSamplingPeriod int32  `json:"bulk_upload_sampling_period"` // bulk_upload_sampling_freq int(11)
	Beep                     int32  `json:"beep"`                        // beep int(11)
	RealFirmware             string `json:"real_firmware"`               // real_firmware varchar(32)
	State                    int32  `json:"state"`                       // state int(11)
	Type                     int32  `json:"type"`                        // type int(11)
	Configurable             bool   `json:"configurable"`                // configurable tinyint(1)
}
