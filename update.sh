#!/bin/bash
cp $(go env GOROOT)/src/archive/zip/reader.go   internal/zip/
cp $(go env GOROOT)/src/archive/zip/register.go internal/zip/
cp $(go env GOROOT)/src/archive/zip/struct.go   internal/zip/
cp $(go env GOROOT)/src/archive/zip/writer.go   internal/zip/
cp $(go env GOROOT)/LICENSE                     internal/zip/
