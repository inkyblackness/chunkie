# InkyBlackness Chunkie

This is a tool as part of the [InkyBlackness](https://inkyblackness.github.io) project, written in [Go](http://golang.org/). This tool provides import/export access to resource files for modification of media content.

## Usage

### Command Line Interface

```
Usage:
  chunkie export <resource-file> <chunk-id> [--block=<block-id>] [--raw] [<folder>]
  chunkie import <resource-file> <chunk-id> [--block=<block-id>] [--data-type=<id>] <source-file>
  chunkie -h | --help
  chunkie --version

Options:
  <resource-file>     The resource file to work on.
  <chunk-id>          The chunk identifier. Defaults to decimal, use "0x" as prefix for hexadecimal.
  --block=<block-id>  The block identifier. Defaults to decimal, use "0x" as prefix for hexadecimal. [default: 0]
	--raw               With this flag, the chunk will be exported without conversion to a common file format.
	--data-type=<id>    The type of the chunk to write.
  <folder>            The path of the folder to use. [default: .]
	<source-file>       The source file to import.
  -h --help           Show this screen.
  --version           Show version.
```

The base file name of files is ```XXXX_YYY.ZZZ```. XXXX is the hexadecimal presentation of the chunk number. YYY is decimal for the block number. ZZZ is the type of the file, defaulting to ```bin```.

For exporting, audio chunks are saved as .wav files. Specifying --raw will export the chunk in its raw format.
Files are imported raw as well, unless a conversion is known. This is the case for .wav files when writing sound chunks.

## License

The project is available under the terms of the **New BSD License** (see LICENSE file).
