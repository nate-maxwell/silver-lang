// Package astcache stores parsed Silver programs on disk.
package astcache

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"silver/ast"
)

const (
	// Version must change whenever the serialized AST representation or the
	// parse/optimization pipeline changes.
	Version uint32 = 10

	cacheSuffix   = ".astc"
	maxPathLength = 1 << 20
	maxCacheSize  = 64 << 20
)

var magic = [8]byte{'S', 'L', 'V', 'R', 'A', 'S', 'T', 0}

func init() {
	// Statements and expressions are stored behind interfaces in the AST, so
	// gob needs every concrete implementation registered before decoding.
	gob.Register(&ast.BlockStatement{})
	gob.Register(&ast.ExpressionStatement{})
	gob.Register(&ast.LetStatement{})
	gob.Register(&ast.MemberAssignmentStatement{})
	gob.Register(&ast.ReturnStatement{})
	gob.Register(&ast.EnumStatement{})
	gob.Register(&ast.StructStatement{})

	gob.Register(&ast.ArrayLiteral{})
	gob.Register(&ast.Boolean{})
	gob.Register(&ast.CallExpression{})
	gob.Register(&ast.FloatLiteral{})
	gob.Register(&ast.FunctionLiteral{})
	gob.Register(&ast.HashLiteral{})
	gob.Register(&ast.Identifier{})
	gob.Register(&ast.IfExpression{})
	gob.Register(&ast.ImportExpression{})
	gob.Register(&ast.IndexExpression{})
	gob.Register(&ast.InfixExpression{})
	gob.Register(&ast.IntegerLiteral{})
	gob.Register(&ast.MemberExpression{})
	gob.Register(&ast.PrefixExpression{})
	gob.Register(&ast.StringLiteral{})
	gob.Register(&ast.StructLiteral{})
}

// Path returns the cache filename associated with a source file.
func Path(sourcePath string) string {
	return sourcePath + cacheSuffix
}

// Load returns a cached program only when its format, source path, and source
// contents match. Any cache problem is reported as a miss so callers can fall
// back to parsing the source.
func Load(sourcePath string, source []byte) (*ast.Program, bool) {
	file, err := os.Open(Path(sourcePath))
	if err != nil {
		return nil, false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.Size() > maxCacheSize {
		return nil, false
	}

	reader := bufio.NewReader(file)
	var cachedMagic [len(magic)]byte
	if _, err := io.ReadFull(reader, cachedMagic[:]); err != nil || cachedMagic != magic {
		return nil, false
	}

	var version uint32
	if err := binary.Read(reader, binary.BigEndian, &version); err != nil || version != Version {
		return nil, false
	}

	var cachedHash [sha256.Size]byte
	if _, err := io.ReadFull(reader, cachedHash[:]); err != nil || cachedHash != sha256.Sum256(source) {
		return nil, false
	}

	var pathLength uint32
	if err := binary.Read(reader, binary.BigEndian, &pathLength); err != nil || pathLength > maxPathLength {
		return nil, false
	}
	cachedPath := make([]byte, pathLength)
	if _, err := io.ReadFull(reader, cachedPath); err != nil || string(cachedPath) != sourcePath {
		return nil, false
	}

	var program *ast.Program
	if err := gob.NewDecoder(reader).Decode(&program); err != nil || program == nil {
		return nil, false
	}
	return program, true
}

// Store atomically writes program's cache. The source file remains the source
// of truth; callers may safely ignore a returned error.
func Store(sourcePath string, source []byte, program *ast.Program) error {
	if program == nil {
		return errors.New("cannot cache a nil AST")
	}

	cachePath := Path(sourcePath)
	temporary, err := os.CreateTemp(filepath.Dir(cachePath), "."+filepath.Base(cachePath)+"-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	writer := bufio.NewWriter(temporary)
	writeError := write(writer, sourcePath, source, program)
	if writeError == nil {
		writeError = writer.Flush()
	}
	if closeError := temporary.Close(); writeError == nil {
		writeError = closeError
	}
	if writeError != nil {
		return writeError
	}

	if err := os.Rename(temporaryPath, cachePath); err != nil {
		return fmt.Errorf("replace AST cache: %w", err)
	}
	return nil
}

func write(writer io.Writer, sourcePath string, source []byte, program *ast.Program) error {
	if _, err := writer.Write(magic[:]); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, Version); err != nil {
		return err
	}
	digest := sha256.Sum256(source)
	if _, err := writer.Write(digest[:]); err != nil {
		return err
	}
	pathBytes := []byte(sourcePath)
	if len(pathBytes) > maxPathLength {
		return errors.New("source path is too long to cache")
	}
	if err := binary.Write(writer, binary.BigEndian, uint32(len(pathBytes))); err != nil {
		return err
	}
	if _, err := io.Copy(writer, bytes.NewReader(pathBytes)); err != nil {
		return err
	}
	return gob.NewEncoder(writer).Encode(program)
}
