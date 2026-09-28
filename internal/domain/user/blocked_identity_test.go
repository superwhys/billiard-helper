package user

import (
	"strconv"
	"testing"
)

func TestNormalizeBlockedIdentity(t *testing.T) {
	tests := []struct {
		name    string
		kind    BlockedIdentityKind
		value   string
		want    string
		wantErr bool
	}{
		{name: "email lowercase", kind: KindEmail, value: "Alice@Example.COM", want: "alice@example.com"},
		{name: "open id preserves case", kind: KindOpenID, value: "WxAbC123", want: "WxAbC123"},
		{name: "unknown kind", kind: "phone", value: "123456789", wantErr: true},
		{name: "empty email", kind: KindEmail, value: "", wantErr: true},
		{name: "invalid email", kind: KindEmail, value: "not-an-email", wantErr: true},
		{name: "empty open id", kind: KindOpenID, value: "", wantErr: true},
		{name: "open id with spaces", kind: KindOpenID, value: " WxAbC123 ", wantErr: true},
		{name: "user id", kind: KindUserID, value: "42", want: "42"},
		{name: "user id leading zeros", kind: KindUserID, value: "00042", want: "42"},
		{name: "maximum user id", kind: KindUserID, value: strconv.FormatUint(uint64(^uint(0)), 10), want: strconv.FormatUint(uint64(^uint(0)), 10)},
		{name: "empty user id", kind: KindUserID, value: "", wantErr: true},
		{name: "zero user id", kind: KindUserID, value: "0", wantErr: true},
		{name: "zero padded user id", kind: KindUserID, value: "000", wantErr: true},
		{name: "negative user id", kind: KindUserID, value: "-1", wantErr: true},
		{name: "signed user id", kind: KindUserID, value: "+1", wantErr: true},
		{name: "overflow user id", kind: KindUserID, value: "18446744073709551616", wantErr: true},
		{name: "hex user id", kind: KindUserID, value: "0x2a", wantErr: true},
		{name: "fractional user id", kind: KindUserID, value: "1.5", wantErr: true},
		{name: "user id invalid character", kind: KindUserID, value: "4a", wantErr: true},
		{name: "user id with spaces", kind: KindUserID, value: " 42 ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeBlockedIdentity(tt.kind, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeBlockedIdentity(%q, %q) error = %v, wantErr %v", tt.kind, tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("NormalizeBlockedIdentity(%q, %q) = %q, want %q", tt.kind, tt.value, got, tt.want)
			}
		})
	}
}
