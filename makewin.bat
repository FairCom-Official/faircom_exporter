@echo off
setlocal

set target=build-windows-amd64
if NOT "x%1" == "x" (
	set target=%1
)


set BUILD_TIME=%DATE:~-10%_%TIME:~-8%
for /f "usebackq" %%I in (`git describe --tags --always --dirty`) do (
	set VERSION=%%I
)
for /f "usebackq" %%I in (`git rev-parse --short HEAD`) do (
	set COMMIT_SHA=%%I
)

REM invoke nmake, using above variables to override the default values in the makefile
nmake -E -f Makefile-windows %target%


