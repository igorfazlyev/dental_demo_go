@echo off
echo Checking DentalAI Platform Setup...
echo.

echo 1. Checking Go installation...
go version >nul 2>&1
if %errorlevel% == 0 (
    echo    [OK] Go is installed
    go version
) else (
    echo    [ERROR] Go is NOT installed
    echo    Please install from: https://golang.org/dl/
    exit /b 1
)

echo.
echo 2. Checking directory structure...

if exist "main.go" (
    echo    [OK] main.go found
) else (
    echo    [ERROR] main.go NOT found
    exit /b 1
)

if exist "templates" (
    echo    [OK] templates directory found
) else (
    echo    [ERROR] templates directory NOT found
    exit /b 1
)

if exist "static\css" (
    echo    [OK] static/css directory found
) else (
    echo    [ERROR] static/css directory NOT found
    exit /b 1
)

if exist "static\css\style.css" (
    echo    [OK] style.css found
) else (
    echo    [ERROR] style.css NOT found
)

echo.
echo 3. Listing template files...
dir /b templates\*.html 2>nul || echo    [ERROR] No template files found

echo.
echo [OK] Setup looks good!
echo.
echo To run the application:
echo    go run main.go
echo.
echo Then open: http://localhost:8080
