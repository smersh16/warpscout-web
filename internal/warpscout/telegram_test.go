package warpscout

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestReachedDCsMaskAndSlowest(t *testing.T) {
	answer := map[string]time.Duration{
		telegramDCs[0]: 10 * time.Millisecond,
		telegramDCs[2]: 40 * time.Millisecond,
		telegramDCs[4]: 20 * time.Millisecond,
	}
	dial := func(ctx context.Context, addr string) (net.Conn, error) {
		d, ok := answer[addr]
		if !ok {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		select {
		case <-time.After(d):
			a, b := net.Pipe()
			b.Close()
			return a, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	rtt, reached := reachedDCs(context.Background(), dial, telegramDCs, 300*time.Millisecond)
	if reached != 0b10101 {
		t.Fatalf("reached mask = %05b, want 10101", reached)
	}
	if rtt < 40*time.Millisecond || rtt > 200*time.Millisecond {
		t.Fatalf("rtt must be the slowest answer (~40ms), got %v", rtt)
	}
	if reached == allDCs {
		t.Fatal("3 of 5 DCs must not count as Telegram-ok")
	}
	if got := tgPartialStr(reached); got != "3/5" {
		t.Fatalf("partial = %q", got)
	}
	if got := dcNames(allDCs &^ reached); got != "DC2, DC4" {
		t.Fatalf("missing DCs = %q", got)
	}
}

func TestReachedDCsAllBlockedCostsOneTimeout(t *testing.T) {
	dial := func(ctx context.Context, addr string) (net.Conn, error) {
		<-ctx.Done()
		return nil, errors.New("blocked")
	}
	start := time.Now()
	_, reached := reachedDCs(context.Background(), dial, telegramDCs, 150*time.Millisecond)
	if reached != 0 || tgPartialStr(reached) != "blocked" {
		t.Fatalf("reached = %05b", reached)
	}
	if elapsed := time.Since(start); elapsed > 450*time.Millisecond {
		t.Fatalf("blocked exit must cost one timeout, not five: %v", elapsed)
	}
}

func TestMTProtoProbeAcceptsFramedAnswer(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go func() {
		defer server.Close()
		buf := make([]byte, 48)
		if _, err := io.ReadFull(server, buf); err != nil {
			return
		}
		if buf[0] != 0xee || binary.LittleEndian.Uint32(buf[4:8]) != 40 {
			return
		}
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], 84)
		server.Write(hdr[:])
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := mtprotoProbe(ctx, client); err != nil {
		t.Fatal(err)
	}
}

func TestMTProtoProbeRejectsGarbageAndSilence(t *testing.T) {
	// A DPI box that completes the TCP handshake and then eats the payload must fail.
	client, server := net.Pipe()
	go func() { io.Copy(io.Discard, server) }()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if err := mtprotoProbe(ctx, client); err == nil {
		t.Fatal("silence must not count as an answer")
	}
	client.Close()
	server.Close()

	client, server = net.Pipe()
	defer client.Close()
	go func() {
		defer server.Close()
		io.ReadFull(server, make([]byte, 48))
		server.Write([]byte{0xff, 0xff, 0xff, 0x7f}) // absurd frame length
	}()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	if err := mtprotoProbe(ctx2, client); err == nil {
		t.Fatal("implausible frame length must be rejected")
	}
}
