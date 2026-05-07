@echo off
echo ========================================
echo   Building WebAPI Go Project
echo ========================================
echo.

REM Устанавливаем переменные окружения для Windows
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

REM Скачиваем зависимости
echo [1/3] Downloading dependencies...
go mod download
go mod tidy

REM Собираем проект
echo [2/3] Building project...
REM go build -ldflags="-H windowsgui" -o WebAPI.exe
go build  -o WebAPI.exe

REM Проверяем результат
if exist WebAPI.exe (
    echo.
    echo ========================================
    echo   ✅ Build successful!
    echo   📁 Output: WebAPI.exe
    echo   📏 Size: 
    dir WebAPI.exe | find "WebAPI.exe"
    echo ========================================
) else (
    echo.
    echo ========================================
    echo   ❌ Build failed!
    echo ========================================
)

echo.
pause