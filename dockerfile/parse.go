package dockerfile

import (
	"bufio"
	"fmt"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"io/ioutil"
	"os"
	"strings"
)

type dockerfile struct {
	file *os.File
	ast  *parser.Result
}

// Error types
type updateError struct {
	message string
}

func (e *updateError) Error() string {
	return e.message
}

func Parse(file *os.File) (err error, inst *dockerfile) {
	// Parse the open file handle for the AST
	result, err := parser.Parse(file)
	if err != nil {
		return
	}

	// Ensure the Dockerfile is valid
	_, _, err = instructions.Parse(result.AST)
	if err != nil {
		return
	}

	inst = &dockerfile{
		file: file,
		ast:  result,
	}

	return
}

func (d *dockerfile) Find(needle string) (vars []*parser.Node) {
	for _, node := range d.ast.AST.Children {
		switch node.Value {
		// ARGs are easy.. NAME="value"
		case "arg":

			name, value := SplitArg(node)
			if value == "" {
				// arg contains no value. keep looking
				continue
			}

			// Found the ARG we're looking for. Return it
			if name == needle {
				vars = append(vars, node)
				continue
			}
		default:
			// We can't use this node type yet..
			continue
		}
	}

	return
}

func (d *dockerfile) Update(node *parser.Node, value string) (err error) {
	switch node.Value {
	case "arg":
		return d.updateArg(node, value)
	default:
		return
	}
}

func (d *dockerfile) UpdateArg(name string, value string) (err error) {
	args := d.Find(name)
	if len(args) < 1 {
		return nil
	}
	for _, arg := range args {
		if err = d.Update(arg, value); err != nil {
			return
		}
	}
	return nil
}

func (d *dockerfile) updateArg(node *parser.Node, value string) (err error) {
	// First modify the original line from the node
	parts := strings.SplitN(node.Original, "=", 2)
	if len(parts) != 2 {
		return &updateError{
			fmt.Sprintf("ARG %s has no value to update", parts[0]),
		}
	}

	// Then write that line back to the Dockerfile
	line := parts[0] + "=" + escapeAndQuote(value, parts[1])
	err = d.updateLine(node.StartLine, line)

	// Re-parse the output file to ensure it is still valid
	result, err := parser.Parse(d.file)
	if err != nil {
		return &updateError{"Failed to produce a valid Dockerfile"}
	}
	_, _, err = instructions.Parse(result.AST)
	if err != nil {
		return &updateError{"Failed to produce a valid Dockerfile"}
	}

	return nil
}

func (d *dockerfile) updateLine(n int, line string) (err error) {
	// Always start from the beginning of the file
	d.file.Seek(0, 0)
	scanner := bufio.NewScanner(d.file)

	// Find the end of the part of the file we want to keep
	writeTo := int64(0)
	for ; n > 1; n-- {
		scanner.Scan()
		bytes := scanner.Bytes()
		writeTo += int64(len(bytes)) + 1
	}

	// Scan past the line we're replacing
	if !scanner.Scan() {
		return scanner.Err()
	}
	writeRemainder := writeTo + int64(len(scanner.Bytes()))
	d.file.Seek(writeRemainder, 0)

	// Buffer the remainder of the file into memory
	// TODO: Perhaps do not buffer the entire file into memory
	remainder, err := ioutil.ReadAll(d.file)
	if err != nil {
		return err
	}

	// Write the line and then the rest of the file
	d.file.Seek(writeTo, 0)
	d.file.Write([]byte(line))
	d.file.Write(remainder)

	// Truncate the file to the current location
	pos, err := d.file.Seek(0, 1)
	if err != nil {
		return err
	}
	return d.file.Truncate(pos)
}

func escapeAndQuote(value string, original string) string {

	quote := "\""
	firstChar := rune(original[0])
	isQuoted := false

	// If the first char is single or double quote
	if strings.ContainsRune("\"'", firstChar) {
		quote = string(firstChar)
		isQuoted = true
	}

	// Quote backslashes
	value = strings.Replace(value, "\\", "\\\\", -1)
	// Quote the quotes
	value = strings.Replace(value, quote, "\\"+quote, -1)

	// Check if string contains whitespace characters
	// Always quote if the string was quoted before
	// Always quote if the string contains any kind of quote
	if isQuoted ||
		len(strings.Fields(value)) > 1 ||
		strings.ContainsRune(value, '\\') ||
		strings.ContainsAny(value, "\"'") {
		// Quotes need to be added around the string
		value = quote + value + quote
	}

	return value
}
