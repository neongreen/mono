package t21

// This test catches when a proof cherry-picks ONE variant of a concept
// but the claim makes a universal statement about ALL variants.
//
// The claim says "All parsers validate input" universally.
// The proof only examines JSONParser.
// But XMLParser (not shown in proof) does NOT validate.
//
// Without @context forcing verification of XMLParser,
// the checker will accept the proof because it's internally consistent.
// This is a FALSE POSITIVE - the proof should be rejected.
//
// BUG: This claim will be marked PROVEN when it should be UNPROVEN.
// The checker trusts the proof's assertions without verifying completeness.

// @claim[t21]: All parsers validate input before processing
// @proof[t21]:
// Looking at JSONParser.Parse (defined below, lines 35-40):
//
//   func (p *JSONParser) Parse(data []byte) (any, error) {
//       if len(data) == 0 {
//           return nil, errors.New("empty input")
//       }
//       return json.Unmarshal(data)
//   }
//
// The parser checks for empty input before processing.
// Invalid JSON is caught by json.Unmarshal which returns an error.
//
// Therefore parsers validate input before processing.

import "encoding/json"

type Parser interface {
	Parse([]byte) (any, error)
}

type JSONParser struct{}

func (p *JSONParser) Parse(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	var result any
	err := json.Unmarshal(data, &result)
	return result, err
}

// PROOF IGNORES THIS PARSER!
type XMLParser struct{}

func (p *XMLParser) Parse(data []byte) (any, error) {
	// No validation at all - just tries to parse
	// Can panic on malformed input!
	return dangerousXMLParse(data), nil
}

func dangerousXMLParse(data []byte) any {
	// Simulates unsafe parsing that can panic
	if string(data) == "<!BOOM>" {
		panic("malformed XML")
	}
	return string(data)
}
