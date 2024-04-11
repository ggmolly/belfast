package misc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type Field struct {
	Label string
	Name  string
	Type  string
	Index int
}

var (
	defaultErrFields = []Field{
		{
			Name: "No fields found",
			Type: "",
		},
	}

	fieldRegex = regexp.MustCompile(`(required|repeated|optional)\s(\w+)\s(\w+)\s\=\s(\d+);`)
)

func parseFile(path string) []Field {
	var output []Field
	file, err := os.Open(path)
	if err != nil {
		return defaultErrFields
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// check if the line matches the regex
		if fieldRegex.MatchString(scanner.Text()) {
			// parse the line
			matches := fieldRegex.FindStringSubmatch(scanner.Text())
			fieldIndex, err := strconv.Atoi(matches[4])
			if err != nil {
				fieldIndex = -1
			}
			output = append(output, Field{
				Label: matches[1],
				Type:  matches[2],
				Name:  matches[3],
				Index: fieldIndex,
			})
		}
	}
	return output
}

func GetPacketFields(packetId int) []Field {
	globResults, _ := filepath.Glob("packets/protobuf_src/" + fmt.Sprintf("*_%d.proto", packetId))
	if len(globResults) == 0 {
		return defaultErrFields
	}
	return parseFile(globResults[0])
}
