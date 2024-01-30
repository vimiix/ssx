package ssx

var (
	// Reference: https://github.com/vimiix/ssx/issues/14
	reservedWords = []string{
		"l", "ls", "list",
		"t", "tag",
		"d", "del", "delete",
		"a", "add",
		"i", "info",
		"u", "update",
		"cp", "scp",
		"stats", "top", "share",
		"ssx",
	}
	reservedWordsMap = map[string]bool{}
)

func init() {
	for _, word := range reservedWords {
		reservedWordsMap[word] = true
	}
}

func isReservedWord(word string) bool {
	return reservedWordsMap[word]
}
