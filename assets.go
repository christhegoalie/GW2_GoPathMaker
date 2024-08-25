package main

import fcopy "github.com/otiai10/copy"

func CopyAssets(srcDir string, dstDir string) {
	fcopy.Copy(srcDir, dstDir, fcopy.Options{})
}
