package web

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

// PickFolderNative abre o seletor nativo de pastas do Windows e retorna o caminho completo.
func PickFolderNative(ctx context.Context) (string, error) {
	if runtime.GOOS != "windows" {
		return "", nil
	}
	script := `Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.FolderBrowserDialog; $d.Description = 'Selecione a pasta do projeto para monitorar'; $d.UseDescriptionForTitle = $true; if ($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $d.SelectedPath }`
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-STA", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
