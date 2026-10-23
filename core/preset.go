package core

// Preset is one of the scan presets offered by the clients.
type Preset string

const (
	PresetStandard Preset = "standard"
	PresetDurable  Preset = "durable"
	PresetFull     Preset = "full"

	DefaultProtocol = ProtocolAWG
)

// ExpertOptions are the user-editable fields of the expert form. Its zero
// value is not the default form: use DefaultExpertOptions.
type ExpertOptions struct {
	Protocol         Protocol `json:"protocol"`
	InnerProtocol    Protocol `json:"innerProtocol"`
	IPv6             bool     `json:"ipv6"`
	Port             int      `json:"port"`
	TimeoutSec       int      `json:"timeoutSec"`
	Jobs             int      `json:"jobs"`
	CustomTarget     string   `json:"customTarget"`
	TunnelPingCount  int      `json:"tunnelPingCount"`
	AWGJunkCount     int      `json:"awgJunkCount"`
	AWGJunkMin       int      `json:"awgJunkMin"`
	AWGJunkMax       int      `json:"awgJunkMax"`
	AWGI1            string   `json:"awgI1"`
	AWGAutoI1        bool     `json:"awgAutoI1"`
	MASQUESNI        string   `json:"masqueSni"`
	MASQUEAttempts   int      `json:"masqueAttempts"`
	IncludeNodes     []string `json:"includeNodes"`
	IncludeCountries []string `json:"includeCountries"`
	ExcludeNodes     []string `json:"excludeNodes"`
	ExcludeCountries []string `json:"excludeCountries"`
	MTU              int      `json:"mtu"`
	DNS              []string `json:"dns"`
	SpeedTest        bool     `json:"speedTest"`
	BestBy           string   `json:"bestBy"`
	SweepPorts       string   `json:"sweepPorts"`
	PingTarget       string   `json:"pingTarget"`
	ThroughEndpoint  string   `json:"throughEndpoint"`
	Telegram         bool     `json:"telegram"`
	TelegramOnly     bool     `json:"telegramOnly"`
}

func DefaultExpertOptions() ExpertOptions {
	return ExpertOptions{
		Protocol:       DefaultProtocol,
		InnerProtocol:  ProtocolWG,
		TimeoutSec:     2,
		Jobs:           10,
		AWGAutoI1:      true,
		MASQUEAttempts: 3,
		BestBy:         "ping",
	}
}

// ScanSpec is what a client sends: a preset, whether expert mode is on, and
// the expert form. Protocol and IPv6, when set, apply even with expert mode
// off (the web UI keeps those two outside the expert panel).
type ScanSpec struct {
	Preset   Preset   `json:"preset"`
	Expert   bool     `json:"expert"`
	Protocol Protocol `json:"protocol,omitempty"`
	IPv6     *bool    `json:"ipv6,omitempty"`
	// Telegram and TelegramOnly, when set, apply even with expert mode off.
	Telegram     *bool          `json:"telegram,omitempty"`
	TelegramOnly *bool          `json:"telegramOnly,omitempty"`
	Options      *ExpertOptions `json:"options,omitempty"`
}

// Resolve mirrors the Android ScanOptionsResolver: a disabled expert mode
// falls back to default values, and the preset decides sampling and the
// number of tunnel pings.
func (s ScanSpec) Resolve() ScanOptions {
	preset := s.Preset
	switch preset {
	case PresetStandard, PresetDurable, PresetFull:
	default:
		preset = PresetStandard
	}
	sel := DefaultExpertOptions()
	if s.Expert && s.Options != nil {
		sel = *s.Options
	}
	sel = sel.normalized()
	if s.Protocol != "" {
		sel.Protocol = s.Protocol
	}
	if s.IPv6 != nil {
		sel.IPv6 = *s.IPv6
	}
	if s.Telegram != nil {
		sel.Telegram = *s.Telegram
	}
	if s.TelegramOnly != nil {
		sel.TelegramOnly = *s.TelegramOnly
	}

	var presetPings int
	switch preset {
	case PresetStandard:
		presetPings = sel.TunnelPingCount
	case PresetDurable:
		presetPings = 10
	case PresetFull:
		presetPings = 10
		if s.Expert && sel.TunnelPingCount > 0 {
			presetPings = sel.TunnelPingCount
		}
	}
	pings := presetPings
	if sel.PingTarget != "" && presetPings <= 0 {
		pings = 10
	}

	sample := 5
	if preset == PresetFull {
		sample = 0
	}
	i1 := sel.AWGI1
	if sel.AWGAutoI1 {
		i1 = ""
	}
	autoI1 := sel.AWGAutoI1
	return ScanOptions{
		Protocol:         sel.Protocol,
		InnerProtocol:    sel.InnerProtocol,
		IPv6:             sel.IPv6,
		Port:             sel.Port,
		TimeoutSec:       sel.TimeoutSec,
		Jobs:             sel.Jobs,
		SamplePerSubnet:  sample,
		Full:             preset == PresetFull,
		TunnelPingCount:  pings,
		CustomTarget:     sel.CustomTarget,
		AWGJunkCount:     sel.AWGJunkCount,
		AWGJunkMin:       sel.AWGJunkMin,
		AWGJunkMax:       sel.AWGJunkMax,
		AWGI1:            i1,
		AWGAutoI1:        &autoI1,
		MASQUESNI:        sel.MASQUESNI,
		MASQUEAttempts:   sel.MASQUEAttempts,
		IncludeNodes:     sel.IncludeNodes,
		IncludeCountries: sel.IncludeCountries,
		ExcludeNodes:     sel.ExcludeNodes,
		ExcludeCountries: sel.ExcludeCountries,
		MTU:              sel.MTU,
		DNS:              sel.DNS,
		SpeedTest:        sel.SpeedTest,
		BestBy:           sel.BestBy,
		SweepPorts:       sel.SweepPorts,
		PingTarget:       sel.PingTarget,
		ThroughEndpoint:  sel.ThroughEndpoint,
		Telegram:         sel.Telegram || sel.TelegramOnly,
		TelegramOnly:     sel.TelegramOnly,
	}
}

// normalized replaces missing or nonsensical values with defaults and drops
// blank list entries, so a hand-written or older request still resolves.
func (o ExpertOptions) normalized() ExpertOptions {
	d := DefaultExpertOptions()
	if o.Protocol == "" {
		o.Protocol = d.Protocol
	}
	if o.InnerProtocol == "" {
		o.InnerProtocol = d.InnerProtocol
	}
	if o.TimeoutSec <= 0 {
		o.TimeoutSec = d.TimeoutSec
	}
	if o.Jobs <= 0 {
		o.Jobs = d.Jobs
	}
	if o.MASQUEAttempts <= 0 {
		o.MASQUEAttempts = d.MASQUEAttempts
	}
	if o.BestBy == "" {
		o.BestBy = d.BestBy
	}
	if o.Port < 0 {
		o.Port = 0
	}
	if o.TunnelPingCount < 0 {
		o.TunnelPingCount = 0
	}
	o.IncludeNodes = cleanList(o.IncludeNodes)
	o.IncludeCountries = cleanList(o.IncludeCountries)
	o.ExcludeNodes = cleanList(o.ExcludeNodes)
	o.ExcludeCountries = cleanList(o.ExcludeCountries)
	o.DNS = cleanList(o.DNS)
	return o
}

func cleanList(in []string) []string {
	var out []string
	for _, v := range in {
		if v = trimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
