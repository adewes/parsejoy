# Parsejoy

Parsejoy is a (currently highly experimental and incomplete) Python/Go/C++
framework for rapid parser prototyping and language development. It allows you
to write simple EBNF-style YAML grammars and use them to interactively parse
files without having to compile a single line of code. The aim of this tool
is to make it easier to rapidly write and test parsers for different formats
and programming languages.

As an example, the toolkit contains a fully-functional Python grammar adapted
from the official EBNF-grammar including the tokenizer.

## Current Status

This library is highly experimental and unfinished, please do not use this code
except for experimenting with it.

## Organization

Currently there are three reference implementations of the Parsejoy parser, one
written in C++, one in Go(lang) and one in Python. There are several
experimental parser scripts in Python that implement e.g. a GLR parser.

## Performance

The performance of the parser is (as is expected) worse than that of a
hand-coded parser or one generated with a tool like Bison. The aim is to be able
to parse at least 50.000 lines of code / second on a standard computer (2016),
which is sufficiently fast for the intended use cases of the library.

## License

The whole repository is licensed under MIT, except for included third-party code
, where the license is indicated in the file header.