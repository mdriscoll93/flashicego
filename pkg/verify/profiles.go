package verify

import "fmt"

type Profile struct {
	Name          string
	EdgeBytes     int64
	RandomSamples int
	Full          bool
}

func ResolveProfile(name string) (Profile, error) {
	switch name {
	case "quick", "":
		return Profile{Name: "quick", EdgeBytes: 8 * 1024 * 1024, RandomSamples: 3}, nil
	case "thorough":
		return Profile{Name: "thorough", EdgeBytes: 32 * 1024 * 1024, RandomSamples: 10}, nil
	case "full":
		return Profile{Name: "full", Full: true}, nil
	default:
		return Profile{}, fmt.Errorf("unknown verify profile %q", name)
	}
}
