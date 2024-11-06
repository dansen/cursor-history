@echo off
setlocal EnableDelayedExpansion

:: Check if version parameter is provided
if "%~1"=="" (
    echo Please provide a version number, e.g.: release.bat 1.0.7
    exit /b 1
)

set VERSION=%~1

:: Update version in internal/app/version.go
echo Updating version.go...
powershell -Command "(Get-Content internal/app/version.go) -replace 'Version = \".*\"', 'Version = \"%VERSION%\"' | Set-Content internal/app/version.go -Encoding UTF8"

:: Update version in build.bat
echo Updating build.bat...
powershell -Command "(Get-Content build.bat) -replace 'set VERSION=.*', 'set VERSION=%VERSION%' | Set-Content build.bat -Encoding UTF8"

:: Update version in setup.iss
echo Updating setup.iss...
powershell -Command "(Get-Content setup.iss) -replace '#define AppVersion \".*\"', '#define AppVersion \"%VERSION%\"' | Set-Content setup.iss -Encoding UTF8"

:: Git operations
echo Committing changes...
git add internal/app/version.go build.bat setup.iss
git commit -m "release: v%VERSION%"

:: Create and push tag
echo Creating tag v%VERSION%...
git tag v%VERSION%
git push origin v%VERSION%
git push

echo Version v%VERSION% released successfully!
