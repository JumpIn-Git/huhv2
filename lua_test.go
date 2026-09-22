package main

import (
	"os"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test uses 3088700 (Driving Rogue) with keys replaced with junk

func newTestApp(t *testing.T) *App {
	t.Helper()
	root := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	app := &App{
		SlsConfig:     root,
		SLSconfigPath: t.TempDir() + "/config.yaml",
		Depotcache:    t.TempDir(),
		Out:           NewOutput(os.Stdout, false),
	}
	for _, f := range []struct {
		key    string
		target **yaml.Node
		kind   yaml.Kind
		tag    string
	}{
		{"AdditionalApps", &app.AdditionalApps, yaml.SequenceNode, "!!seq"},
		{"AdditionalDepots", &app.AdditionalDepots, yaml.SequenceNode, "!!seq"},
		{"DecryptionKeys", &app.DecryptionKeys, yaml.MappingNode, "!!map"},
		{"ManifestIds", &app.ManifestIds, yaml.MappingNode, "!!map"},
		{"AppTokens", &app.AppTokens, yaml.MappingNode, "!!map"},
	} {
		res, err := EnsureKey(root, f.key, f.kind, f.tag)
		if err != nil {
			t.Fatalf("EnsureKey: %v", err)
		}
		*f.target = res
	}
	return app
}

func TestParseLuaAddsMainAppidToAdditionalApps(t *testing.T) {
	a := newTestApp(t)
	luab := []byte(`addappid(3088700, 1, "deadbeef000000000000000000000000000000000000000000000000000000ab")
`)
	if err := a.parseLua(luab, 3088700, "Driving Rogue"); err != nil {
		t.Fatalf("parseLua: %v", err)
	}
	if !SeqContainsInt(a.AdditionalApps, 3088700, "") {
		t.Errorf("AdditionalApps should contain 3088700, has: %v", seqValues(a.AdditionalApps))
	}
	if !mapHasKey(a.DecryptionKeys, 3088700) {
		t.Errorf("DecryptionKeys should contain 3088700")
	}
}

func TestParseLuaAddsDepotsAndKeys(t *testing.T) {
	a := newTestApp(t)
	luab := []byte(`addappid(3088700, 1, "deadbeef000000000000000000000000000000000000000000000000000000ab")
addappid(3088701, 1, "feedface000000000000000000000000000000000000000000000000000000cd")
setManifestid(3088701, "5509957476394576514")
addappid(228989, 1, "cafebabe000000000000000000000000000000000000000000000000000000ef")
setManifestid(228989, "5753583882400741046")
`)
	if err := a.parseLua(luab, 3088700, "Driving Rogue"); err != nil {
		t.Fatalf("parseLua: %v", err)
	}
	wantApps := []int{3088700}
	for _, want := range wantApps {
		if !SeqContainsInt(a.AdditionalApps, want, "") {
			t.Errorf("AdditionalApps should contain %d, has: %v", want, seqValues(a.AdditionalApps))
		}
	}
	wantDepots := []int{3088701, 228989}
	for _, want := range wantDepots {
		if !SeqContainsInt(a.AdditionalDepots, want, "") {
			t.Errorf("AdditionalDepots should contain %d, has: %v", want, seqValues(a.AdditionalDepots))
		}
	}
	wantKeys := []int{3088700, 3088701, 228989}
	for _, want := range wantKeys {
		if !mapHasKey(a.DecryptionKeys, want) {
			t.Errorf("DecryptionKeys should contain %d", want)
		}
	}
	wantManifests := map[int]string{
		3088701: "5509957476394576514",
		228989:  "5753583882400741046",
	}
	for k, v := range wantManifests {
		if !mapHasValue(a.ManifestIds, k, v) {
			t.Errorf("ManifestIds should have %d = %q", k, v)
		}
	}
}

func TestParseLuaDoesNotDuplicateMainAppid(t *testing.T) {
	a := newTestApp(t)
	luab := []byte(`addappid(3088700, 1, "deadbeef000000000000000000000000000000000000000000000000000000ab")
`)
	if err := a.parseLua(luab, 3088700, "Driving Rogue"); err != nil {
		t.Fatalf("parseLua: %v", err)
	}
	count := countInSeq(a.AdditionalApps, 3088700)
	if count != 1 {
		t.Errorf("AdditionalApps should contain 3088700 exactly once, got %d", count)
	}
}

func TestParseLuaAddsMainAppidViaSimpleCall(t *testing.T) {
	a := newTestApp(t)
	luab := []byte(`addappid(3088700)
`)
	if err := a.parseLua(luab, 3088700, "Driving Rogue"); err != nil {
		t.Fatalf("parseLua: %v", err)
	}
	if !SeqContainsInt(a.AdditionalApps, 3088700, "") {
		t.Errorf("AdditionalApps should contain 3088700, has: %v", seqValues(a.AdditionalApps))
	}
}

// --- helpers ---

func seqValues(seq *yaml.Node) []int {
	if seq == nil {
		return nil
	}
	vals := make([]int, 0, len(seq.Content))
	for _, item := range seq.Content {
		v, _ := strconv.Atoi(item.Value)
		vals = append(vals, v)
	}
	return vals
}

func mapHasKey(m *yaml.Node, key int) bool {
	if m == nil {
		return false
	}
	want := strconv.Itoa(key)
	for i := 0; i < len(m.Content); i += 2 {
		if m.Content[i].Value == want {
			return true
		}
	}
	return false
}

func mapHasValue(m *yaml.Node, key int, wantVal string) bool {
	if m == nil {
		return false
	}
	want := strconv.Itoa(key)
	for i := 0; i < len(m.Content); i += 2 {
		if m.Content[i].Value == want {
			return m.Content[i+1].Value == wantVal
		}
	}
	return false
}

func countInSeq(seq *yaml.Node, value int) int {
	if seq == nil {
		return 0
	}
	want := strconv.Itoa(value)
	count := 0
	for _, item := range seq.Content {
		if item.Value == want {
			count++
		}
	}
	return count
}
