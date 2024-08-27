package trailbuilder

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CompileResources(srcPath string) error {
	filesPath := fmt.Sprintf("%s/%s/", srcPath, CompiledAssetsDir)
	dstRoot := fmt.Sprintf("%s/%s/", srcPath, AssetsDir)
	files := readFiles(srcPath)

	for _, f := range files {
		dstPath := dstRoot + strings.TrimPrefix(f, filesPath)
		dstPath = strings.TrimSuffix(dstPath, CompiledTrailExtension) + TrailExtension

		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}
		dstInfo, err := os.Stat(dstPath)
		//Skip recompiling the resource if no changes have been made
		if err == nil && dstInfo.ModTime().After(srcInfo.ModTime()) {
			continue
		}

		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}
		lines := strings.Split(string(b), "\n")
		fileData, err := LinesToTRLBytes(lines)
		if err != nil {
			log.Printf("Error compiling resource: %s, Error: %s", f, err.Error())
			continue
		}

		os.MkdirAll(filepath.Dir(dstPath), fs.ModePerm)
		err = os.WriteFile(dstPath, fileData, fs.ModePerm)
		if err != nil {
			log.Printf("Error saving compiled resource: %s, Error: %s", f, err.Error())
			continue
		}
	}
	return nil
}
