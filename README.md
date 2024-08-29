# File structure based TACO pathing generation tool for Guild Wars 2.

## Download the latest Marker Pack
[Download Latest Marker Pack](https://github.com/christhegoalie/GW2_GoPathMaker/releases/latest/download/ShellshotMarkerPack.taco)

# Building

## Requires GO Version 1.21 or higher
1. https://go.dev/doc/install

# Building the default marker pack from scratch
1. `go build`
1. `./gw2_markers_gen -n ShellshotMarkerPack`
1. copy build/ShellshotMarkerPack.zip to your blish/taco marker pack directory (Typically `C:\Users\{user}\Documents\Guild Wars 2\addons\blishhud\markers`)

## Build your own marker pack
1. Build the package generator: `go build`
1. Create a marker pack directory `XXXMarkerPack`
1. Add your catagories directory. `XXXMarkerPack/categories`
1. Define categories using your directory structure. Example: `XXXMarkerPack/categories/Janthir/Chests` generates the Category: `Janthir.Chests`
1. Any edge category requiring configuration (including icons), may be defined using a `.cat` file instead
1. Add your maps directory. `XXXMarkerPack/maps`
1. Add a map you intend to add markers for: `XXXMarkerPack/maps/JanthirSyntri`
1. Create `mapinfo.txt` in your map directory containing the map id. EX: `id=1554` (Can be easily found using the "Marker Pack Assistant" module from blish)
1. Create any number of `.poi` and `.trail` files containing marker location information. (any sub directory structure may be used)
1. Generate your package zip file: `./gw2_markers_gen -n XXXMarkerPack`

### .cat file format
- Every line defines a key/value pair describing category attributes. (See `https://www.gw2taco.com/2016/01/how-to-create-your-own-marker-pack.html` for a list of valid attributes)
- Key/Value MUST be separated by the `=` sign
- All non Key/Value pair lines will be skipped
### .poi file format
- Line 1 MUST reference a marker category present in your category directory. EX: `category=ShellshotMarkerPack.Janthir.GatherNodes.ChargedOre`
- Every subsequent line references a single marker
- Every marker line is defined as a list of Key/Value pairs
- Value pairs MUST be seperated by the space character
- Key/Value MUST be separated by the `=` sign
- Every marker line MUST contain X,Y,Z position information (as copied using the "Marker Pack Assistant" module from blish)
- Every marker line MAY overwrite marker attributes
- Example Line: `xpos="-290.0943" ypos="32.79265" zpos="-283.0596" Behavior="0"`
### .trail file format
- Line 1 MUST reference a marker category present in your category directory. EX: `category=ShellshotMarkerPack.Janthir.GatherNodes.ChargedOre`
- Every subsequent line references a single marker
- Every marker line is defined as a list of Key/Value pairs
- Value pairs MUST be seperated by the space character
- Key/Value MUST be separated by the `=` sign
- Every marker line MUST contain the `trailData` key pointing to a `.trl` file. (See `https://www.gw2taco.com/2016/01/how-to-create-your-own-marker-pack.html` for trail creation)
- Every marker line MAY overwrite marker attributes
- Example Line: `trailData="assets/trails/janthir_lowlands/honeybey_jp.trl" color="ffffffff"`
### mapinfo.txt format
- Every line defines a key/value pair describing map information
- Key/Value MUST be separated by the `=` sign
- The file MUST contain the `id` key
- All other information in the file will be skipped
### .rtrl file format
- All Lines MUST be a list of Key/Value Pairs seperated by the space character
- Key/Values MUST be seperated by the `=` sign
- Line 1 MUST contain the `mapid` key (and other keys will be ignored)
- Subsequent lines MUST contain X,Y,Z position information (as copied using the "Marker Pack Assistant" module from blish)
- All Other Keys are ignored
- Lines without position information are skipped
