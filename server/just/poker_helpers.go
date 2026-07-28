package just

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetHandScore(cards ...string) (int, error) {
	args := strings.Join(cards, " ")
	Logger.Debugf("getting hand score for: %s", args)
	cmd := exec.Command(Env.PokerEvalCLI, cards...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Run()
	outString := stdout.String()
	errString := stderr.String()

	outString = strings.Trim(outString, "\n")
	Logger.Debugf("evaluating hand: %s [out: %s] [err: %s]", args, outString, errString)

	if errString != "" {
		return 0, fmt.Errorf("encountered error from eval thing: %s", errString)
	}

	outInt, err := strconv.Atoi(outString)
	if err != nil {
		return 0, err
	}

	return outInt, nil
}
