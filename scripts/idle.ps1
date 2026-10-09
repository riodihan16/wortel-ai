$exe = 'C:/Users/Omama/Documents/1. PROJECT/companion/build/bin/companion.exe'
$sp = Start-Process -FilePath $exe -PassThru
$hwnd = 0
for ($i = 0; $i -lt 40; $i++) {
  Start-Sleep -Milliseconds 250
  $sp.Refresh()
  if ($sp.MainWindowHandle -ne 0) { $hwnd = $sp.MainWindowHandle; break }
}
if ($hwnd -eq 0) { Write-Output 'NO_HWND'; exit 1 }
$cores = [int]$env:NUMBER_OF_PROCESSORS
# settle startup, lalu ukur idle 10 detik
Start-Sleep -Seconds 5
$p = Get-Process -Id $sp.Id
$t0 = $p.TotalProcessorTime.TotalSeconds
Start-Sleep -Seconds 10
$p.Refresh()
$t1 = $p.TotalProcessorTime.TotalSeconds
$cpu = [math]::Round(100.0 * ($t1 - $t0) / (10 * $cores), 2)
$ws = [math]::Round($p.WorkingSet64 / 1MB, 1)
Write-Output ("IDLE cores={0} cpu_pct={1} ws_mb={2}" -f $cores, $cpu, $ws)
if ($cpu -lt 1.0) { Write-Output 'IDLE_OK' } else { Write-Output 'IDLE_HIGH' }
Stop-Process -Name companion -Force -ErrorAction SilentlyContinue
Write-Output 'IDLE_DONE'
