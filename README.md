# InkyBlackness Chunkie

This is a tool as part of the [InkyBlackness](https://inkyblackness.github.io) project, written in [Go](http://golang.org/). This tool provides import/export access to resource files for modification of media content.

## Usage

### Command Line Interface

```
Usage:
  chunkie export <resource-file> <chunk-id> [--block=<block-id>] [<folder>]
  chunkie -h | --help
  chunkie --version

Options:
  <resource-file>     The resource file to work on.
  <chunk-id>          The chunk identifier. Defaults to decimal, use "0x" as prefix for hexadecimal.
  --block=<block-id>  The block identifier. Defaults to decimal, use "0x" as prefix for hexadecimal. [default: 0]
  <folder>            The path of the folder to use. [default: .]
  -h --help           Show this screen.
  --version           Show version.
```

The base file name of files is ```XXXX_YYY.ZZZ```. XXXX is the hexadecimal presentation of the chunk number. YYY is decimal for the block number. ZZZ is the type of the file, defaulting to ```bin```.

## License

The project is available under the terms of the **New BSD License** (see LICENSE file).
