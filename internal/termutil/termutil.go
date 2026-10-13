package termutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

func IsTerminal(f *os.File) bool {
	if f == nil {
		f = os.Stdin
	}
	return isatty.IsTerminal(f.Fd())
}

func Confirm(prompt string, force bool) (bool, error) {
	if force {
		return true, nil
	}
	if !IsTerminal(os.Stdin) {
		return false, nil
	}
	fmt.Printf("%s [y/N]: ", prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}
