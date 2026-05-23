package application

import (
	"context"
	"fmt"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
)

// CreateQuestionCommand contains the data needed to create a question.
type CreateQuestionCommand struct {
	Level       examd.JLPTLevel
	Section     examd.Section
	Prompt      string
	Context     string
	AnswerValue string
	AnswerNote  string
	Options     []OptionInput
}

// OptionInput is a DTO for creating an option.
type OptionInput struct {
	Value string
	Label string
}

// CreateQuestionHandler orchestrates question creation.
type CreateQuestionHandler struct {
	repo domain.QuestionRepository
}

func NewCreateQuestionHandler(repo domain.QuestionRepository) *CreateQuestionHandler {
	return &CreateQuestionHandler{repo: repo}
}

func (h *CreateQuestionHandler) Handle(ctx context.Context, cmd CreateQuestionCommand) (*domain.Question, error) {
	// TODO: implement question creation logic
	return nil, fmt.Errorf("not implemented")
}
