$exe = 'C:/Users/Omama/Documents/1. PROJECT/companion/build/bin/companion.exe'
$shot = 'C:/Users/Omama/Documents/1. PROJECT/companion/shot-m1.png'
Add-Type @"
using System;
using System.Runtime.InteropServices;
public struct RECT { public int Left; public int Top; public int Right; public int Bottom; }
public class Win32 { [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect); }
"@
Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName System.Windows.Forms
$sp = Start-Process -FilePath $exe -PassThru
$hwnd = 0
for ($i = 0; $i -lt 40; $i++) {
  Start-Sleep -Milliseconds 250
  $sp.Refresh()
  if ($sp.MainWindowHandle -ne 0) { $hwnd = $sp.MainWindowHandle; break }
}
if ($hwnd -eq 0) { Write-Output 'NO_HWND'; exit 1 }
Start-Sleep -Seconds 2
$rect = New-Object RECT
[Win32]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
$rw = $rect.Right - $rect.Left
$rh = $rect.Bottom - $rect.Top
Write-Output ("rect={0},{1} {2}x{3}" -f $rect.Left, $rect.Top, $rw, $rh)
$screen = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$m = 80
$x = [math]::Max(0, $rect.Left - $m)
$y = [math]::Max(0, $rect.Top - $m)
$w = [math]::Min($screen.Width - $x, $rw + 2 * $m)
$h = [math]::Min($screen.Height - $y, $rh + 2 * $m)
$bmp = New-Object System.Drawing.Bitmap($w, $h)
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen($x, $y, 0, 0, $bmp.Size)
$bmp.Save($shot, [System.Drawing.Imaging.ImageFormat]::Png)
$g.Dispose(); $bmp.Dispose()
Write-Output ("shot={0} {1}x{2}" -f $shot, $w, $h)
Stop-Process -Name companion -Force -ErrorAction SilentlyContinue
Write-Output 'SHOT_OK'
