package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"gw2_markers_gen/categories"
	"gw2_markers_gen/files"
	"gw2_markers_gen/maps"
	trailbuilder "gw2_markers_gen/trail_builder"
	"gw2_markers_gen/utils"
	"io/fs"
	"log"
	"os"
)

const DefaultPackageName = "ShellshotMarkerPack"
const buildPath = "build"

var srcDirectory string

// Custom install function
// Input: relative marker pack location (EX: build/MarkerPack.taco)
// This is run on marker pack build, and can be used to automate marker pack installation
// You can override this method with a local init file.
// See installer.go.example for example code (copy the file as "installer.go")
var installScript = func(packageFile string) {}
var preInstallScript = func(packageFile string) {}

func main() {
	outputPackage := *flag.String("n", DefaultPackageName, "Output Package Name")
	srcDirectory = *flag.String("s", outputPackage, "Package directory containing definition")
	flag.Parse()

	packageZipName := fmt.Sprintf("%s.taco", outputPackage)
	outputZipPath := fmt.Sprintf("%s/%s", buildPath, packageZipName)
	buildFolder := fmt.Sprintf("%s/%s/", buildPath, outputPackage)
	preInstallScript(outputZipPath)

	maps.SetValidation(validateFile)
	categories.SetValidation(validateFile)

	os.RemoveAll(buildPath)
	os.Mkdir(buildPath, fs.ModePerm)

	trailbuilder.CompileResources(srcDirectory)
	packageCatagories, warnings, err := categories.Compile(fmt.Sprintf("%s/categories", srcDirectory))
	if err != nil {
		log.Println(err)
		return
	}
	for _, w := range warnings {
		log.Println(w)
	}
	packageMaps, warnings := maps.Compile(packageCatagories, fmt.Sprintf("%s/maps", srcDirectory))
	for _, w := range warnings {
		log.Println(w)
	}

	CopyAssets(fmt.Sprintf("%s/assets", srcDirectory), fmt.Sprintf("%s/assets", buildFolder))
	files.Copy(fmt.Sprintf("%s/pack.lua", srcDirectory), fmt.Sprintf("%s/pack.lua", buildFolder))
	categories.Save(packageCatagories, buildFolder)
	maps.Save(packageMaps, buildFolder)
	err = makeZip(buildFolder, outputZipPath)
	if err != nil {
		panic(err)
	}

	installScript(outputZipPath)
}

func makeZip(path string, dstfile string) error {
	outFile, err := os.Create(dstfile)
	if err != nil {
		log.Println(err)
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
	v = utils.Trim(v)
	fname := fmt.Sprintf("%s/%s", srcDirectory, v)
	if _, err := os.Stat(fname); errors.Is(err, os.ErrNotExist) {
		return fmt.Sprintf("File %s not found", v)
	}
	return ""
}
