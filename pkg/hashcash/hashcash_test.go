package hashcash

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		bits    int
		sender  string
		want    *HashCash
		wantErr error
	}{
		"success":    {bits: 5, sender: "127.0.0.1:5040", wantErr: nil, want: &HashCash{zeroBits: 5, senderData: "127.0.0.1:5040", date: time.Now()}},
		"wrong bits": {bits: 0, sender: "127.0.0.1:5040", wantErr: ErrZeroBitsMustBeMoreThanZero, want: nil},
	}

	for name, tc := range tests {
		got, err := NewHashCash(tc.bits, tc.sender)
		if err != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("%s: want err: %v, got: %v", name, tc.wantErr, err)
			}
		}
		if got != nil {
			if got.senderData != tc.want.senderData || got.zeroBits != tc.want.zeroBits {
				t.Fatalf("%s: want: %v, got: %v", name, tc.want, got)
			}
		}
	}
}

func TestNewFromString(t *testing.T) {
	inputDate := time.Date(2024, 9, 28, 11, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		input   string
		want    *HashCash
		wantErr error
	}{
		"success":                         {input: "1:5:20240928110000:127.0.0.1:5040::YI2Pdg==:MA==", want: &HashCash{zeroBits: 5, date: inputDate, senderData: "127.0.0.1:5040", counter: 0}, wantErr: nil},
		"incorrect header format":         {input: "1:5:5040::YI2Pdg==:MA==", want: nil, wantErr: ErrIncorrectHeaderFormat},
		"incorrect version":               {input: "2:5:20240928110000:127.0.0.1:5040::YI2Pdg==:MA==", want: nil, wantErr: ErrIncorrectVersion},
		"incorrect zerobits not a number": {input: "1:w:20240928110000:127.0.0.1:5040::YI2Pdg==:MA==", want: nil, wantErr: ErrZeroBitsMustBeNumber},
		"incorrect zerobits negative":     {input: "1:-1:20240928110000:127.0.0.1:5040::YI2Pdg==:MA==", want: nil, wantErr: ErrZeroBitsMustBeMoreThanZero},
		"incorrect date":                  {input: "1:5:20240928110:127.0.0.1:5040::YI2Pdg==:MA==", want: nil, wantErr: ErrIncorrectDate},
		"incorrect counter not a number":  {input: "1:5:20240928110000:127.0.0.1:5040::YI2Pdg==:dw==", want: nil, wantErr: ErrCounterMustBePositiveNumber},
		"incorrect counter less than 0":   {input: "1:5:20240928110000:127.0.0.1:5040::YI2Pdg==:LTE=", want: nil, wantErr: ErrCounterMustBePositiveNumber},
	}

	for name, tc := range tests {
		got, err := NewHashCashFromString(tc.input)
		if err != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("%s: want err: %v, got: %v", name, tc.wantErr, err)
			}
		}

		if got != nil {
			if got.senderData != tc.want.senderData || got.zeroBits != tc.want.zeroBits || !got.date.Equal(inputDate) {
				t.Fatalf("%s: want: %v, got: %v", name, tc.want, got)
			}
		}
	}
}

func TestHashCash_Calculate(t *testing.T) {
	tests := map[string]struct {
		input       string
		maxTries    int
		wantCounter int
		wantErr     error
	}{
		"success":            {input: "1:5:20240929083645:127.0.0.1:5040::J+xOIQ==:MA==", maxTries: 1e6, wantCounter: 419882, wantErr: nil},
		"max tries exceeded": {input: "1:5:20240929083645:127.0.0.1:5040::J+xOIQ==:MA==", maxTries: 1e5, wantCounter: 419882, wantErr: ErrMaxTriesExceeded},
	}
	for name, tc := range tests {
		hc, _ := NewHashCashFromString(tc.input)
		err := hc.Calculate(tc.maxTries)
		if err != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("%s: want err: %v, got: %v", name, tc.wantErr, err)
			}
		} else if tc.wantCounter != hc.Counter() {

			t.Fatalf("%s: want: %d, got: %d", name, tc.wantCounter, hc.Counter())
		}

	}
}
