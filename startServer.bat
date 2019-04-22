rem *******************************Code Start*****************************
@echo off
net session >nul 2>&1
if not "%errorLevel%" == "0" (
  echo Set UAC = CreateObject^("Shell.Application"^) > "%temp%\getadmin.vbs"
  echo UAC.ShellExecute "%~s0", "%*", "", "runas", 1 >> "%temp%\getadmin.vbs"
  "%temp%\getadmin.vbs"
  exit /b 2
)
set pa=%~dp0
cd /d %~dp0
nssm install maiyajia.com "%pa%maiyajia.com.exe"
nssm start maiyajia.com
echo "服务启动成功"
pause