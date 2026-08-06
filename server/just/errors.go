package just

import (
	"fmt"
)

type ErrorCode int

const (
	Unknown      ErrorCode = 1000
	UserNotFound ErrorCode = 1001
	GameNotFound ErrorCode = 1002

	TurnOrderViolation  ErrorCode = 2000
	InvalidActionType   ErrorCode = 2001
	NotEnoughChips      ErrorCode = 2002
	InvalidBetAmount    ErrorCode = 2003
	InvalidChipExchange ErrorCode = 2004

	GameAlreadyStarted  ErrorCode = 2020
	GameIsFull          ErrorCode = 2021
	PlayerAlreadyJoined ErrorCode = 2022
	GameIsPaused        ErrorCode = 2023
)

type PokerError struct {
	Message string
	Code    ErrorCode
	Cause   error
}

func NewPokerError(message string, code ErrorCode) *PokerError {
	return &PokerError{
		Message: message,
		Code:    code,
	}
}

func (e *PokerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf(
			"[%d]: %s ( %v )",
			e.Code,
			e.Message,
			e.Cause,
		)
	}

	return fmt.Sprintf(
		"[%d]: %s",
		e.Code,
		e.Message,
	)
}
