This directory encapsulates the files used to edit and build the Tabletop Simulator integration.
- `/extensions` - VSCode extensions
- `/src` - Standalone TTS Lua files
- `TS_Save_2.json` - Full, self-contained save artifact for the mod
- `TS_Save_2.png` - Preview image for loading this save

The JSON file "is" the mod. It is an entirely self-contained serialization of the scene and its scripts. You can edit this file in 2 ways: In Tabletop Simulator, or in VSCode. Each has its place, for editing the Lua you are definitely going to want to use VSCode.


## Setting up VSCode

Download VSCode [here](https://code.visualstudio.com/download?_exp_download=fb315fc982).

Once installed, I recommend creating a profile specifically for modding TTS. Open VSCode and go to the Profiles page (via Command Palette or `File > Preferences > Profile`), create a profile for TTS-Lua, and set it to Active.

Here's a couple extensions you can find on the marketplace that I recommend:
1. Lua (sumneko.lua) - Provides a language server for Lua
2. TODO Highlight (wayou.vscode-todo-highlight) - Highlights keywords like "Todo", "FixMe" (I use "Note" often, you'll have to add that keyword yourself in the settings)


## Installing the TabletopSimulator-Lua-VSCode

Rolandostar made a VSCode extension that provides several features conducive to mod development.

His repository uses the MIT license, so I went ahead and included a copy of the extension here. If you want to build it yourself, go to the [Building the extension from source](#building-the-extension-from-source) section.

In VSCode, open the Command Palette and select `Install from VSIX...`. Navigate to the `tabletopsimulator-lua-2.0.0-rc1.vsix` file and select it.

Once the extension is loaded, go to `File > Preferences > Settings`, navigate to the `Extensions` section, then to the `Tabletop Simulator Lua` section, and in the Miscellaneous section check the boxes for `Debug Search Paths` and `Disable Directory Warning`.

## Using TabltopSimulator-Lua-VSCode

[Documentation](https://tts-vscode.rolandostar.com/) - [Github](https://github.com/rolandostar/tabletopsimulator-lua-vscode)

First, copy the `TS_Save_2.*` files to your Tabletop simulator save directory:
- Windows: `~\Documents\My Games\Tabletop Simulator\Saves`
- Linux: `~/.local/share/Tabletop Simulator/Saves`

Open Tabletop Simulator and load your save file. Go to VSCode and press Ctrl+Shift+L. It will prompt you if you want to load files from Tabletop Simulator--say yes.

Warning: The accuracy of these next statements depends on if you are using the correct version of the extension.

You'll get a new section in your Explorer for `Tabletop Simulator Files` containing directories for the "objects" in the scene and files for the scripts. From what I can tell, you can't edit the properties (position, description, etc.) of objects here--that still must be done in game. The scripts you can edit and the changes will be reflected in game upon pressing Ctrl+Shift+S.

## Workflow
When you make changes to a script, you must reload the scene to see your changes in effect by pressing Ctrl+Shift+S.

These changes are not saved to the TS save file by doing this. To save these changes to the file, you must create/overwrite a save in Tabletop Simulator in the "Games" menu.


## Building the extension from source
While you can install the extension via the Extensions tab in VSCode (rolandostar.tabletopsimulator-lua) or from the [release page on Github](https://github.com/rolandostar/tabletopsimulator-lua-vscode/releases/tag/1.1.3), I strongly recommend you build the extension from source: The release version is over 5 years old, the main branch has been developed well past that of the official release and contains the additional features that are described in the documentation.

To build the extension from the source, you will need npm, webpack and the webpack-cli, and VSCE. Here's the step-by-step I used:
1. Clone the repository (https://github.com/rolandostar/tabletopsimulator-lua-vscode)
2. [Install npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)
2. Open your CLI to the root of the repository
3. Install the required dependencies:
    - `npm i -g webpack`
    - `npm i -g @vscode/vsce`
    - `npm i -g webpack-cli`
4. Compile the project (you may be prompted to install webpack-cli if you haven't done so already)
    - `npm run compile`
5. Compile the extension to .vsix
    - `vsce package`

The final command outputs a file that looks like `tabletopsimulator-lua-2.0.0-rc1.vsix`.

Notes: The `dev` branch is 29 commits ahead of main, was last active 2 years ago, and describes more features. I tried compiling this version but was hit with `npm error 402 Payment Required - GET https://gitpkg.vercel.app/rolandostar/tts-tools/packages/savefile?dev`. If I revisit this and figure out a workaround, I'll update this section.

