@echo off
cd /d "%~dp0\..\..\.."
set PYTHONPATH=src
python -m src.cli.main %*
echo.
pause
