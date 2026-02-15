package utils

import "fmt"

// HellBotException represents a general bot error
type HellBotException struct {
	Message string
}

func (e *HellBotException) Error() string {
	return e.Message
}

// NewHellBotException creates a new HellBotException
func NewHellBotException(message string) *HellBotException {
	return &HellBotException{Message: message}
}

// ChangeVCException represents voice chat change error
type ChangeVCException struct {
	Message string
}

func (e *ChangeVCException) Error() string {
	return e.Message
}

// NewChangeVCException creates a new ChangeVCException
func NewChangeVCException(message string) *ChangeVCException {
	return &ChangeVCException{Message: message}
}

// JoinGCException represents group chat join error
type JoinGCException struct {
	Message string
}

func (e *JoinGCException) Error() string {
	return e.Message
}

// NewJoinGCException creates a new JoinGCException
func NewJoinGCException(message string) *JoinGCException {
	return &JoinGCException{Message: message}
}

// JoinVCException represents voice chat join error
type JoinVCException struct {
	Message string
}

func (e *JoinVCException) Error() string {
	return e.Message
}

// NewJoinVCException creates a new JoinVCException
func NewJoinVCException(message string) *JoinVCException {
	return &JoinVCException{Message: message}
}

// UserException represents user-related error
type UserException struct {
	Message string
}

func (e *UserException) Error() string {
	return e.Message
}

// NewUserException creates a new UserException
func NewUserException(message string) *UserException {
	return &UserException{Message: message}
}

// WrapError wraps an error with a custom message
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsHellBotException checks if error is HellBotException
func IsHellBotException(err error) bool {
	_, ok := err.(*HellBotException)
	return ok
}

// IsChangeVCException checks if error is ChangeVCException
func IsChangeVCException(err error) bool {
	_, ok := err.(*ChangeVCException)
	return ok
}

// IsJoinGCException checks if error is JoinGCException
func IsJoinGCException(err error) bool {
	_, ok := err.(*JoinGCException)
	return ok
}

// IsJoinVCException checks if error is JoinVCException
func IsJoinVCException(err error) bool {
	_, ok := err.(*JoinVCException)
	return ok
}

// IsUserException checks if error is UserException
func IsUserException(err error) bool {
	_, ok := err.(*UserException)
	return ok
}
