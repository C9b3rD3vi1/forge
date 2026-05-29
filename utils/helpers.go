package utils

import (
	"math"
	"strings"
	"github.com/google/uuid"
    "encoding/json"
    "github.com/C9b3rD3vi1/forge/models"
    
)

func ComputeReadingTime(content string) int {
	words := len(strings.Fields(content))
	if words == 0 {
		return 1
	}
	minutes := int(math.Ceil(float64(words) / 200.0))
	if minutes < 1 {
		return 1
	}
	return minutes
}

// Template helper functions
func HasTechStack(project *models.Projects, techStackID uuid.UUID) bool {
    for _, ts := range project.TechStacks {
        if ts.ID == techStackID {
            return true
        }
    }
    return false
}

// Template function to parse JSON (for gallery display)
func ParseJSON(s string) []string {
    var result []string
    if err := json.Unmarshal([]byte(s), &result); err != nil {
        return []string{}
    }
    return result
}

func SplitString(s string, sep string) []string {
    return strings.Split(s, sep)
}

func Add(a int, b int) int {
    return a + b
}


func Trim(s string) string {
    return strings.Trim(s, " ")
}


func Seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}

func ColorClass(i int) string {
	classes := []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8"}
	return classes[i%8]
}
