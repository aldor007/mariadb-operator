package mysql

import (
	"fmt"
	"strings"
)

// Escape escapes a string
func Escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		escape = 0
		switch sql[i] {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
		case '\n': /* Must be escaped for logs */
			escape = 'n'
		case '\r':
			escape = 'r'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"': /* Better safe than sorry */
			escape = '"'
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, sql[i])
		}
	}

	return string(dest)
}

func escapeID(id string) string {
	if id == "*" {
		return id
	}

	// don't allow using ` in id name
	id = strings.ReplaceAll(id, "`", "")

	return fmt.Sprintf("`%s`", id)
}
