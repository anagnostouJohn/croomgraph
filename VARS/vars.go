package vars

import "sync"

type RoomConfig struct {
	Sensors  SensorsIP `toml:"RoomConfig"`
	Database Database  `toml:"Database"`
}

type Database struct {
	Server   string
	User     string
	Password string
	Port     string
}

type SensorsIP struct {
	Ips []string `toml:"ips"`
}

type AllData struct {
	Room        string
	SensorsData []SensorOrdered
}

type SensorOrdered struct {
	Sensor      string
	Temperature []float64
	Humidity    []float64
	HeatIndex   []float64
}

type RoomData struct {
	Name         string        `json:"name"`
	Date         string        `json:"date"`
	Uptime       string        `json:"uptime"`
	Scale        int           `json:"scale"`
	MacAddr      string        `json:"macaddr"`
	DevType      string        `json:"devtype"`
	Refresh      string        `json:"refresh"`
	Channel      string        `json:"channel"`
	PicVers      string        `json:"picvers"`
	ResetButton  int           `json:"reset_button"`
	Interval     string        `json:"interval"`
	GtmdInterval string        `json:"gtmd_interval"`
	Version      string        `json:"version"`
	Port         int           `json:"port"`
	IP           string        `json:"ip"`
	Serial       string        `json:"serial"`
	GtmdDisabled string        `json:"gtmd_disabled"`
	TimeConfig   TimeConfig    `json:"time_config"`
	EthConfig    EthConfig     `json:"eth_config"`
	Sensors      []Sensor      `json:"sensor"`
	SignalTower  []SignalTower `json:"signal_twr"`
	Relay        []Relay       `json:"relay"`
	SSen         []SwitchSen   `json:"s_sen"`
}

type TimeConfig struct {
	Timezone         string `json:"timezone"`
	Format           string `json:"format"`
	Display          string `json:"display"`
	DaylightSavingEn string `json:"daylight_saving_en"`
}

type EthConfig struct {
	Mtu       int    `json:"mtu"`
	ArpEn     int    `json:"arpen"`
	Negotiate string `json:"negotiate"`
}

type Sensor struct {
	Lab     string `json:"lab"`
	Tf      string `json:"tf"`
	Tc      string `json:"tc"`
	Hf      string `json:"hf"`
	Hc      string `json:"hc"`
	Lf      string `json:"lf"`
	Lc      string `json:"lc"`
	Ala     int    `json:"ala"`
	Profile int    `json:"profile"`
	T       int    `json:"t"`
	En      int    `json:"en"`
	H       string `json:"h,omitempty"`
	Hh      string `json:"hh,omitempty"`
	Lh      string `json:"lh,omitempty"`
	Hi      string `json:"hi,omitempty"`
	Hic     string `json:"hic,omitempty"`
	Hhi     string `json:"hhi,omitempty"`
	Hhic    string `json:"hhic,omitempty"`
	Lhi     string `json:"lhi,omitempty"`
	Lhic    string `json:"lhic,omitempty"`
	Hen     int    `json:"hen,omitempty"`
	Dpc     string `json:"dpc,omitempty"`
	Dpf     string `json:"dpf,omitempty"`
	Dphf    string `json:"dphf,omitempty"`
	Dphc    string `json:"dphc,omitempty"`
	Dplf    string `json:"dplf,omitempty"`
	Dplc    string `json:"dplc,omitempty"`
	Volts   string `json:"volts,omitempty"`
	Highv   string `json:"highv,omitempty"`
	Lowv    string `json:"lowv,omitempty"`
	Units   *Units `json:"units,omitempty"`
}

type Units struct {
	Refmin string `json:"refmin"`
	Refmax string `json:"refmax"`
	Min    string `json:"min"`
	Max    string `json:"max"`
	Sym    string `json:"sym"`
	En     string `json:"en"`
}

type SignalTower struct {
	RE         Status `json:"RE"`
	OR         Status `json:"OR"`
	GR         Status `json:"GR"`
	BL         Status `json:"BL"`
	WH         Status `json:"WH"`
	A1         Status `json:"A1"`
	A2         Status `json:"A2"`
	RY         Status `json:"RY"`
	AttachType int    `json:"attach_type"`
	Lab        string `json:"lab"`
	TowerID    int    `json:"tower_id"`
}

type Status struct {
	En   int `json:"en"`
	Stat int `json:"stat"`
}

type Relay struct {
	Lab     string `json:"lab"`
	Stat    int    `json:"stat"`
	RelayID int    `json:"relay_id"`
}

type SwitchSen struct {
	Lab     string `json:"lab"`
	En      int    `json:"en"`
	Ala     int    `json:"ala"`
	Profile int    `json:"profile"`
	Stat    int    `json:"stat"`
}

type Data struct {
	Value string `json:"value"` // JSON field should match the frontend data key
}

type ListCounter struct {
	Av             int
	CountForAv     int
	DataToRetreave int64
	Mu             sync.Mutex
}

func (l *ListCounter) SetNew(x int, y int64) {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.CountForAv = 0
	l.DataToRetreave = y
	l.Av = x
}

func (l *ListCounter) SetZero() {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.CountForAv = 0
}

func (l *ListCounter) Increase() {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.CountForAv++
}

func (l *ListCounter) Check() bool {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	return l.CountForAv < l.Av

}
