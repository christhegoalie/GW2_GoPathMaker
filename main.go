package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"gw2_markers_gen/categories"
	"gw2_markers_gen/maps"
	"io/fs"
	"log"
	"os"
	"strings"
)

const PackageName = "ShellshotMarkerPack"
const PackageRoot = PackageName
const buildPath = "build"

var packageZipName = fmt.Sprintf("%s.zip", PackageName)
var buildFolder = fmt.Sprintf("%s/%s/", buildPath, PackageName)
var outputZipPath = fmt.Sprintf("%s/%s", buildPath, packageZipName)

//outputTacoName := fmt.Sprintf("build/%s.taco", PackageName)

var installScript = func() {}

func main() {
	maps.SetValidation(validateFile)
	categories.SetValidation(validateFile)

	os.RemoveAll(buildPath)
	os.Mkdir(buildPath, fs.ModePerm)

	packageCatagories, warnings, err := categories.Compile(fmt.Sprintf("%s/categories", PackageRoot))
	if err != nil {
		log.Println(err)
		return
	}
	for _, w := range warnings {
		log.Println(w)
	}
	packageMaps, warnings := maps.Compile(packageCatagories, fmt.Sprintf("%s/maps", PackageRoot))
	for _, w := range warnings {
		log.Println(w)
	}

	CopyAssets(fmt.Sprintf("%s/assets", PackageRoot), fmt.Sprintf("%s/assets", buildFolder))
	categories.Save(packageCatagories, buildFolder)
	maps.Save(packageMaps, buildFolder)
	err = makeZip(buildFolder, outputZipPath)
	if err != nil {
		panic(err)
	}
	//makeTaco(outputZipName, outputTacoName)

	installScript()
}

func makeZip(path string, dstfile string) error {
	outFile, err := os.Create(dstfile)
	if err != nil {
		fmt.Println(err)
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	err = addFiles(w, path, "")
	if err != nil {
		return err
	}
	err = w.Close()
	return err
}

func addFiles(w *zip.Writer, basePath, baseInZip string) error {
	//fetch file list
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() { //Write non-directory files to zip
			dat, err := os.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() { //recurse on directories
			newBase := basePath + file.Name() + "/"
			addFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}
	return nil
}

func validateFile(v string) string {
	v = trim(v)
	fname := fmt.Sprintf("%s/%s", PackageRoot, v)
	if _, err := os.Stat(fname); errors.Is(err, os.ErrNotExist) {
		return fmt.Sprintf("File %s not found", v)
	}
	return ""
}

func trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
}
