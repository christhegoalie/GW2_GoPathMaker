# File structure based TACO pathing generation tool for Guild Wars 2.

## Download the latest Marker Pack
[Download Latest Marker Pack](https://github.com/christhegoalie/GW2_GoPathMaker/releases/latest/download/ShellshotMarkerPack.taco)

## Features
- Janthir Lowlands
1. Warclaw Caches
1. Buzzy Treetops Jumping Puzzle
1. Vale Brazier Jumping Puzzle
1. Hidden Achievments
- Janthir Syntri
1. Warclaw Caches
1. Major Kodan Caches
1. Charted Titan Ore Gather Nodes
1. Rotten Titan Amber Gather Nodes

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

## Appendix
### Directory Structure
#### `maps` directory
- Location for storing map information
- Every directory under the maps directory MUST definte a map [see description below]
#### `map` directory
- No Directory structure is required
- The root or subdirectory MAY contain any number of [.poi`](#.poi-file-extension) files
- The root or subdirectory MAY contain any number of [.trail](#.trail-file-extension) files
- MUST contain a `mapinfo.txt` file defining the mapid.
- MAY contain a `barriers.txt` file for trail creation. (defines regions path generation is unable to cross)
- MAY contain a `paths.txt` file for trail creation (defining a list of paths/shortcuts. Typically these are bouncing mushrooms or ways to bypass barriers)
- MAY contain a `waypoints.txt` file for trail creation (used to generate starting location)
#### `categories` directory
- Location for storing category definitions
- Directory structure determines category name.
- EX: `categories/Janthir/Chests/MajorCaches.cat` generates the Category: `Janthir.Chests.MajorCaches`
- Display Names will be generated from directory names (spaces will be added When casing alternates)
#### `assets` directory
- No Directory structure is required
- General location for storing assets (images/binary trail data)
- No verification takes place
- Compiled assets are placed in this directory [See compiled_assets for more information]
#### `compiled_assets` directory
- No Directory structure is required
- Location for storing trail definition files that will be compiled to `.trl` files
- Assets compiled from this directory will be stored in the `assets` directory
- EX: `compiled_assets/mytrails/trail1.rtrl` will be compiled to `assets/mytrails/trail1.trl`
#### User directories
- All files/directories inside your marker pack root directory will be zipped into the output path file.
- This can be used to add any custom data required

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
### .atrl
- Every line defines a key/value pair describing map information
- Key/Value MUST be separated by the `=` sign
- The file MUST contain the `map` key
- the file MUST contain a `file` key
- All Other Keys are ignored
- Lines without position information are skipped
- The `map` value MUST match the name of a directory in your `maps` folder
- The `file` value MUST be a valid path relative to the map directory defined in the `map` field

