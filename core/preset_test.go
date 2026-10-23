package core

import "testing"

func expert(mut func(*ExpertOptions)) *ExpertOptions {
	o := DefaultExpertOptions()
	mut(&o)
	return &o
}

func TestAutoI1IgnoresManualInput(t *testing.T) {
	r := ScanSpec{Preset: PresetStandard, Expert: true, Options: expert(func(o *ExpertOptions) { o.AWGI1 = "<r 4>" })}.Resolve()
	if r.AWGAutoI1 == nil || !*r.AWGAutoI1 || r.AWGI1 != "" {
		t.Fatalf("auto I1 must drop manual input: %+v", r)
	}
}

func TestManualI1Preserved(t *testing.T) {
	for _, v := range []string{"", "<r 4>"} {
		r := ScanSpec{Preset: PresetStandard, Expert: true, Options: expert(func(o *ExpertOptions) { o.AWGAutoI1 = false; o.AWGI1 = v })}.Resolve()
		if *r.AWGAutoI1 || r.AWGI1 != v {
			t.Fatalf("manual I1 %q lost: %+v", v, r)
		}
	}
	r := ScanSpec{Preset: PresetStandard, Expert: false, Options: expert(func(o *ExpertOptions) { o.AWGAutoI1 = false; o.AWGI1 = "<r 4>" })}.Resolve()
	if !*r.AWGAutoI1 || r.AWGI1 != "" {
		t.Fatalf("disabled expert must use auto I1: %+v", r)
	}
}

func TestDisabledExpertUsesDefaults(t *testing.T) {
	r := ScanSpec{Preset: PresetStandard, Expert: false, Options: expert(func(o *ExpertOptions) {
		o.Protocol, o.InnerProtocol, o.IPv6, o.Port, o.TimeoutSec, o.Jobs = ProtocolMASQUEH2, ProtocolAWG, true, 8443, 30, 64
		o.CustomTarget, o.TunnelPingCount, o.IncludeNodes, o.SpeedTest = "example.test", 25, []string{"FRA"}, true
		o.BestBy, o.SweepPorts, o.PingTarget, o.ThroughEndpoint = "speed", "all", "example.com", "127.0.0.1:500"
	})}.Resolve()
	if r.Protocol != ProtocolAWG || r.InnerProtocol != ProtocolWG || r.IPv6 || r.Port != 0 || r.TimeoutSec != 2 || r.Jobs != 10 ||
		r.CustomTarget != "" || r.TunnelPingCount != 0 || len(r.IncludeNodes) != 0 || r.SpeedTest || r.BestBy != "ping" ||
		r.SweepPorts != "" || r.PingTarget != "" || r.ThroughEndpoint != "" {
		t.Fatalf("disabled expert leaked values: %+v", r)
	}
}

func TestEnabledExpertUsesSelected(t *testing.T) {
	r := ScanSpec{Preset: PresetStandard, Expert: true, Options: expert(func(o *ExpertOptions) {
		o.IPv6, o.Port, o.TimeoutSec, o.Jobs, o.TunnelPingCount = true, 2408, 8, 24, 7
		o.BestBy, o.SweepPorts, o.PingTarget = "speed", "open", "one.one.one.one"
	})}.Resolve()
	if !r.IPv6 || r.Port != 2408 || r.TimeoutSec != 8 || r.Jobs != 24 || r.TunnelPingCount != 7 ||
		r.BestBy != "speed" || r.SweepPorts != "open" || r.PingTarget != "one.one.one.one" {
		t.Fatalf("expert values lost: %+v", r)
	}
}

func TestPresets(t *testing.T) {
	if r := (ScanSpec{Preset: PresetDurable}).Resolve(); r.TunnelPingCount != 10 || r.Full || r.SamplePerSubnet != 5 {
		t.Fatalf("durable: %+v", r)
	}
	if r := (ScanSpec{Preset: PresetFull}).Resolve(); r.TunnelPingCount != 10 || !r.Full || r.SamplePerSubnet != 0 {
		t.Fatalf("full: %+v", r)
	}
	if r := (ScanSpec{Preset: PresetFull, Expert: true, Options: expert(func(o *ExpertOptions) { o.TunnelPingCount = 20 })}).Resolve(); r.TunnelPingCount != 20 {
		t.Fatalf("full+expert pings: %+v", r)
	}
	if r := (ScanSpec{Preset: "bogus"}).Resolve(); r.Full || r.TunnelPingCount != 0 {
		t.Fatalf("unknown preset must be standard: %+v", r)
	}
}

func TestPingTargetEnablesBurst(t *testing.T) {
	r := ScanSpec{Preset: PresetStandard, Expert: true, Options: expert(func(o *ExpertOptions) { o.PingTarget = "example.com" })}.Resolve()
	if r.TunnelPingCount != 10 {
		t.Fatalf("got %d", r.TunnelPingCount)
	}
}

func TestProtocolAndIPv6OverridesApplyWithoutExpert(t *testing.T) {
	v6 := true
	r := ScanSpec{Preset: PresetStandard, Protocol: ProtocolMASQUEH3, IPv6: &v6}.Resolve()
	if r.Protocol != ProtocolMASQUEH3 || !r.IPv6 {
		t.Fatalf("%+v", r)
	}
}

func TestNormalizationOfBlankInput(t *testing.T) {
	r := ScanSpec{Expert: true, Options: &ExpertOptions{IncludeNodes: []string{" FRA ", "", "  "}, TimeoutSec: -1}}.Resolve()
	if len(r.IncludeNodes) != 1 || r.IncludeNodes[0] != "FRA" || r.TimeoutSec != 2 || r.Jobs != 10 || r.BestBy != "ping" || r.Protocol != ProtocolAWG {
		t.Fatalf("%+v", r)
	}
}

func TestTelegramOptions(t *testing.T) {
	yes, no := true, false
	if r := (ScanSpec{}).Resolve(); r.Telegram || r.TelegramOnly {
		t.Fatalf("telegram must be off by default: %+v", r)
	}
	if r := (ScanSpec{Telegram: &yes}).Resolve(); !r.Telegram || r.TelegramOnly {
		t.Fatalf("telegram check: %+v", r)
	}
	if r := (ScanSpec{TelegramOnly: &yes}).Resolve(); !r.Telegram || !r.TelegramOnly {
		t.Fatalf("telegram-only must imply the check: %+v", r)
	}
	// the spec-level switch beats a stale value in a disabled expert form
	stale := expert(func(o *ExpertOptions) { o.Telegram = true })
	if r := (ScanSpec{Expert: true, Options: stale, Telegram: &no}).Resolve(); r.Telegram {
		t.Fatalf("explicit off must win: %+v", r)
	}
	if r := (ScanSpec{Expert: false, Options: stale}).Resolve(); r.Telegram {
		t.Fatalf("disabled expert must ignore the form: %+v", r)
	}
}
