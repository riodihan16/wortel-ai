# Idle CPU/RAM for the release build. WebView2 children scoped to companion's tree.
# Usage: powershell -ExecutionPolicy Bypass -File scripts/perf.ps1
$exe = (Join-Path $PSScriptRoot '..\build\bin\companion.exe' | Resolve-Path).Path
$cores = [int]$env:NUMBER_OF_PROCESSORS
$sp = Start-Process -FilePath $exe -PassThru
$hwnd = 0
for ($i = 0; $i -lt 60; $i++) {
  Start-Sleep -Milliseconds 250
  $sp.Refresh()
  if ($sp.MainWindowHandle -ne 0) { $hwnd = $sp.MainWindowHandle; break }
}
if ($hwnd -eq 0) { Write-Output 'NO_HWND'; exit 1 }
$startupMs = $i * 250

function Get-Tree([int]$root) {
  $all = Get-CimInstance Win32_Process | Select-Object ProcessId, ParentProcessId
  $set = [System.Collections.Generic.HashSet[int]]::new()
  [void]$set.Add($root)
  do {
    $before = $set.Count
    foreach ($p in $all) {
      if ($set.Contains([int]$p.ParentProcessId)) { [void]$set.Add([int]$p.ProcessId) }
    }
  } while ($set.Count -ne $before)
  # Comma prevents PowerShell from unrolling the HashSet into an array.
  return ,$set
}

Start-Sleep -Seconds 6
$tree = Get-Tree $sp.Id
$ids = [int[]]@($tree)
$procs = Get-Process -Id $ids -ErrorAction SilentlyContinue
$cpu0 = ($procs | Measure-Object TotalProcessorTime -Sum).Sum.TotalSeconds
Start-Sleep -Seconds 30
$procs = Get-Process -Id $ids -ErrorAction SilentlyContinue
$cpu1 = ($procs | Measure-Object TotalProcessorTime -Sum).Sum.TotalSeconds

$app = Get-Process -Id $sp.Id
$wv = $procs | Where-Object { $_.Id -ne $sp.Id }
$wsApp = [math]::Round($app.WorkingSet64 / 1MB, 1)
$wsWv = [math]::Round((($wv | Measure-Object WorkingSet64 -Sum).Sum) / 1MB, 1)
$cpu = [math]::Round(100.0 * ($cpu1 - $cpu0) / (30 * $cores), 3)
Write-Output ("STARTUP_MS={0}" -f $startupMs)
Write-Output ("APP idle_cpu_pct={0} ws_mb={1}" -f $cpu, $wsApp)
Write-Output ("WEBVIEW2 procs={0} ws_mb={1}" -f ($wv | Measure-Object).Count, $wsWv)
Write-Output ("TOTAL ws_mb={0}" -f [math]::Round($wsApp + $wsWv, 1))
if ($cpu -lt 1.0) { Write-Output 'IDLE_OK' } else { Write-Output 'IDLE_HIGH' }
Stop-Process -Id $sp.Id -Force -ErrorAction SilentlyContinue
Write-Output 'PERF_DONE'
