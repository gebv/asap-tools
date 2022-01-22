package clickup

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func ParseSyncPreferences(in io.Reader) (*SyncPreferences, error) {
	res := &SyncPreferences{}
	err := yaml.NewDecoder(in).Decode(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type SyncPreferences struct {
	MirrorTaskRules []MirrorTaskSpecification `yaml:"mirror_task_rules"`
}

func (s *SyncPreferences) AllUsedTeamIDs() []string {
	found := map[string]bool{}
	res := []string{}
	for idx := range s.MirrorTaskRules {
		for _, teamID := range s.MirrorTaskRules[idx].UsedTeamIDs() {
			if !found[teamID] {
				found[teamID] = true
				res = append(res, teamID)
			}
		}
	}
	return res
}

// example https://app.clickup.com/2431928/v/f/96471870/42552884
func folderIDFromURL(in string) string {
	parse, err := url.Parse(in)
	if err != nil {
		zap.L().Error("Failed to parse URL", zap.String("url", in), zap.Error(err))
		return ""
	}

	args := strings.Split(parse.Path, "/")
	if len(args) < 4 {
		zap.L().Warn("Failed to get folder ID (invalid format url?)", zap.String("url", in))
		return ""
	}

	if args[2] != "v" {
		zap.L().Warn("Failed valid URL (invalid format url?)", zap.String("url", in))
		return ""
	}

	if args[3] != "f" {
		zap.L().Warn("Failed valid URL (invalid format url?)", zap.String("url", in))
		return ""
	}

	folderID := args[4]
	if folderID == "" {
		zap.L().Warn("Empty folder ID (invalid format url?)", zap.String("url", in))
		return ""
	}

	if _, err := strconv.ParseInt(folderID, 10, 64); err != nil {
		zap.L().Warn("Invalid folder ID (invalid format url?)", zap.String("url", in), zap.Error(err))
		return ""
	}
	return folderID
}

// example https://app.clickup.com/2431928/v/li/174318787
func listIDFromURL(in string) string {
	parse, err := url.Parse(in)
	if err != nil {
		zap.L().Error("Failed to parse URL", zap.String("url", in), zap.Error(err))
		return ""
	}

	args := strings.Split(parse.Path, "/")
	if len(args) < 4 {
		zap.L().Warn("Failed to get list ID (invalid format url?)", zap.String("url", in))
		return ""
	}

	if args[2] != "v" {
		zap.L().Warn("Failed valid URL (invalid format url?)", zap.String("url", in))
		return ""
	}

	if args[3] != "li" {
		zap.L().Warn("Failed valid URL (invalid format url?)", zap.String("url", in))
		return ""
	}

	listID := args[4]
	if listID == "" {
		zap.L().Warn("Empty list ID (invalid format url?)", zap.String("url", in))
		return ""
	}

	if _, err := strconv.ParseInt(listID, 10, 64); err != nil {
		zap.L().Warn("Invalid list ID (invalid format url?)", zap.String("url", in), zap.Error(err))
		return ""
	}
	return listID
}

// https://app.clickup.com/2431928/v/f/96471870/42552884
// https://app.clickup.com/2431928/v/li/174386179
func teamIDFromURL(in string) string {
	parse, err := url.Parse(in)
	if err != nil {
		zap.L().Error("Failed to parse URL", zap.String("url", in), zap.Error(err))
		return ""
	}

	args := strings.Split(parse.Path, "/")
	if len(args) < 2 {
		zap.L().Warn("Failed to get team ID (invalid format url?)", zap.String("url", in))
		return ""
	}
	teamID := args[1]
	if teamID == "" {
		zap.L().Warn("Empty teamID ID (invalid format url?)", zap.String("url", in))
		return ""
	}

	if _, err := strconv.ParseInt(teamID, 10, 64); err != nil {
		zap.L().Warn("Invalid team ID (invalid format url?)", zap.String("url", in), zap.Error(err))
		return ""
	}
	return teamID
}
