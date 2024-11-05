@echo off
echo Building optimized version...

:: Set version
set VERSION=1.0.7

:: Set environment variables
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

:: Generate Windows resource with both manifest and icon
@REM rsrc -manifest rsrc.manifest -ico logo.ico -o rsrc.syso

:: Copy icon to assets directory
if not exist internal\assets mkdir internal\assets
copy /Y logo.ico internal\assets\

:: Optimize using -ldflags
go build -ldflags="-s -w -H windowsgui -X cursor_history/internal/app.Version=%VERSION%" -trimpath -o CursorHistory.exe ./

:: Create installer directory if not exists
if not exist installer mkdir installer

:: Compile installer
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=%VERSION% setup.iss

echo Build complete!
