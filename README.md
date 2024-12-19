# Использование
0. Перейти в корень проекта
1. **Компиляция байткода:** `go run ./cmd/compiler/main.go <args>`.

Аргументы:

`-in=path_to_input_program`

`-out=path_to_output_bytecode`

`-v` - verbose mode

2. **Запуск виртуальной машины**: `go run ./cmd/vm/main.go <args>`

Аргументы:

`-in=path_to_bytecode`

`-v` - verbose mode

Пример: 

`go run ./cmd/compiler/main.go -in=./samples/input7 -out=./bin/out -v`

`go run ./cmd/vm/main.go -in=./bin/out -v`
