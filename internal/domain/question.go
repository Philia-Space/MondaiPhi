package domain

import (
	"time"

	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-core/domain"
)

// Question is the aggregate root for a JLPT question.
type Question struct {
	domain.AggregateRoot
	ID             examd.QuestionID
	Level          examd.JLPTLevel
	Section        examd.Section
	Prompt         string
	Context        string
	AnswerValue    string // "1","2","3","4" — NEVER exposed to public endpoints
	AnswerNote     string
	PassageID      examd.PassageID
	SourceGroupKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Option is a value object representing one selectable choice.
type Option struct {
	ID          string
	QuestionID  examd.QuestionID
	Value       string // "1","2","3","4"
	Label       string
	SortOrder   int
}

// Passage is a reading/listening passage aggregate that groups questions.
type Passage struct {
	ID             examd.PassageID
	PassageNumber  int
	Title          string
	Content        string
	Level          examd.JLPTLevel
	Section        examd.Section
	CreatedAt      time.Time
}

// Asset represents an audio or image resource.
type Asset struct {
	ID         examd.AssetID
	Type       string // "audio" | "image"
	SourceURL  string
	S3Key      string
	LocalPath  string
	QuestionID examd.QuestionID
	PassageID  examd.PassageID
	CreatedAt  time.Time
}

// PackageTemplate defines a blueprint for generating an exam session.
type PackageTemplate struct {
	ID             string
	Name           string
	Level          examd.JLPTLevel
	SectionCounts  map[examd.Section]int
	TotalQuestions int
	IsDefault      bool
}
