# Beat Saber BPM Changer

A tool for changing the BPM of a beatsaber track without changing the position of the boxes and walls relative to the music.  
This is particularly useful for making song in 6/8 and other non-standard time signatures.

![screenshot](./screenshot.png)

In the screenshot above, I am creating a beatmap for a song in 6/8 with a BPM of 120. Since EditSaber doesn't support 6/8 songs yet, I had to edit the song in 360 BPM (3x the true BPM). This works, but causes the boxes to come at you much faster in game than they should. Bpm-saber fixes that problem by converting the BPM back to the correct tempo and the adjusting all the boxes and walls back to their correct position within the song.

This tool only creates the `DIFFICULTY_LEVEL.json` files in the output folder. You will have to copy over the other files (info.json, song.ogg, cover.jpg, etc.) yourself.

## Installation
Simply download and run bpm-saber.exe from the [releases page](https://github.com/zevdg/bpm-saber/releases).  

_Linux and mac builds are also available on that page.  
The mac build is untested but should work._

## Description of sections

### input song info.json

This is the info.json inside the folder where you are editing the song.

### output folder

This is the folder that you want the BPM corrected version of the song to be saved.  
**WARNING: The contents of this folder will be overwritten!**

### input bpm

This is the current BPM of the track that needs to be adjusted. You can use the button to load the BPM from the input song info.json, or enter it directly.

### output bpm

This is the desired BPM of the output after correction. Generally it is some multiple of the input BPM.  
This value can be derived from the input BPM using the built-in calculator, loaded from the output folder (assuming it contains an info.json), or entered directly.

### built-in calculator

If you know what output BPM you want, you can totally ignore this section. It is only provided for convenience. see the "output bpm" section for more details.

## Related tools

Apparently someone had already made a python script that does basically the same thing but without a GUI.  
You can find it here https://discordapp.com/channels/441805394323439646/443569023951568906/459771392054001666
