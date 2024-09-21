package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gw2_markers_gen/blish"
	"gw2_markers_gen/files"
	"io/fs"
	"os"

	"github.com/google/uuid"
)

const janthir_files_local = "files1_syntri"
const janthir_files_remote = "files2_syntri"
const lowland_files_local = "files1_lowlands"
const lowland_files_remote = "files2_lowlands"
const remote_target_type = ""

type exec struct {
	src1Path    string
	src2Path    string
	diff1Output string
	diff2Output string
}

var ignore func() blish.PoiList

func main() {
	execs := []exec{
		{
			src1Path:    lowland_files_local,
			src2Path:    lowland_files_remote,
			diff1Output: fmt.Sprintf("missing_%s.json", lowland_files_local),
			diff2Output: fmt.Sprintf("missing_%s.xml", lowland_files_remote),
		},
		{
			src1Path:    janthir_files_local,
			src2Path:    janthir_files_remote,
			diff1Output: fmt.Sprintf("missing_%s.json", janthir_files_local),
			diff2Output: fmt.Sprintf("missing_%s.xml", janthir_files_remote),
		},
	}
	for _, ex := range execs {
		os.Remove(ex.diff1Output)
		os.Remove(ex.diff2Output)

		points1 := files.ReadAllPoints(ex.src1Path)
		points2 := files.ReadAllPoints(ex.src2Path)

		diff1 := calcDiff(points1, points2, ignore())
		diff2 := calcDiff(points2, points1, ignore())

		if len(diff1) > 0 {
			b, _ := json.MarshalIndent(diff1, "", "  ")
			os.WriteFile(ex.diff1Output, b, fs.ModePerm)
		}
		if len(diff2) > 0 {
			for i, p := range diff2 {
				p.Behavior = points2[0].Behavior
				p.MapID = points2[0].MapID
				p.Type = remote_target_type
				p.GUID = newUUID()
				diff2[i] = p
			}
			b, _ := xml.MarshalIndent(diff2, "", "  ")
			os.WriteFile(ex.diff2Output, b, fs.ModePerm)
		}
	}
}

func newUUID() string {
	uuid := uuid.New()
	return base64.StdEncoding.EncodeToString(uuid[:])
}

func calcDiff(ls1 blish.PoiList, ls2 blish.PoiList, ignore blish.PoiList) blish.PoiList {
	out := make(blish.PoiList, 0)
	for _, p := range ls2 {
		if ignore != nil && ignore.Contains(p.Point()) {
			continue
		}
		if !ls1.Contains(p.Point()) {
			out = append(out, p)
		}
	}
	return out
}
