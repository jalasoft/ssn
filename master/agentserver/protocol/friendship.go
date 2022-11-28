package protocol

import (
	"fmt"
	"strings"
)

type HelloMessage struct {
	Name  string
	Props map[string]string
}

type SkillMessage struct {
	Name  string
	Type  string
	Props map[string]string
}

func IsThatsAllsMessage(b []byte) bool {
	message := string(b)

	return strings.HasPrefix(message, "[THATSALL]")
}

func IsIamStillHereMessage(b []byte) bool {
	message := string(b)

	return strings.HasPrefix(message, "[IAMSTILLHERE]")
}

func IsByeMessage(b []byte) bool {
	message := string(b)
	return strings.HasPrefix(message, "[BYE]")
}

func ParseHelloMessage(b []byte) (HelloMessage, error) {
	message := string(b)

	body, err := removeBoundaries(message, "[HELLO;")

	if err != nil {
		return HelloMessage{}, err
	}

	segments, err := parseSegments(message, body)

	if err != nil {
		return HelloMessage{}, err
	}

	name, ok := segments["name"]

	if !ok {
		return HelloMessage{}, fmt.Errorf("Message '%s' does not provide 'name'", message)
	}

	delete(segments, "name")

	return HelloMessage{
		name,
		segments,
	}, nil
}

func ParseSkillMessage(b []byte) (SkillMessage, error) {
	message := string(b)

	body, err := removeBoundaries(message, "[SKILL;")

	if err != nil {
		return SkillMessage{}, err
	}

	segments, err := parseSegments(message, body)

	if err != nil {
		return SkillMessage{}, err
	}

	name, ok := segments["name"]

	if !ok {
		return SkillMessage{}, fmt.Errorf("Message '%s' does not provide 'name'", message)
	}

	t, ok := segments["type"]

	if !ok {
		return SkillMessage{}, fmt.Errorf("Message '%s' does not provide 'type'", message)
	}

	delete(segments, "name")
	delete(segments, "type")

	return SkillMessage{
		name,
		t,
		segments,
	}, nil
}

func parseSegments(message string, segmentsStr string) (map[string]string, error) {
	segments := strings.Split(segmentsStr, ";")

	if len(segments) == 0 {
		return nil, fmt.Errorf("No content in message '%s'", message)
	}

	result := make(map[string]string)

	for _, segment := range segments {
		keyValue := strings.Split(segment, "=")

		if len(keyValue) != 2 {
			return nil, fmt.Errorf("Content of message '%s' has incompatible pair '%s'", message, segment)
		}

		if _, ok := result[keyValue[0]]; ok {
			return nil, fmt.Errorf("Message '%s' contains multiple key-value pairs with the same key '%s'", message, keyValue[0])
		}

		result[keyValue[0]] = keyValue[1]
	}

	return result, nil
}

func removeBoundaries(message string, prefix string) (string, error) {
	if !strings.HasPrefix(message, prefix) {
		return "", fmt.Errorf("Message '%s' does not have expected prefix '%s'", message, prefix)
	}

	endBrackedIndex := strings.Index(message, "]")

	if endBrackedIndex < 0 {
		return "", fmt.Errorf("Message '%s' does not ends with ']'", message)
	}

	result := strings.TrimLeft(message[:endBrackedIndex], prefix)
	return result, nil
}
