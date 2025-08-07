@echo off
rem ##############################################################################
rem mvx Wrapper Script for Windows
rem
rem This script acts as a wrapper for mvx, automatically downloading and 
rem caching the appropriate binary version for Windows.
rem
rem Similar to Maven Wrapper (mvnw.cmd), this allows projects to use mvx without
rem requiring users to install it separately.
rem ##############################################################################

setlocal enabledelayedexpansion

rem Default values
set DEFAULT_MVX_VERSION=latest
set DEFAULT_DOWNLOAD_URL=https://github.com/gnodet/mvx/releases

rem Determine the mvx version to use
set MVX_VERSION_TO_USE=%MVX_VERSION%
if "%MVX_VERSION_TO_USE%"=="" (
    if exist ".mvx\wrapper\mvx-wrapper.properties" (
        for /f "tokens=2 delims==" %%i in ('findstr "^mvxVersion=" ".mvx\wrapper\mvx-wrapper.properties" 2^>nul') do set MVX_VERSION_TO_USE=%%i
    )
)
if "%MVX_VERSION_TO_USE%"=="" (
    if exist ".mvx\version" (
        set /p MVX_VERSION_TO_USE=<".mvx\version"
    )
)
if "%MVX_VERSION_TO_USE%"=="" (
    set MVX_VERSION_TO_USE=%DEFAULT_MVX_VERSION%
)

rem Remove any whitespace
set MVX_VERSION_TO_USE=%MVX_VERSION_TO_USE: =%

rem Determine download URL
set DOWNLOAD_URL_TO_USE=%MVX_DOWNLOAD_URL%
if "%DOWNLOAD_URL_TO_USE%"=="" (
    set DOWNLOAD_URL_TO_USE=%DEFAULT_DOWNLOAD_URL%
)

rem Get user home directory
set HOME_DIR=%USERPROFILE%
if "%HOME_DIR%"=="" set HOME_DIR=%HOMEDRIVE%%HOMEPATH%
if "%HOME_DIR%"=="" set HOME_DIR=.

echo mvx Wrapper
echo Platform: windows-amd64
echo Requested version: %MVX_VERSION_TO_USE%

rem Check for local binary first (priority order)
set LOCAL_BINARIES=.\mvx-local.exe .\mvx-binary.exe .\mvx-dev.exe .\mvx.exe
for %%i in (%LOCAL_BINARIES%) do (
    if exist "%%i" (
        echo Using local mvx binary: %%i
        echo.
        "%%i" %*
        goto :eof
    )
)

rem Resolve version (handle "latest")
set RESOLVED_VERSION=%MVX_VERSION_TO_USE%
if "%MVX_VERSION_TO_USE%"=="latest" (
    echo Resolving latest version...
    call :get_latest_version RESOLVED_VERSION
    if errorlevel 1 (
        echo Error: Could not determine latest version
        exit /b 1
    )
    echo Latest version: !RESOLVED_VERSION!
)

rem Check cached version
set CACHE_DIR=%HOME_DIR%\.mvx\versions\%RESOLVED_VERSION%
set CACHED_BINARY=%CACHE_DIR%\mvx.exe

if exist "%CACHED_BINARY%" (
    echo Using cached mvx binary: %CACHED_BINARY%
    echo.
    "%CACHED_BINARY%" %*
    goto :eof
)

rem Need to download
echo mvx %RESOLVED_VERSION% not found, downloading...

rem Create cache directory
if not exist "%CACHE_DIR%" mkdir "%CACHE_DIR%"

rem Construct download URL
set BINARY_NAME=mvx-windows-amd64.exe
set DOWNLOAD_URL_FULL=%DOWNLOAD_URL_TO_USE%/download/v%RESOLVED_VERSION%/%BINARY_NAME%

echo Downloading mvx %RESOLVED_VERSION% for windows-amd64...
echo Downloading from: %DOWNLOAD_URL_FULL%

rem Download using PowerShell (available on Windows 7+ with .NET 4.0+)
powershell -Command "& {[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%DOWNLOAD_URL_FULL%' -OutFile '%CACHED_BINARY%' -UseBasicParsing}"

if not exist "%CACHED_BINARY%" (
    echo Error: Failed to download mvx binary
    exit /b 1
)

echo Using mvx binary: %CACHED_BINARY%
echo.

rem Execute mvx with all arguments
"%CACHED_BINARY%" %*
goto :eof

rem Function to get latest version from GitHub API
:get_latest_version
set "result_var=%~1"
set API_URL=https://api.github.com/repos/gnodet/mvx/releases/latest

rem Use PowerShell to get the latest version
for /f "delims=" %%i in ('powershell -Command "& {[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $response = Invoke-RestMethod -Uri '%API_URL%' -UseBasicParsing; $response.tag_name -replace '^v', ''}"') do (
    set "%result_var%=%%i"
)

if "!%result_var%!"=="" (
    exit /b 1
)
exit /b 0
