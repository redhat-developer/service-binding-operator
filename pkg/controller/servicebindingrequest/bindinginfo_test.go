package servicebindingrequest

import "testing"

// TestNewBindingInfo exercises annotation binding information parsing.
func TestNewBindingInfo(t *testing.T) {
	type args struct {
		s string
		d string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *BindingInfo
	}{
		{
			args: args{s: "status.configMapRef-password", d: "binding"},
			want: &BindingInfo{
				FieldPath:  "status.configMapRef",
				Descriptor: "binding:password",
				Path:       "password",
			},
			name:    "{fieldPath}-{path} annotation",
			wantErr: false,
		},
		{
			args: args{s: "status.connectionString", d: "binding"},
			want: &BindingInfo{
				Descriptor: "binding:status.connectionString",
				FieldPath:  "status.connectionString",
				Path:       "status.connectionString",
			},
			name:    "{path} annotation",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBindingInfo(tt.args.s, tt.args.d)
			if err != nil && !tt.wantErr {
				t.Errorf("NewBindingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				requireYamlEqual(t, tt.want, b)
			}
		})
	}
}
