# Be-Music Translator (BMT)

Converts your BMS levels to a modern file format (osu, qua or json), ready to import & play.

This fork is used for my BMX2OSU converter. See it [here](https://github.com/sxturndev/bmx2osu).

## How to Use

Output to Quaver: `bmt.exe -i /path/to/input -o /path/to/output`

Output to osu: `bmt.exe -i /path/to/input -o /path/to/output -type osu`

Get JSON info only: `bmt.exe -i /path/to/input -o /path/to/output -json-only` (see last section for details in info included)

## Options

| Option | Arguments? | Optional? | Description  | Default |
| ------------ | ---- | --- | ---------- | ---- |
|  `-i` | Yes | **No** | Path of input folder containing folders of BMS charts. | N/A |
|  `-o` | Yes | **No** | Path to output the converted files to. | N/A |
|  `-vol` | Yes | Yes | Volume of hit sounds. (0-100) | 100 |
|  `-type` | Yes | Yes | Which type of file to convert to. You can choose `quaver` or `osu`. | quaver |
|  `-hp` | Yes | Yes | **osu! only.** Specify the HP drain rate. (0.0-10.0) | 8.5 |
|  `-od` | Yes | Yes | **osu! only.** Specify the overall difficulty. (0.0-10.0) | 8.0 |
|  `-v` | No | Yes | If this is specified, all logs (including some debug information) will be shown. Useful if you want to know why some maps didn't convert. | N/A |
|  `-auto-scratch` | No | Yes | If this is specified, all notes in the scratch lane will be replaced with sound effects instead, and the scratch lane will not be shown in all clients.
|  `-keep-subtitles` | No | Yes | If this is specified, [implicit subtitles](https://hitkey.nekokan.dyndns.info/cmds.htm#TITLE-IMPLICIT-SUBTITLE) will **not** be removed from song titles. | N/A |
|  `-no-storyboard` | No | Yes | **osu! only.** If this is specified, background animation frames won't be parsed or inserted into the output files. | N/A |
|  `-no-measure-lines` | No | Yes | If this is specified, timing points will **not** be added at the end of each track to create visible measure lines. (It's a cosmetic thing and doesn't affect gameplay, but it might make slowjam unreadable. Some BMS files' notes will appear unsnapped if this is enabled.) | N/A |
|  `-no-timing-points` | No | Yes | If this is specified, **no** timing points will be added to the output file. This means no SV changes and is useful for SV maps which don't convert correctly. | N/A |
|  `-json` | No | Yes | In addition to the output, an accompanying .json file will be created for each chart, with information about the file (start times, metadata, etc). These will be placed in the same output folder. | N/A |
|  `-json-only` | No | Yes | When specified, no zips will be created, only .json files. `-json` becomes irrelevant if you enable this. | N/A |

## Limitations

- `#IF 1` will always be used. All other `#IF` blocks will be ignored. If an `#IF N` block where `n` is not 1 is never terminated with `#ENDIF`, the rest of the track will not be read.
- `#SWITCH` blocks can't be parsed (yet). If people want this i'll add it in.
- Almost no BMS maps use long notes in channels `51-59`, they use `#LNOBJ`. As a result, LNs placed in channels `51-59` are **untested**, but they are implemented. If you find a problem with them, please open an issue.
- BMS maps that use images as frames for the Background Animation can't be reliably parsed if the frames are <1ms apart, since osu! requires truncation of the decimal.
- If a BPM change occurs at any point within a STOP command, BMTranslator will still be able to parse the map, but the timing of the rest of the song will most likely be fucked. *However*, this has not appeared in a single map that I've tested, and by this reasoning, I think the only way to do this is by editing a BMS file by hand.

## Understanding the JSON output

- The `hit_objects` field has an array of hit objects, ordered by the lane they appear in (index 0 = lane 1...index 7 = lane 8)
- Each `key_sounds` field contains a `sample` value with an integer. If there is no key sound on that note, the `sample` value will always be 0. The `sample` value, minus 1, is the index of the file name that should play in reference to `sample_index`. For example, if the `sample` is `3`, you would look at `sound_effect_index[2]` to figure out which file to play.
- `end_time` values in hit objects will always be `0` unless `is_long_note` is `true`.

## Credits

This tool wouldn't be possible without help from others. Thank you to:

- [Swan](https://github.com/Swan) for helping me with .qua files
- [yahweh](https://osu.ppy.sh/users/10465260) for moral support and for being a cool friend (and everyone in arrowsmashers)
- [mat](https://osu.ppy.sh/users/6283029) for helping me with the math side of things
- [Jakads](https://osu.ppy.sh/users/259972) for helping with 0 BPM change in osu
- [hitkey](https://hitkey.nekokan.dyndns.info) for their very helpful [BMS notes/documentation](https://hitkey.nekokan.dyndns.info/cmds.htm). It is very thorough and an excellent reference for anyone looking to work with BMS and similar formats.
- Various people on the osu! forums who explained how timing points in osu! work.

## License

MIT License. See the LICENSE file for more information.