# PyBundler

`PyBundler` is a tool that allows you to bundle your Python application into a single binary executable using Go. It creates an embedded Python interpreter and packages your application with it, allowing you to run your Python code without needing a separate Python installation.

This tool is particularly useful for distributing Python applications as standalone executables, making it easier to share and deploy your code without worrying about dependencies or environment setup.

Each entry point will become a separate command in the final binary, allowing you to run different parts of your application with ease. 

> Note: If you only have a single entry point, this will be the root command of the binary.

## Installation
To install `PyBundler`, you need to have Go installed on your system. You can install it using the following command:

```sh
git clone https://github.com/JenspederM/pybundler.git
```

Then navigate to the project directory and run:
```sh
go build -o pybundler main.go
```
This will create an executable named `pybundler` in the current directory.

## Usage
To use `PyBundler`, navigate to the directory containing your Python application and run the following command:

```sh
pybundler bundle --path <path_to_python_files> --output <output_directory> [--overwrite]
```
- `--path`: The path to the directory containing your Python files. This should be the root directory of your Python application.
- `--output`: The directory where the bundled executable will be created.
- `--overwrite`: Optional flag to overwrite the output directory if it already exists.
- `--help`: Print help information.

## Features
- Bundles Python applications into a single binary executable
- Supports multiple entry points
- Easy to use command-line interface
- Cross-platform compatibility (Linux, macOS, Windows)

## Examples

### Your basic Python Application
```sh
go run . bundle --output ./.bundle -p ./examples/basic --overwrite

# Print help
./.bundle/main --help

# Use basic entrypoint
./.bundle/main basic
> Hello from basic!

# Use cli entrypoint
./.bundle/main cli
> Hello from cli!
```

### A Typer Cli Application
```sh
go run . bundle --output ./.bundle-typer -p ./examples/typer-cli --overwrite

# Print help
./.bundle-typer/main --help

# Say hello
./.bundle-typer/main hello John 
> Hello John!

# Say goodbye
./.bundle-typer/main goodbye John 
> Goodbye John!

./.bundle-typer/main goodbye John --formal
> Goodbye John. Have a good day.

```