package user

import "testing"

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
