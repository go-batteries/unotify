package helpers

import "regexp"

func ToProjectID(issueID string) string {
	projectID := issueID

	re := regexp.MustCompile(`^([\w\d]+)-\d{1,}`)
	matches := re.FindStringSubmatch(issueID)
	if len(matches) == 2 {
		projectID = matches[1]
	}

	return projectID
}
