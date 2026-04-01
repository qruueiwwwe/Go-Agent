package file

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidatePath 测试路径权限校验
func TestValidatePath(t *testing.T) {
	// 获取当前工作目录中的 data 目录
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取工作目录失败: %v", err)
	}
	dataDir := filepath.Join(wd, "../../..", "data")

	fr := NewFileReader(dataDir)

	tests := []struct {
		name    string
		path    string
		valid   bool
		wantErr bool
	}{
		{
			name:    "路径穿透 ../",
			path:    "data/../../../etc/passwd",
			valid:   false,
			wantErr: true,
		},
		{
			name:    "直接使用 ../",
			path:    "../main.go",
			valid:   false,
			wantErr: true,
		},
		{
			name:    "绝对路径 /etc/passwd",
			path:    "/etc/passwd",
			valid:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := fr.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if valid != tt.valid {
				t.Errorf("ValidatePath() got %v, want %v", valid, tt.valid)
			}
		})
	}
}
