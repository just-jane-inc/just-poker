package just

import (
	"fmt"

	"github.com/google/uuid"
)

type ErrorCode int

const (
	SkillIssue ErrorCode = 67

	Unknown              ErrorCode = 1000
	UserNotFound         ErrorCode = 1001
	GameNotFound         ErrorCode = 1002
	TokenMissing         ErrorCode = 1003
	MalformedRequestBody ErrorCode = 1004

	TurnOrderViolation  ErrorCode = 2000
	InvalidActionType   ErrorCode = 2001
	NotEnoughChips      ErrorCode = 2002
	InvalidBetAmount    ErrorCode = 2003
	InvalidChipExchange ErrorCode = 2004
	ExpectedAnteAction  ErrorCode = 2005

	GameAlreadyStarted       ErrorCode = 2020
	GameIsFull               ErrorCode = 2021
	PlayerAlreadyJoined      ErrorCode = 2022
	GameIsPaused             ErrorCode = 2023
	InvalidGameConfiguration ErrorCode = 2024
	HandAlreadyInProgress    ErrorCode = 2025
	InvalidChipDenomination  ErrorCode = 2026
	InvalidChipCount         ErrorCode = 2027
)

type PokerError struct {
	Message string
	Code    ErrorCode
	Cause   error
}

func NewCriticalInternalError() (*PokerError, string) {
	id := uuid.NewString()

	return NewPokerError(
		fmt.Sprintf("critical error - altert game master - %s", id),
		Unknown,
	), id
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
