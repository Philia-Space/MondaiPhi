package domain

import (
	"context"

	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-core/domain"
)

// QuestionRepository defines the contract for question bank persistence.
type QuestionRepository interface {
	domain.Repository[Question]
	FindByLevelAndSection(ctx context.Context, level examd.JLPTLevel, section examd.Section, limit int) ([]Question, error)
	FindByPassageID(ctx context.Context, id examd.PassageID) ([]Question, error)
	FindBySourceGroupKey(ctx context.Context, key string) ([]Question, error)
	FindWithOptions(ctx context.Context, id examd.QuestionID) (*Question, []Option, error)
	FindByIDs(ctx context.Context, ids []examd.QuestionID) ([]Question, error)
	FindPassageByID(ctx context.Context, id examd.PassageID) (*Passage, error)
	FindAssetsForQuestions(ctx context.Context, ids []examd.QuestionID) (map[examd.QuestionID][]Asset, error)
	FindAssetsForPassages(ctx context.Context, ids []examd.PassageID) (map[examd.PassageID][]Asset, error)
	FindAssetByID(ctx context.Context, id string) (*Asset, error)
	ListTemplates(ctx context.Context, level examd.JLPTLevel) ([]PackageTemplate, error)
}
