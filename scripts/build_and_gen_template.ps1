$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

go build .

./dsbg.exe -template