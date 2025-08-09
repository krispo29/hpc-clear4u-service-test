@echo off
echo Running MAWB System Integration migrations...
go run cmd/migrate/main.go
pause