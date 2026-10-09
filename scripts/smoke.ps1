Start-Process -FilePath 'C:/Users/Omama/Documents/1. PROJECT/companion/build/bin/companion.exe'
Start-Sleep -Seconds 8
$p = Get-Process companion -ErrorAction SilentlyContinue
if (-not $p) { Write-Output 'PROC_NOT_FOUND'; exit 1 }
foreach ($proc in $p) {
  $ws = [math]::Round($proc.WorkingSet64 / 1MB, 1)
  Write-Output ("MAIN id={0} ws_mb={1} hwnd={2} title={3}" -f $proc.Id, $ws, $proc.MainWindowHandle, $proc.MainWindowTitle)
}
Stop-Process -Name companion -Force
Start-Sleep -Seconds 2
$left = Get-Process companion -ErrorAction SilentlyContinue
if ($left) { Write-Output 'KILL_FAIL'; exit 1 }
Write-Output 'SMOKE_OK'
