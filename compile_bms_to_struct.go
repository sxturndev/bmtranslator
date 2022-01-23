package main

import (
	"bufio"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	titleRegex = regexp.MustCompile("\\(([^)]*)\\)|-([^-]*)-|\\[([^]]*)]|'([^']*)'|\"([^\"]*)\"|~([^~]*)~")
)

// CompileBMSToStruct converts a BMS file into a struct (BMSFileData) which can then be interpreted by the rest
// of the program. It does not do any position calculation, only makes the data readable.
func (conf *ProgramConfig) CompileBMSToStruct(inputPath string, bmsFileName string) (*BMSFileData, error) {
	file, err := os.Open(path.Join(inputPath, bmsFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	fileData := &BMSFileData{
		Metadata: BMSMetadata{
			Title:      "No title",
			Artist:     "Unknown artist",
			Difficulty: "Unnamed Difficulty",
		},
		TrackLines: map[int][]Line{},
		HitObjects: map[int][]HitObject{},
		Indices: IndexData{
			BPMChanges: map[string]float64{},
			Stops:      map[string]float64{},
			BGA:        map[string]string{},
		},
		Audio: AudioData{
			StringArray:      make([]string, 0),
			HexadecimalArray: make([]string, 0),
		},
		SoundEffects: make([]SoundEffect, 0),
		TimingPoints: map[float64]float64{},
		StartingBPM:  DefaultStartingBPM,
	}

	// Should be true if the value of #IF n is anything other than 2. Resets at the #END(IF) mark.
	// Prevents this line from being read and immediately skip to the next one.
	ignoreLine := false

	lineIndex := 0
	for scanner.Scan() {
		lineIndex++
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			continue
		}
		lineLower := strings.ToLower(line)

		// Reached an #END(IF) header, we can continue parsing.
		if strings.HasPrefix(lineLower, "#end") && ignoreLine {
			ignoreLine = false
			continue
		}
		if strings.HasPrefix(lineLower, "#if") && !ignoreLine {
			if len(line) < 5 {
				// Invalid #IF, should ignore just to be safe
				ignoreLine = true
				continue
			}
			if line[4] != '1' {
				// #IF is not 1, ignore it
				ignoreLine = true
			}
			continue
		}
		if ignoreLine {
			continue
		}

		// If this is true, the line is a header.
		if len(line) < 7 || line[6] != ':' {
			if strings.HasPrefix(lineLower, "#player") {
				if len(line) < 9 {
					if conf.Verbose {
						color.HiRed("* Player type cannot be determined. (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				switch line[8] {
				case '1':
					break
				case '2':
					if conf.Verbose {
						color.HiYellow("* Map specified #PLAYER 2; skipping")
					}
					return nil, nil
				case '3':
					if conf.Verbose {
						color.HiYellow("* Double play mode; skipping")
					}
					return nil, nil
				default:
					if conf.Verbose {
						color.HiYellow("* Even though player header was defined, there was no valid input (Line: %d)", lineIndex)
					}
					return nil, nil
				}
			} else if strings.HasPrefix(lineLower, "#genre") {
				if len(line) < 8 {
					if conf.Verbose {
						color.HiYellow("* #genre is invalid, ignoring (Line: %d)", lineIndex)
					}
					fileData.Metadata.Tags = "BMS"
					continue
				}
				lineBytes := []byte(line[7:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if conf.Verbose {
						color.HiYellow("* #genre couldn't be converted via ShiftJIS (Line: %d)", lineIndex)
					}
				}
				fileData.Metadata.Tags = strings.Replace(b, "'", "\\'", -1)
			} else if strings.HasPrefix(lineLower, "#subtitle") {
				if len(line) < 11 {
					if conf.Verbose {
						color.HiYellow("* #subtitle is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[10:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if conf.Verbose {
						color.HiYellow("* #subtitle couldn't be converted via ShiftJIS (Line: %d)", lineIndex)
					}
				}
				fileData.Metadata.Subtitle = strings.Replace(b, "'", "\\'", -1)
			} else if strings.HasPrefix(lineLower, "#subartist") {
				if len(line) < 12 {
					if conf.Verbose {
						color.HiYellow("* #subartist is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[11:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if conf.Verbose {
						color.HiYellow("* #subartist couldn't be converted via ShiftJIS (Line: %d)", lineIndex)
					}
				}
				fileData.Metadata.SubArtists = append(fileData.Metadata.SubArtists, strings.Replace(b, "'", "\\'", -1))
			} else if strings.HasPrefix(lineLower, "#title") {
				if len(line) < 8 {
					if conf.Verbose {
						color.HiYellow("* #title is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[7:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if conf.Verbose {
						color.HiYellow("* #title couldn't be converted via ShiftJIS (Line: %d)", lineIndex)
					}
				}
				fileData.Metadata.Title = strings.Replace(b, "'", "\\'", -1)
			} else if strings.HasPrefix(lineLower, "#lnobj") {
				if len(line) < 8 {
					if conf.Verbose {
						color.HiRed("* #lnobj is not a valid length (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				if len(lineLower[7:]) != 2 {
					if conf.Verbose {
						color.HiRed("* #lnobj was specified, but not 2 bytes in length. (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				fileData.LNObject = lineLower[7:]
			} else if strings.HasPrefix(lineLower, "#artist") {
				if len(line) < 9 {
					if conf.Verbose {
						color.HiYellow("* #artist is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[8:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if conf.Verbose {
						color.HiYellow("* #artist couldn't be converted via ShiftJIS (Line: %d)", lineIndex)
					}
				}
				fileData.Metadata.Artist = strings.Replace(b, "'", "\\'", -1)
			} else if strings.HasPrefix(lineLower, "#playlevel") {
				if len(line) < 12 {
					if conf.Verbose {
						color.HiYellow("* #playlevel is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				fileData.Metadata.Difficulty = line[11:]
			} else if strings.HasPrefix(lineLower, "#stagefile") {
				if len(line) < 12 {
					if conf.Verbose {
						color.HiYellow("* #stagefile is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				exists := FileExists(path.Join(inputPath, line[11:]))
				if !exists {
					color.HiYellow("* \"%s\" (#stagefile) wasn't found; ignoring (Line: %d)", line[7:], lineIndex)
				}
				fileData.Metadata.StageFile = line[11:]
			} else if strings.HasPrefix(lineLower, "#banner") {
				if len(line) < 9 {
					if conf.Verbose {
						color.HiYellow("* #banner is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				exists := FileExists(path.Join(inputPath, line[8:]))
				if !exists {
					color.HiYellow("* \"%s\" (#banner) wasn't found; ignoring (Line: %d)", line[8:], lineIndex)
				}

				fileData.Metadata.Banner = line[11:]
			} else if strings.HasPrefix(lineLower, "#bpm ") {
				if len(line) < 6 {
					if conf.Verbose {
						color.HiRed("* #bpm XX is invalid, parsing cannot continue (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				i, e := strconv.ParseFloat(line[5:], 64)
				if e != nil {
					if conf.Verbose {
						color.HiRed("* #bpm XX is not a number, parsing cannot continue (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				// Here the initial BPM is set. Also set the timing point.
				fileData.StartingBPM = i
			} else if strings.HasPrefix(lineLower, "#bpm") {
				if len(line) < 8 {
					color.HiYellow("* BPM change invalid. will be ignored (Line: %d)", lineIndex)
					continue
				}
				i, e := strconv.ParseFloat(line[7:], 64)
				if e != nil {
					color.HiYellow("* BPM change is not a number. will be ignored (Line: %d)", lineIndex)
					continue
				}
				fileData.Indices.BPMChanges[lineLower[4:6]] = i
			} else if strings.HasPrefix(lineLower, "#bmp") {
				if len(line) < 8 {
					color.HiYellow("* BMP invalid, ignoring (Line: %d)", lineIndex)
					continue
				}
				exists := FileExists(path.Join(inputPath, line[7:]))
				if !exists {
					color.HiYellow("* \"%s\" wasn't found; ignoring (Line: %d)", line[7:], lineIndex)
					continue
				}
				fileData.Indices.BGA[lineLower[4:6]] = line[7:]
			} else if strings.HasPrefix(lineLower, "#stop") {
				if len(line) < 9 {
					color.HiYellow("* STOP isn't correctly formatted, not going to use it (Line: %d)", lineIndex)
					continue
				}
				i, e := strconv.ParseFloat(line[8:], 64)
				if e != nil {
					color.HiYellow("* STOP is not a valid number, not going to use it (Line: %d)", lineIndex)
					continue
				}
				if i < 0.0 {
					color.HiYellow("* STOP is negative (< 0.0), not going to use it (Line: %d)", lineIndex)
					continue
				}
				fileData.Indices.Stops[lineLower[5:7]] = i
			} else if strings.HasPrefix(lineLower, "#wav") {
				if len(line) < 8 {
					color.HiYellow("* WAV command invalid, all notes/sfx associated with it won't be placed (Line: %d)", lineIndex)
					continue
				}

				// Correct the extension used. E.g. if it's #WAV FILE.mp3 but actually FILE.wav, this will fix that.
				// Most BMS players ignore the extension. However, I am not entirely sure if Quaver/osu also behave the same.
				// Just to be safe and to future-proof, we search the filesystem for the right file.
				// TODO: I think that this puts unnecessary strain on the filesystem. I'll look into a better way someday.
				soundEffect := SearchForSoundFile(inputPath, line[7:])
				if len(soundEffect) == 0 {
					color.HiYellow("* (#WAV) \"%s\" wasn't found or isn't either .wav, .mp3, .ogg, or .3gp. ignoring (Line: %d)", line[7:], lineIndex)
					continue
				}
				fileData.Audio.StringArray = append(fileData.Audio.StringArray, soundEffect)
				fileData.Audio.HexadecimalArray = append(fileData.Audio.HexadecimalArray, lineLower[4:6])
			}
			continue
		}

		tInt, e := strconv.ParseInt(line[1:4], 10, 64)
		if e != nil {
			color.HiRed("* Failed to parse track #, cannot continue parsing (Line: %d, Content: %s)", lineIndex, line)
			return nil, nil
		}
		channel := lineLower[4:6]

		// TODO: remove this; we just ignore mines now lmao
		//if len(mineRegex.FindString(channel)) > 0 {
		//	color.HiYellow("* Cannot parse maps with mines/fakes due to often being coupled with per-column SV, which neither quaver or osu support (Line: %d)", lineIndex)
		//	return nil, nil
		//}
		thisLineData := Line{
			Channel: channel,
		}
		if len(line) > 7 {
			thisLineData.Message = strings.ToLower(line[7:])
		}
		fileData.TrackLines[int(tInt)] = append(fileData.TrackLines[int(tInt)], thisLineData)
		lineIndex++
	}

	// Append missing measures to TrackLines with blank BGM
	for index := 0; index < len(fileData.TrackLines); index++ {
		_, exists := fileData.TrackLines[index]
		if !exists {
			blankMeasure := Line{
				Channel: "01",
				Message: "00",
			}
			fileData.TrackLines[index] = append(fileData.TrackLines[index], blankMeasure)
		}
	}

	// Subtitle wasn't found and user doesn't want to keep implicit subtitles intact. Try
	// to find the subtitle in the title.
	if !conf.KeepSubtitles && len(fileData.Metadata.Subtitle) == 0 {
		rs := titleRegex.FindAllString(fileData.Metadata.Title, -1)
		if len(rs) > 0 {
			fileData.Metadata.Subtitle = rs[len(rs)-1]
			fileData.Metadata.Title = strings.Replace(fileData.Metadata.Title, rs[len(rs)-1], "", -1)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return fileData, nil
}
