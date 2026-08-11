package pocketbase

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateTime_MarshalJSON(t *testing.T) {
	tm := time.Date(2024, 5, 1, 12, 34, 56, 789_000_000, time.UTC)
	var d DateTime
	if err := d.Scan(tm); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	want := `"2024-05-01 12:34:56.789Z"`
	if string(data) != want {
		t.Errorf("MarshalJSON() = %s, want %s", data, want)
	}
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{`"2024-05-01 12:34:56.789Z"`, "2024-05-01 12:34:56.789Z"},
		{`"2024-05-01T12:34:56Z"`, "2024-05-01 12:34:56.000Z"},
		{`""`, ""},
	}

	for _, tc := range cases {
		var d DateTime
		if err := json.Unmarshal([]byte(tc.raw), &d); err != nil {
			t.Errorf("UnmarshalJSON(%s) failed: %v", tc.raw, err)
			continue
		}
		if got := d.String(); got != tc.want {
			t.Errorf("UnmarshalJSON(%s).String() = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestDateTime_ScanValue(t *testing.T) {
	var d DateTime
	if err := d.Scan("2024-01-15 08:00:00.000Z"); err != nil {
		t.Fatalf("Scan(string) failed: %v", err)
	}
	if got := d.Time().Unix(); got != 1705305600 {
		t.Errorf("Scan(string) Unix() = %d, want 1705305600", got)
	}

	v, err := d.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}
	if v != "2024-01-15 08:00:00.000Z" {
		t.Errorf("Value() = %v", v)
	}

	if err := d.Scan(int64(1705305600)); err != nil {
		t.Fatalf("Scan(unix) failed: %v", err)
	}
	if got := d.String(); got != "2024-01-15 08:00:00.000Z" {
		t.Errorf("Scan(unix).String() = %q", got)
	}
}

func TestDateTime_Parse(t *testing.T) {
	d, err := ParseDateTime("2024-02-02 10:00:00.000Z")
	if err != nil {
		t.Fatalf("ParseDateTime failed: %v", err)
	}
	if d.IsZero() {
		t.Error("ParseDateTime returned zero value")
	}
	if !d.Equal(DateTime{}) && d.Compare(NowDateTime()) >= 0 {
		t.Error("unexpected Compare result")
	}

	if !NowDateTime().After(DateTime{}) {
		t.Error("NowDateTime should be after zero")
	}
}

func TestGeoPoint_MarshalJSON(t *testing.T) {
	p := GeoPoint{Lon: 126.9779, Lat: 37.5665}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if got := string(data); got != `{"lon":126.9779,"lat":37.5665}` {
		t.Errorf("MarshalJSON() = %s", got)
	}

	var decoded GeoPoint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if decoded.Lon != p.Lon || decoded.Lat != p.Lat {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}
