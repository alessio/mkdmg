package hdiutil_test

import (
	"os"
	"testing"

	"github.com/alessio/mkdmg/pkg/hdiutil"
)

func TestInit(t *testing.T) {
	hdiutil.SetLogWriter(os.Stderr)
	configs := []hdiutil.Config{
		{
			OutputPath:      "test.dmg",
			VolumeName:      "test",
			VolumeSizeMb:    100,
			SandboxSafe:     true,
			FileSystem:      "APFS",
			SigningIdentity: "",
			ImageFormat:     "ULFO",
			Simulate:        true,
			SourceDir:       "test",
		},
		{
			OutputPath:       "test.dmg",
			VolumeName:       "test",
			FileSystem:       "APFS",
			SigningIdentity:  "",
			HDIUtilVerbosity: 1,
			Simulate:         true,
			SourceDir:        "test",
		},
	}

	type args struct {
		c *hdiutil.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"sandbox_safe_with_volume_size_should_fail", args{&configs[0]}, true},
		{"valid_config_should_succeed", args{&configs[1]}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := hdiutil.New(tt.args.c)
			if err := r.Setup(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if err := r.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
