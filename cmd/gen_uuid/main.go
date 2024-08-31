package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"gw2_markers_gen/files"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/google/uuid"
)

const DefaultPackageName = "ShellshotMarkerPack"

func main() {
	srcDirectory := *flag.String("s", DefaultPackageName, "Package directory map definitions")
	flag.Parse()

	mapsDir := fmt.Sprintf("%s/%s", srcDirectory, files.MapsDirectory)
	files := files.FilesByExtension(mapsDir, files.MarkerPoiExtension, files.MarkerTrailExtension)
	for _, f := range files {
		addUUID(f, 1) //skip first line
	}
}

func addUUID(fname string, skipLines int) error {
	lines, err := readLines(fname)
	if err != nil {
		return err
	}

	changed := false
	for i := skipLines; i < len(lines); i++ {
		if !strings.Contains(lines[i], `GUID="`) {
			lines[i] = fmt.Sprintf(`%s GUID="%s"`, lines[i], newUUID())
			changed = true
		}
	}
	if changed {
		return writeLines(fname, lines)
	}
	return nil
}

// Read a whole file into the memory and store it as array of lines
func readLines(fname string) ([]string, error) {
	var lines []string
	f, err := os.Open(fname)
	if err != nil {
		return lines, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	buffer := bytes.NewBuffer(make([]byte, 0))

	var part []byte
	var prefix bool
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return lines, err
}

func writeLines(fname string, lines []string) error {
	buf := bytes.NewBuffer([]byte{})
	for i, l := range lines {
		if i > 0 {
			_, err := buf.WriteString("\r\n")
			if err != nil {
				return err
			}
		}
		_, err := buf.WriteString(l)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(fname, buf.Bytes(), fs.ModePerm)
}

func newUUID() string {
	uuid := uuid.New()
	return base64.StdEncoding.EncodeToString(uuid[:])
}
